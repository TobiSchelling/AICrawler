package server

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/yuin/goldmark"

	"github.com/TobiSchelling/AICrawler/internal/database"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

var md = goldmark.New()

// StorylineView bundles a storyline narrative with its articles and feedback for template rendering.
type StorylineView struct {
	Narrative database.StorylineNarrative
	Articles  []ArticleView
	Feedback  string // "useful", "not_useful", or ""
}

// ArticleView bundles an article with its triage and feedback for template rendering.
type ArticleView struct {
	Article     database.Article
	Triage      *database.ArticleTriage
	Feedback    string // "positive", "negative", or ""
	StorylineID int64
}

// Server is the HTTP server for serving briefings.
type Server struct {
	db    *database.DB
	pages map[string]*template.Template
	mux   *http.ServeMux
}

// New creates a new Server.
func New(db *database.DB) (*Server, error) {
	funcMap := template.FuncMap{
		"markdown":     renderMarkdown,
		"formatPeriod": database.FormatPeriodDisplay,
		"deref": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
	}

	// Parse base template first
	base, err := template.New("base.html").Funcs(funcMap).ParseFS(templateFS, "templates/base.html")
	if err != nil {
		return nil, fmt.Errorf("parsing base template: %w", err)
	}

	// For each page template, clone the base and parse the page into the clone.
	// This gives each page its own {{define "content"}} and {{define "title"}}.
	pageNames := []string{"index.html", "briefing.html", "priorities.html"}
	pages := make(map[string]*template.Template, len(pageNames))
	for _, name := range pageNames {
		clone, err := base.Clone()
		if err != nil {
			return nil, fmt.Errorf("cloning base for %s: %w", name, err)
		}
		_, err = clone.ParseFS(templateFS, "templates/"+name)
		if err != nil {
			return nil, fmt.Errorf("parsing template %s: %w", name, err)
		}
		pages[name] = clone
	}

	s := &Server{db: db, pages: pages, mux: http.NewServeMux()}
	s.routes()
	return s, nil
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	// Static files
	staticSub, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	// Routes
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/briefing/", s.handleBriefing)
	s.mux.HandleFunc("/feedback/storyline/", s.handleStorylineFeedback)
	s.mux.HandleFunc("/feedback/article/", s.handleArticleFeedback)
	s.mux.HandleFunc("/priorities", s.handlePriorities)
	s.mux.HandleFunc("/priorities/add", s.handleAddPriority)
	s.mux.HandleFunc("/priorities/", s.handlePriorityAction)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	briefings, err := s.db.GetAllBriefings()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.render(w, "index.html", map[string]any{
		"Briefings": briefings,
	})
}

func (s *Server) handleBriefing(w http.ResponseWriter, r *http.Request) {
	periodID := strings.TrimPrefix(r.URL.Path, "/briefing/")
	if periodID == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	briefing, _ := s.db.GetBriefing(periodID)

	// Build structured storyline views
	var storylines []StorylineView
	narratives, _ := s.db.GetNarrativesForPeriod(periodID)
	sfMap, _ := s.db.GetStorylineFeedbackMap(periodID)

	// Collect all article IDs for batch feedback lookup
	var allArticleIDs []int64
	type narrativeArticles struct {
		articles []database.Article
	}
	naArticles := make([]narrativeArticles, len(narratives))

	for i, n := range narratives {
		articles, _ := s.db.GetStorylineArticles(n.StorylineID)
		naArticles[i] = narrativeArticles{articles: articles}
		for _, a := range articles {
			allArticleIDs = append(allArticleIDs, a.ID)
		}
	}

	afMap, _ := s.db.GetArticleFeedbackMap(allArticleIDs)

	for i, n := range narratives {
		sv := StorylineView{
			Narrative: n,
			Feedback:  sfMap[n.StorylineID],
		}
		for _, a := range naArticles[i].articles {
			triage, _ := s.db.GetTriage(a.ID)
			sv.Articles = append(sv.Articles, ArticleView{
				Article:     a,
				Triage:      triage,
				Feedback:    afMap[a.ID],
				StorylineID: n.StorylineID,
			})
		}
		storylines = append(storylines, sv)
	}

	// Fallback: when no storylines exist, load articles directly for feedback
	var articles []ArticleView
	if len(storylines) == 0 && briefing != nil {
		allArticles, err := s.db.GetArticlesForPeriod(periodID)
		if err != nil {
			log.Printf("error fetching articles for period %s: %v", periodID, err)
		}
		var articleIDs []int64
		for _, a := range allArticles {
			articleIDs = append(articleIDs, a.ID)
		}
		afMap, _ := s.db.GetArticleFeedbackMap(articleIDs)
		for _, a := range allArticles {
			triage, _ := s.db.GetTriage(a.ID)
			if triage == nil || triage.Verdict != "relevant" {
				continue
			}
			articles = append(articles, ArticleView{
				Article:  a,
				Triage:   triage,
				Feedback: afMap[a.ID],
			})
		}
	}

	s.render(w, "briefing.html", map[string]any{
		"Briefing":   briefing,
		"PeriodID":   periodID,
		"Storylines": storylines,
		"Articles":   articles,
	})
}

func (s *Server) handleStorylineFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Parse /feedback/storyline/{id}/{rating}
	path := strings.TrimPrefix(r.URL.Path, "/feedback/storyline/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	rating := parts[1]

	periodID := r.FormValue("period_id")
	if periodID == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Toggle: if current == submitted, delete; otherwise upsert
	current, _ := s.db.GetStorylineFeedback(id)
	if current != nil && current.Rating == rating {
		s.db.DeleteStorylineFeedback(id)
	} else {
		s.db.UpsertStorylineFeedback(id, periodID, rating)
	}

	http.Redirect(w, r, fmt.Sprintf("/briefing/%s#storyline-%d", periodID, id), http.StatusFound)
}

func (s *Server) handleArticleFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Parse /feedback/article/{id}/{rating}
	path := strings.TrimPrefix(r.URL.Path, "/feedback/article/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	rating := parts[1]

	periodID := r.FormValue("period_id")
	storylineID := r.FormValue("storyline_id")
	if periodID == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Toggle: if current == submitted, delete; otherwise upsert
	current, _ := s.db.GetArticleFeedback(id)
	if current != nil && current.Rating == rating {
		s.db.DeleteArticleFeedback(id)
	} else {
		s.db.UpsertArticleFeedback(id, rating)
	}

	anchor := ""
	if storylineID != "" {
		anchor = "#storyline-" + storylineID
	}
	http.Redirect(w, r, fmt.Sprintf("/briefing/%s%s", periodID, anchor), http.StatusFound)
}

func (s *Server) handlePriorities(w http.ResponseWriter, r *http.Request) {
	priorities, _ := s.db.GetAllPriorities()
	s.render(w, "priorities.html", map[string]any{
		"Priorities": priorities,
	})
}

func (s *Server) handleAddPriority(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/priorities", http.StatusFound)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))

	if title != "" {
		s.db.InsertPriority(title, description, nil)
	}

	http.Redirect(w, r, "/priorities", http.StatusFound)
}

func (s *Server) handlePriorityAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/priorities", http.StatusFound)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/priorities/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Redirect(w, r, "/priorities", http.StatusFound)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Redirect(w, r, "/priorities", http.StatusFound)
		return
	}

	switch parts[1] {
	case "toggle":
		s.db.TogglePriority(id)
	case "delete":
		s.db.DeletePriority(id)
	case "edit":
		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))
		if title != "" {
			s.db.UpdatePriority(id, &title, &description, nil)
		}
	}

	http.Redirect(w, r, "/priorities", http.StatusFound)
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	tmpl, ok := s.pages[name]
	if !ok {
		log.Printf("Template %s not found", name)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
	}
}

func renderMarkdown(text string) template.HTML {
	var buf bytes.Buffer
	if err := md.Convert([]byte(text), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(text))
	}
	return template.HTML(buf.String()) //nolint: gosec
}

// Serve starts the HTTP server on the given port.
func Serve(db *database.DB, port int) error {
	srv, err := New(db)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		if isAddrInUse(err) {
			return fmt.Errorf("port %d already in use%s", port, identifyPortHolder(port))
		}
		return err
	}

	log.Printf("Server listening on http://%s", addr)
	return http.Serve(ln, srv.Handler())
}

func isAddrInUse(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var sysErr *os.SyscallError
		if errors.As(opErr.Err, &sysErr) {
			return errors.Is(sysErr.Err, syscall.EADDRINUSE)
		}
	}
	return false
}

// identifyPortHolder uses lsof to find which process holds the port.
func identifyPortHolder(port int) string {
	out, err := exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%d", port)).Output()
	if err != nil || len(out) == 0 {
		return ""
	}

	pid := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	cmd, err := exec.Command("ps", "-p", pid, "-o", "command=").Output()
	if err != nil || len(cmd) == 0 {
		return fmt.Sprintf(" (pid %s)", pid)
	}

	return fmt.Sprintf(" (pid %s: %s)", pid, strings.TrimSpace(string(cmd)))
}

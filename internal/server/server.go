package server

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"

	"github.com/TobiSchelling/AICrawler/internal/database"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

var md = goldmark.New()

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

	s.render(w, "briefing.html", map[string]any{
		"Briefing": briefing,
		"PeriodID": periodID,
	})
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
	log.Printf("Server listening on http://%s", addr)
	return http.ListenAndServe(addr, srv.Handler())
}

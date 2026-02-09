package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TobiSchelling/AICrawler/internal/collect"
	"github.com/TobiSchelling/AICrawler/internal/config"
	"github.com/TobiSchelling/AICrawler/internal/database"
	"github.com/TobiSchelling/AICrawler/internal/pipeline"
	"github.com/TobiSchelling/AICrawler/internal/server"
	"github.com/spf13/cobra"
)

var version = "dev"

var (
	verbose    bool
	configPath string
	cfg        *config.Config
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "aicrawler",
	Short:   "Daily AI news briefings",
	Long:    "AICrawler collects, triages, clusters, and narrates AI developments into daily briefings.",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
		} else {
			log.SetFlags(log.LstdFlags)
		}

		// Skip config loading for init and version
		if cmd.Name() == "init" || cmd.Name() == "version" {
			return nil
		}

		path, err := config.ResolveConfigPath(configPath)
		if err != nil {
			return err
		}
		cfg, err = config.Load(path)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to config file")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(prioritiesCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aicrawler", version)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration in ~/.config/aicrawler/",
	RunE: func(cmd *cobra.Command, args []string) error {
		target := filepath.Join(config.ConfigDir(), "config.yaml")
		if _, err := os.Stat(target); err == nil {
			fmt.Printf("Config already exists: %s\n", target)
			return nil
		}

		if err := os.MkdirAll(config.ConfigDir(), 0o755); err != nil {
			return fmt.Errorf("creating config directory: %w", err)
		}

		if err := os.WriteFile(target, config.DefaultConfigYAML, 0o644); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}

		fmt.Printf("Created config: %s\n", target)
		fmt.Println("Edit it to configure feeds, API keys, and LLM provider.")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database and system status",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB()
		if err != nil {
			return err
		}
		defer db.Close()

		stats, err := db.GetStats()
		if err != nil {
			return fmt.Errorf("getting stats: %w", err)
		}

		today := database.GetToday()
		fmt.Printf("Today: %s\n\n", today)
		fmt.Println("Articles:")
		fmt.Printf("  Total collected: %d\n", stats.TotalArticles)
		fmt.Printf("  Triaged: %d\n", stats.TriagedArticles)
		fmt.Printf("  Relevant: %d\n", stats.RelevantArticles)
		fmt.Println("\nOutput:")
		fmt.Printf("  Storylines: %d\n", stats.Storylines)
		fmt.Printf("  Briefings: %d\n", stats.Briefings)
		fmt.Printf("  Days with data: %d\n", stats.PeriodsWithArticles)
		fmt.Println("\nResearch Priorities:")
		fmt.Printf("  Total: %d\n", stats.TotalPriorities)
		fmt.Printf("  Active: %d\n", stats.ActivePriorities)
		return nil
	},
}

// --- collect command ---

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect articles from configured sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB()
		if err != nil {
			return err
		}
		defer db.Close()

		periodID := database.GetToday()
		fmt.Println("Collecting articles from sources...")

		collector := collect.NewCollector(cfg, db, 1)
		result := collector.Collect(periodID)

		fmt.Println("\nCollection complete:")
		fmt.Printf("  Total found: %d\n", result.TotalFound)
		fmt.Printf("  New articles: %d\n", result.NewArticles)
		fmt.Printf("  Duplicates skipped: %d\n", result.Duplicates)

		if len(result.Sources) > 0 {
			fmt.Println("\nArticles by source:")
			// Sort sources by count descending
			type kv struct {
				key string
				val int
			}
			var sorted []kv
			for k, v := range result.Sources {
				sorted = append(sorted, kv{k, v})
			}
			sort.Slice(sorted, func(i, j int) bool { return sorted[i].val > sorted[j].val })
			for _, s := range sorted {
				fmt.Printf("  %s: %d\n", s.key, s.val)
			}
		}
		return nil
	},
}

// --- run command ---

var (
	dryRun   bool
	daysBack int
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the full pipeline: collect -> fetch -> triage -> cluster -> synthesize -> compose",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB()
		if err != nil {
			return err
		}
		defer db.Close()

		today := database.GetToday()
		periodID, effectiveDaysBack, err := resolvePeriod(db, today, daysBack)
		if err != nil {
			return err
		}

		pipe := pipeline.New(cfg, db)
		ctx := context.Background()

		var result *pipeline.Result
		if dryRun {
			result = pipe.DryRun(periodID)
		} else {
			result = pipe.Run(ctx, periodID, effectiveDaysBack)
		}

		for i, step := range result.Steps {
			fmt.Printf("\nStep %d/6: %s\n", i+1, step.Name)
			if step.Err != nil {
				fmt.Printf("  Error: %v\n", step.Err)
			} else {
				fmt.Printf("  %s\n", step.Summary)
			}
		}

		if !dryRun {
			fmt.Println("\nPipeline complete! Run 'aicrawler serve' to view the briefing.")
		}
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
	runCmd.Flags().IntVar(&daysBack, "days-back", 0, "Override lookback window (days)")
}

// resolvePeriod determines the period ID and effective days back based on
// explicit --days-back, catch-up detection, or daily run.
func resolvePeriod(db *database.DB, today string, explicitDaysBack int) (periodID string, effectiveDaysBack int, err error) {
	if explicitDaysBack > 0 {
		if explicitDaysBack == 1 {
			periodID = today
		} else {
			todayDate, _ := time.Parse("2006-01-02", today)
			start := todayDate.AddDate(0, 0, -(explicitDaysBack - 1)).Format("2006-01-02")
			periodID = database.MakePeriodID(start, today)
		}
		fmt.Printf("Collecting %d day(s) of articles (%s).\n", explicitDaysBack, periodID)
		return periodID, explicitDaysBack, nil
	}

	lastRun, _ := db.GetLastRunDate()
	if lastRun == "" {
		fmt.Println("First run detected — collecting today's articles.")
		return today, 1, nil
	}

	lastDate, _ := time.Parse("2006-01-02", lastRun)
	todayDate, _ := time.Parse("2006-01-02", today)
	missedDays := int(todayDate.Sub(lastDate).Hours() / 24)

	if missedDays <= 0 {
		fmt.Printf("Already ran today (%s). Re-running pipeline.\n", today)
		return today, 1, nil
	}

	if missedDays == 1 {
		fmt.Printf("Daily run for %s.\n", today)
		return today, 1, nil
	}

	// Catch-up: missed multiple days
	startDate := lastDate.AddDate(0, 0, 1).Format("2006-01-02")
	periodID = database.MakePeriodID(startDate, today)

	if missedDays > 5 {
		fmt.Printf("Last run was %d days ago (%s).\n", missedDays, lastRun)
		fmt.Printf("Catch up %d days (%s)? This will use more API calls [y/N]: ", missedDays, periodID)

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			return "", 0, fmt.Errorf("aborted")
		}
	} else {
		fmt.Printf("Catching up %d days (%s).\n", missedDays, periodID)
	}

	return periodID, missedDays, nil
}

// --- serve command ---

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the local web server",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB()
		if err != nil {
			return err
		}
		defer db.Close()

		fmt.Printf("Starting server at http://localhost:%d\n", servePort)
		fmt.Println("Press Ctrl+C to stop")
		return server.Serve(db, servePort)
	},
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8000, "Port to run server on")
}

// --- priorities command ---

var prioritiesCmd = &cobra.Command{
	Use:   "priorities",
	Short: "Manage research priorities",
}

var prioritiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all research priorities",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB()
		if err != nil {
			return err
		}
		defer db.Close()

		items, err := db.GetAllPriorities()
		if err != nil {
			return err
		}

		if len(items) == 0 {
			fmt.Println("No priorities defined. Add one with: aicrawler priorities add")
			return nil
		}

		fmt.Println("Research Priorities:")
		fmt.Println()
		for _, p := range items {
			icon := " "
			if p.IsActive {
				icon = "*"
			}
			fmt.Printf("  [%d] %s %s\n", p.ID, icon, p.Title)
			if p.Description != nil && *p.Description != "" {
				desc := *p.Description
				if len(desc) > 60 {
					desc = desc[:60] + "..."
				}
				fmt.Printf("        %s\n", desc)
			}
		}
		return nil
	},
}

var prioritiesAddCmd = &cobra.Command{
	Use:   "add [title] [description]",
	Short: "Add a new research priority",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB()
		if err != nil {
			return err
		}
		defer db.Close()

		title := args[0]
		description := ""
		if len(args) > 1 {
			description = args[1]
		}

		id, err := db.InsertPriority(title, description, nil)
		if err != nil {
			return err
		}
		fmt.Printf("Added priority [%d]: %s\n", id, title)
		return nil
	},
}

var prioritiesRemoveCmd = &cobra.Command{
	Use:   "remove [id]",
	Short: "Remove a research priority",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB()
		if err != nil {
			return err
		}
		defer db.Close()

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid priority ID: %s", args[0])
		}

		priority, err := db.GetPriority(id)
		if err != nil {
			return err
		}
		if priority == nil {
			return fmt.Errorf("priority %d not found", id)
		}

		if err := db.DeletePriority(id); err != nil {
			return err
		}
		fmt.Printf("Removed priority [%d]: %s\n", id, priority.Title)
		return nil
	},
}

var prioritiesToggleCmd = &cobra.Command{
	Use:   "toggle [id]",
	Short: "Toggle a priority's active state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB()
		if err != nil {
			return err
		}
		defer db.Close()

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid priority ID: %s", args[0])
		}

		priority, err := db.GetPriority(id)
		if err != nil {
			return err
		}
		if priority == nil {
			return fmt.Errorf("priority %d not found", id)
		}

		if err := db.TogglePriority(id); err != nil {
			return err
		}
		newState := "disabled"
		if !priority.IsActive {
			newState = "enabled"
		}
		fmt.Printf("Priority [%d] %s: %s\n", id, priority.Title, newState)
		return nil
	},
}

func init() {
	prioritiesCmd.AddCommand(prioritiesListCmd)
	prioritiesCmd.AddCommand(prioritiesAddCmd)
	prioritiesCmd.AddCommand(prioritiesRemoveCmd)
	prioritiesCmd.AddCommand(prioritiesToggleCmd)
}

func openDB() (*database.DB, error) {
	dataDir := cfg.GetDataDir()
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}
	dbPath := filepath.Join(dataDir, "aicrawler.db")
	return database.Open(dbPath)
}

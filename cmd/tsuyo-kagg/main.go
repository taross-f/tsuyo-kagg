package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/taross-f/tsuyo-kagg/pkg/config"
	"github.com/taross-f/tsuyo-kagg/pkg/kaggle"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	outputPath := flag.String("output", "", "Path to output file (overrides config)")
	generateConfig := flag.Bool("generate-config", false, "Generate default configuration file")
	maxPages := flag.Int("max-pages", 0, "Maximum number of pages to scrape (overrides config)")
	useAPI := flag.Bool("use-api", true, "Use Kaggle API instead of web scraping")
	flag.Parse()

	// Generate default configuration if requested
	if *generateConfig {
		cfg := config.DefaultConfig()
		err := config.SaveConfig(cfg, *configPath)
		if err != nil {
			log.Fatalf("Failed to generate configuration file: %v", err)
		}
		log.Printf("Generated default configuration at %s", *configPath)
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Printf("Warning: Failed to load configuration file: %v", err)
		log.Printf("Using default configuration")
		cfg = config.DefaultConfig()
	}

	// Override configuration with command line flags
	if *outputPath != "" {
		cfg.OutputFile = *outputPath
	}
	if *maxPages > 0 {
		cfg.MaxPages = *maxPages
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(cfg.OutputFile)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}

	var users []kaggle.User

	if *useAPI {
		// Use Kaggle API
		log.Printf("Using Kaggle API to fetch users")
		apiClient := kaggle.NewAPIClient(cfg)
		users, err = apiClient.GetUsers()
		if err != nil {
			log.Fatalf("Failed to get users via API: %v", err)
		}
	} else {
		// Use web scraping
		log.Printf("Using web scraping to fetch users")
		scraper := kaggle.NewScraper(cfg)
		users, err = scraper.GetRankings()
		if err != nil {
			log.Fatalf("Failed to get rankings: %v", err)
		}
	}

	// Export to CSV
	log.Printf("Found %d users from %s", len(users), cfg.TargetCountry)

	var csvData string
	if *useAPI {
		apiClient := kaggle.NewAPIClient(cfg)
		csvData = apiClient.ExportToCSV(users)
	} else {
		scraper := kaggle.NewScraper(cfg)
		csvData = scraper.ExportToCSV(users)
	}

	// Write to file
	if err := os.WriteFile(cfg.OutputFile, []byte(csvData), 0644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	log.Printf("Successfully wrote data for %d users to %s", len(users), cfg.OutputFile)
	fmt.Printf("Successfully wrote data for %d users to %s\n", len(users), cfg.OutputFile)
}

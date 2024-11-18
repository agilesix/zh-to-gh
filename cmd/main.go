package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"log/slog"

	"github.com/widal001/zhtogh/pkg/github" // Replace with the actual import path for GitHubClient
)

func main() {

	// Set logs to include short filename
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Define and parse command-line flags
	configFile := flag.String("config", "", "Path to the JSON config file")
	flag.Parse()

	if *configFile == "" {
		fmt.Println("Usage: program -config <path to config JSON>")
		os.Exit(1)
	}

	// Open the JSON file
	file, err := os.Open(*configFile)
	if err != nil {
		slog.Error("Error opening config file", "path", *configFile, "error", err)
		os.Exit(1)
	}
	defer file.Close()

	// Parse the JSON file
	var issuesMap map[string][]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&issuesMap); err != nil {
		slog.Error("Error parsing config file", "path", *configFile, "error", err)
		os.Exit(1)
	}

	// Initialize the GitHub client
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		slog.Error("You must set the GITHUB_TOKEN environment variable")
		os.Exit(1)
	}
	gh := github.NewGitHubClient(token)

	// Create a wait group and channel to synchronize goroutines
	var wg sync.WaitGroup
	resultChan := make(chan github.AddSubIssueResults, len(issuesMap))

	// Process each parent and its children in a goroutine
	for parent, children := range issuesMap {
		wg.Add(1)

		// Launch a goroutine for each parent-child migration
		go func(parent string, children []string) {
			defer wg.Done()

			// Log start of parent issue
			fmt.Println("- Adding sub-issues for:", parent)

			// Perform the migration and collect the result
			resultChan <- gh.AddSubIssuesByUrl(parent, children)

		}(parent, children) // Pass the parent and children to the goroutine
	}

	// Close the channels when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect the results
	for result := range resultChan {
		// Log the parent
		fmt.Printf("\n\n### Results for parent issue: %s\n\n", result.Parent)

		// Log the successes
		fmt.Printf("Added:\t%d\n", len(result.Added))
		for _, url := range result.Added {
			fmt.Println("- ", url)
		}

		// Log the failures
		fmt.Printf("Failed:\t%d\n", len(result.Errors))
		for url, err := range result.Errors {
			fmt.Printf("- %s: %s\n", url, err)
		}

	}
}

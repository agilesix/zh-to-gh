package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

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

	// Process each parent and its children
	for parent, children := range issuesMap {
		slog.Info("Processing parent issue", "parent", parent)
		gh.AddSubIssuesByUrl(parent, children)
	}

	slog.Info("Finished processing all issues")
}

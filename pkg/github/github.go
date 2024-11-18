package github

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/widal001/zhtogh/pkg/graphql"
)

const API_URL = "https://api.github.com/graphql"

// =========================================================
// GitHub client
// =========================================================

type GitHubClient struct {
	graphql.Client
}

func NewGitHubClient(token string) GitHubClient {
	client := graphql.
		NewClient(API_URL, token).
		WithDefaultHeader("GraphQL-Features", "sub_issues").
		WithDefaultHeader("GraphQL-Features", "issue_types")
	return GitHubClient{*client}
}

// =========================================================
// Add sub-issues by their URL
// =========================================================

type AddSubIssueResults struct {
	Parent string
	Added  []string
	Errors map[string]error
}

// Add multiple sub-issues to parent issue using their URLs
func (c GitHubClient) AddSubIssuesByUrl(parent string, children []string) AddSubIssueResults {
	// Instantiate the results
	results := AddSubIssueResults{
		Parent: parent,
		Errors: make(map[string]error),
	}

	// Get the node ID for parent issue
	parentId, err := c.GetIssueIdByURL(parent)
	if err != nil {
		slog.Error("Error getting node ID for parent issue", "parent", parent, "error", err)
		return results // exit early
	}

	// Get the node IDs for each child concurrently
	childIds := c.BatchGetIssueIds(children)

	// Add each child issue to the parent
	for _, child := range childIds {

		// Add the child issue to the parent
		err := c.AddSubIssueById(parentId, child.IssueID)
		// Capture failure
		if err != nil {
			results.Errors[child.URL] = err
			continue
		}
		// Capture success
		results.Added = append(results.Added, child.URL)
	}

	//  Return the results
	return results
}

// =========================================================
// Get the ID for multiple issues by their URLs
// =========================================================

type IssueMapping struct {
	URL     string
	IssueID IssueId
}

func (c *GitHubClient) BatchGetIssueIds(urls []string) []IssueMapping {

	// Create wait group and channels to synchronize goroutines
	var wg sync.WaitGroup
	issueIdChan := make(chan IssueMapping, len(urls))
	errorsChan := make(chan error, len(urls))

	// Launch a goroutine for each child
	for _, url := range urls {
		wg.Add(1)
		go func(child string) {
			defer wg.Done()
			// Get the node ID for the current child
			id, err := c.GetIssueIdByURL(url)
			if err != nil {
				errorsChan <- fmt.Errorf("failed to get ID for %s: %w", url, err)
				return
			}
			issueIdChan <- IssueMapping{URL: url, IssueID: id}
		}(url) // Pass the `child` variable to the goroutine
	}

	// Close channels when all goroutines have finished
	go func() {
		wg.Wait()
		close(errorsChan)
		close(issueIdChan)
	}()

	// Collect errors and issue mappings
	var issues []IssueMapping
	for err := range errorsChan {
		slog.Debug("Failed to get ID", "error", err)
	}
	for issue := range issueIdChan {
		issues = append(issues, issue)
	}

	// Return the issue mappings
	return issues
}

// =========================================================
// Get an issue's ID by its URL
// =========================================================

type IssueId string

// Get the GraphQL node ID of a GitHub issue by its URL.
//
// In order to add a sub-issue to another issue, you need the node ID for both issues.
// Node IDs for issues are an 18-character string prefixed with `I_`, e.g. `I_kwDOMzlMsM6XpAXQ`
func (c *GitHubClient) GetIssueIdByURL(url string) (IssueId, error) {
	// Define the query string with `$url` as an input variable
	queryStr := `
query ($url: URI!) {
  resource(url: $url) {
    ... on Issue { 
	  id
	}
  }
}`
	// Map the input variable to its value
	query := graphql.Query{
		QueryStr: queryStr,
		Vars:     map[string]interface{}{"url": url},
	}
	// Declare a struct that matches the expected response JSON
	var response struct {
		Data struct {
			Resource struct {
				Id string
			}
		}
	}
	// Post the query and parse the response
	err := query.Post(c.Client, &response)
	if err != nil {
		return "", fmt.Errorf("failed to get issue ID by its URL: %v", err)
	}
	return IssueId(response.Data.Resource.Id), nil
}

// =========================================================
// Add a sub-issue to another issue
// =========================================================

// Add a sub-issue to another issue using their GraphQL node IDs
//
// Node IDs for issues are an 18-character string prefixed with `I_`, e.g. `I_kwDOMzlMsM6XpAXQ`
func (c *GitHubClient) AddSubIssueById(parentId, childId IssueId) error {
	// Define the mutation string with `$parentId` and `$childId` as input variables
	queryStr := `
mutation($parentId: ID!, $childId: ID!) {
  addSubIssue(input: { issueId: $parentId, subIssueId: $childId, replaceParent: true }) {
    issue {
      url
    }
    subIssue {
      url
    }
  }
}`
	// Map the input variables to their values
	query := graphql.Query{
		QueryStr: queryStr,
		Vars: map[string]interface{}{
			"parentId": parentId,
			"childId":  childId,
		},
	}
	// Declare a struct that matches the expected response JSON
	var response struct {
		Data struct {
			AddSubIssue struct {
				Issue struct {
					Url string
				}
				SubIssue struct {
					Url string
				}
			}
		}
	}
	// Post the query and parse the response
	err := query.Post(c.Client, &response)
	if err != nil {
		return fmt.Errorf("error adding sub-issue: %v", err)
	}
	return nil
}

package github

import (
	"fmt"
	"log"

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

// Add multiple sub-issues to parent issue using their URLs
func (c GitHubClient) AddSubIssuesByUrl(parent string, children []string) {
	// Get the node ID for parent issue
	parentId, err := c.GetIssueIdByURL(parent)
	if err != nil {
		log.Printf("Error getting node ID for parent issue %s: %s", parent, err)
		return // exit early
	}

	// Add each child issue to the parent
	for _, child := range children {
		// Get the node ID for the current child
		childId, err := c.GetIssueIdByURL(child)
		if err != nil {
			log.Printf("Error getting node ID for parent issue %s: %s", parent, err)
		}
		// Add the child issue to the parent
		err = c.AddSubIssueById(parentId, childId)
		if err != nil {
			log.Printf("Error adding sub-issue %s to issue %s: %s", parent, child, err)
		}

	}

}

// =========================================================
// Get an issue's ID by its URL
// =========================================================

type IssueId string

// Get the GraphQL node ID of a GitHub issue by its URL.
//
// In order to add a sub-issue to another issue, you need the node ID for both issues.
// Node IDs for issues are an 18-character string prefixed with `I_`, e.g. `I_kwDOMzlMsM6XpAXQ`
func (c GitHubClient) GetIssueIdByURL(url string) (IssueId, error) {
	// Define the query string with `$url` as an input variable
	queryStr := `
query ($url: URI!) {
  resource(url: $issueUrl) {
    ... on Issue { id }
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

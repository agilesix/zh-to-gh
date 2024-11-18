package graphql

import (
	"encoding/json"
	"fmt"
)

type Query struct {
	QueryStr string
	Vars     map[string]interface{}
}

// Post a query using the given GraphQL API client and parse the response
func (q *Query) Post(client Client, response interface{}) error {
	// Create the requestBody
	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     q.QueryStr,
		"variables": q.Vars,
	})
	if err != nil {
		return fmt.Errorf("error formatting the request body: %v", err)
	}
	// Make the API request
	responseBody, err := client.Post(requestBody)
	if err != nil {
		return err
	}
	// Check for errors
	errors, exist := ResponseHasErrors(responseBody)
	if exist {
		return fmt.Errorf("the following errors were returned: %v", errors)
	}
	// Unmarshal the JSON response into the struct
	err = json.Unmarshal([]byte(responseBody), &response)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err
	}
	return nil
}

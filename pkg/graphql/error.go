package graphql

import "encoding/json"

type ResponseError struct {
	Errors *[]Error
}

type Error struct {
	Message string
}

// Check response for errors
func ResponseHasErrors(data []byte) ([]string, bool) {
	// Instantiate an empty response and error list
	var errors []string
	var response ResponseError
	// Try to parse a list of errors from the raw response text
	err := json.Unmarshal(data, &response)
	if err != nil {
		errors = append(errors, err.Error())
		return errors, true
	}
	// If the response doesn't contain an errors key return false
	if response.Errors == nil {
		return errors, false
	}
	// Otherwise return true along with the message from each error
	for _, err := range *response.Errors {
		errors = append(errors, err.Message)
	}
	return errors, true
}

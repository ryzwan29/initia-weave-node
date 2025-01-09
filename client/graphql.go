package client

import (
	"encoding/json"
	"fmt"
)

// GraphQLClient defines the logic for interacting with GraphQL APIs.
type GraphQLClient struct {
	httpClient *HTTPClient
	baseURL    string
}

// NewGraphQLClient creates and returns a new GraphQLClient instance.
func NewGraphQLClient(baseURL string, httpClient *HTTPClient) *GraphQLClient {
	return &GraphQLClient{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// Query sends a GraphQL query to the API and decodes the response into the result.
func (c *GraphQLClient) Query(query string, variables map[string]interface{}, result interface{}) error {
	payload := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal GraphQL payload: %w", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	_, err = c.httpClient.Post(c.baseURL, "", headers, payloadBytes, result)
	if err != nil {
		return fmt.Errorf("failed to perform GraphQL query: %w", err)
	}

	return nil
}

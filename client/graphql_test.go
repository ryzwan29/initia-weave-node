package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGraphQLClient_Query_Success tests a successful GraphQL query.
func TestGraphQLClient_Query_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, http.MethodPost, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data":{"message":"GraphQL query successful"}}`))
		assert.NoError(t, err)
	}))
	defer mockServer.Close()

	httpClient := NewHTTPClient()
	graphqlClient := NewGraphQLClient(mockServer.URL, httpClient)

	query := `
		query GetMessage {
			message
		}
	`
	variables := map[string]interface{}{}
	var result struct {
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	}

	err := graphqlClient.Query(query, variables, &result)

	assert.NoError(t, err)
	assert.Equal(t, "GraphQL query successful", result.Data.Message)
}

// TestGraphQLClient_Query_Error tests a GraphQL query with an error response.
func TestGraphQLClient_Query_Error(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"errors":[{"message":"Invalid query"}]}`))
		assert.NoError(t, err)
	}))
	defer mockServer.Close()

	httpClient := NewHTTPClient()
	graphqlClient := NewGraphQLClient(mockServer.URL, httpClient)

	query := `
		query InvalidQuery {
			invalidField
		}
	`
	variables := map[string]interface{}{}
	var result struct {
		Data interface{} `json:"data"`
	}

	err := graphqlClient.Query(query, variables, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to perform GraphQL query")
}

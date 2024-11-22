package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPClient_Get_Success(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message":"success"}`))
		if err != nil {
			t.Fatalf("Error writing response: %v", err)
		}
	}))
	defer mockServer.Close()

	client := NewHTTPClient()

	// Test successful GET request
	var result map[string]string
	_, err := client.Get(mockServer.URL, "", nil, &result)

	assert.NoError(t, err)
	assert.Equal(t, "success", result["message"])
}

// TestHTTPClient_Get_Failure tests a GET request that returns an error
func TestHTTPClient_Get_Failure(t *testing.T) {
	// Create a mock HTTP server that returns a bad status code
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	client := NewHTTPClient()

	// Test GET failure with status code 500
	_, err := client.Get(mockServer.URL, "", nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code")
}

// TestHTTPClient_DownloadFile_Success tests a successful file download
func TestHTTPClient_DownloadFile_Success(t *testing.T) {
	// Mock file download behavior
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("test content"))
		if err != nil {
			t.Fatalf("Error writing response: %v", err)
		}
	}))
	defer mockServer.Close()

	// Create a temporary file to save the content
	destFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(destFile.Name())

	client := NewHTTPClient()

	// Test DownloadFile method
	progress := int64(0)
	totalSize := int64(0)
	err = client.DownloadFile(mockServer.URL, destFile.Name(), &progress, &totalSize)

	assert.NoError(t, err)
	assert.Equal(t, int64(12), totalSize) // The content length of "test content"
	// Verify file content
	content, err := os.ReadFile(destFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}

// TestHTTPClient_DownloadFile_Failure tests a failed file download (non-200 status)
func TestHTTPClient_DownloadFile_Failure(t *testing.T) {
	// Create a mock HTTP server that returns a bad status code
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	client := NewHTTPClient()

	// Test DownloadFile failure with status code 500
	destFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(destFile.Name())

	err = client.DownloadFile(mockServer.URL, destFile.Name(), nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download")
}

// TestHTTPClient_DownloadFile_InvalidURL tests invalid URL handling for file download
func TestHTTPClient_DownloadFile_InvalidURL(t *testing.T) {
	client := NewHTTPClient()

	// Test with invalid URL
	err := client.DownloadFile("http://invalid-url", "/path/to/file", nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}

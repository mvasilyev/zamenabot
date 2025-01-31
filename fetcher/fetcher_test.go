package fetcher

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type MockHTTPClient struct {
	// Add a field for mocking the response or error
	Resp *http.Response
	Err  error
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	return m.Resp, m.Err
}

func TestFetcher_FetchSchedule_Success(t *testing.T) {
	// Define a mock HTTP response
	mockData := "Date,Event\n2025-01-01,New Year\n2025-02-14,Valentine's Day"
	mockResponse := &http.Response{
		StatusCode: 200,
		Body:       http.NoBody, // We'll use mock data instead of actual body
	}

	// Mock the response body with our data
	mockRespBody := ioutil.NopCloser(strings.NewReader(mockData))
	mockResponse.Body = mockRespBody

	mockClient := &MockHTTPClient{
		Resp: mockResponse,
	}

	fetcher := Fetcher{
		SheetID:    "mock-sheet-id",
		HTTPClient: mockClient,
	}

	// Test the FetchSchedule method
	data, err := fetcher.FetchSchedule()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := [][]string{
		{"Date", "Event"},
		{"2025-01-01", "New Year"},
		{"2025-02-14", "Valentine's Day"},
	}

	if !equal(data, expected) {
		t.Fatalf("Expected data %v, got %v", expected, data)
	}
}

func TestFetcher_FetchSchedule_Error(t *testing.T) {
	// Simulate an error in the HTTP client
	mockClient := &MockHTTPClient{
		Err: errors.New("network error"),
	}

	fetcher := Fetcher{
		SheetID:    "mock-sheet-id",
		HTTPClient: mockClient,
	}

	// Test the FetchSchedule method
	_, err := fetcher.FetchSchedule()
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func equal(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
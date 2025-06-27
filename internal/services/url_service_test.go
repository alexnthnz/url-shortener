package services

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestBase62Encoding(t *testing.T) {
	service := &URLService{
		logger: logrus.New(),
	}

	testCases := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{61, "Z"},
		{62, "10"},
		{123, "1Z"},
		{12345, "3d7"},
		{999999, "4c91"}, // Additional test case for larger numbers
	}

	for _, tc := range testCases {
		result := service.encodeBase62(tc.input)
		if result != tc.expected {
			t.Errorf("encodeBase62(%d) = %s; expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestValidateURL(t *testing.T) {
	service := &URLService{
		logger: logrus.New(),
	}

	validURLs := []string{
		"https://example.com",
		"http://example.com",
		"https://subdomain.example.com/path",
		"https://example.com:8080/path?query=value",
		"https://www.example.com/path/to/resource",
		"http://192.168.1.1:3000",
	}

	for _, url := range validURLs {
		if err := service.validateURL(url); err != nil {
			t.Errorf("validateURL(%s) should be valid, got error: %v", url, err)
		}
	}

	invalidURLs := []string{
		"javascript:alert('xss')",
		"file:///etc/passwd",
		"ftp://example.com",
		"data:text/html,<script>alert('xss')</script>",
		"not-a-url",
		"",
		"https://",
		"http://",
	}

	for _, url := range invalidURLs {
		if err := service.validateURL(url); err == nil {
			t.Errorf("validateURL(%s) should be invalid, but passed", url)
		}
	}
}

func TestValidateCustomAlias(t *testing.T) {
	service := &URLService{
		logger: logrus.New(),
	}

	validAliases := []string{
		"my-link",
		"user_123",
		"abc",
		"test-link-123",
		"MyCustomAlias",
		"valid-alias",
		"user123",
	}

	for _, alias := range validAliases {
		if err := service.validateCustomAlias(alias); err != nil {
			t.Errorf("validateCustomAlias(%s) should be valid, got error: %v", alias, err)
		}
	}

	invalidAliases := []string{
		"ab",       // too short
		"a",        // too short
		"",         // empty
		"api",      // reserved word
		"admin",    // reserved word
		"health",   // reserved word
		"my alias", // contains space
		"my@alias", // invalid character
		"my.alias", // invalid character
		"very-long-alias-that-exceeds-twenty-characters", // too long
	}

	for _, alias := range invalidAliases {
		if err := service.validateCustomAlias(alias); err == nil {
			t.Errorf("validateCustomAlias(%s) should be invalid, but passed", alias)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	service := &URLService{
		logger: logrus.New(),
	}

	testCases := []struct {
		input    string
		expected string
	}{
		{"example.com", "https://example.com"},
		{"http://example.com/", "http://example.com"},
		{"https://example.com/path/", "https://example.com/path"},
		{"https://example.com/path?query=value", "https://example.com/path?query=value"},
		{"http://subdomain.example.com", "http://subdomain.example.com"},
	}

	for _, tc := range testCases {
		result := service.normalizeURL(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeURL(%s) = %s; expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestAnalyticsEventCreation(t *testing.T) {
	// Test analytics event structure
	event := AnalyticsEvent{
		ShortCode: "abc123",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		Timestamp: time.Now(),
	}

	if event.ShortCode != "abc123" {
		t.Errorf("Expected short code 'abc123', got %s", event.ShortCode)
	}

	if event.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP address '192.168.1.1', got %s", event.IPAddress)
	}
}

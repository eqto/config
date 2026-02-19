package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// setupTestFile creates a temporary config file for testing purposes.
func setupTestFile(t *testing.T, content string) (string, func()) {
	t.Helper()
	// Use a fixed file path: tmp/test.conf
	tmpDir := "../tmp"
	tmpFile := filepath.Join(tmpDir, "test.conf")

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create tmp dir: %v", err)
	}

	if err := os.WriteFile(tmpFile, []byte(content), 0666); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Return the filename and a cleanup function.
	return tmpFile, func() {
		os.Remove(tmpFile)
	}
}

const testConfigContent = `
[app]
name = TestApp
version = 1.0

[server]
host = localhost
port = 8080

# Database array
-[database]
host = db1.example.com
user = user1

-[database]
host = db2.example.com
user = user2

# Proxy with multiple URLs
[proxy]
url = http://proxy1.com
url = http://proxy2.com
url = http://proxy3.com
`

func TestParseAndGet(t *testing.T) {
	filename, cleanup := setupTestFile(t, testConfigContent)
	defer cleanup()

	f, err := ParseFile(filename)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Test GetOr
	if got := f.GetOr("app.name", "default"); got != "TestApp" {
		t.Errorf(`GetOr("app.name", ...) = %q, want "TestApp"`, got)
	}
	if got := f.GetOr("app.nonexistent", "default"); got != "default" {
		t.Errorf(`GetOr("app.nonexistent", ...) = %q, want "default"`, got)
	}

	// Test GetInt
	if got := f.GetInt("server.port"); got != 8080 {
		t.Errorf(`GetInt("server.port") = %d, want 8080`, got)
	}
	if got := f.GetIntOr("server.nonexistent", 9999); got != 9999 {
		t.Errorf(`GetIntOr("server.nonexistent", ...) = %d, want 9999`, got)
	}
}

func TestGetArray(t *testing.T) {
	filename, cleanup := setupTestFile(t, testConfigContent)
	defer cleanup()

	f, err := ParseFile(filename)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	defer f.Close()

	dbs := f.GetArray("database")
	if len(dbs) != 2 {
		t.Fatalf("GetArray(\"database\") returned %d elements, want 2", len(dbs))
	}

	if got := dbs[0].Get("host"); got != "db1.example.com" {
		t.Errorf(`dbs[0].Get("host") = %q, want "db1.example.com"`, got)
	}
	if got := dbs[1].Get("user"); got != "user2" {
		t.Errorf(`dbs[1].Get("user") = %q, want "user2"`, got)
	}
}

func TestGetStringArray(t *testing.T) {
	filename, cleanup := setupTestFile(t, testConfigContent)
	defer cleanup()

	f, err := ParseFile(filename)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	defer f.Close()

	// Test multi-value key
	urls := f.GetStringArray("proxy.url")
	expectedURLs := []string{"http://proxy1.com", "http://proxy2.com", "http://proxy3.com"}
	if !reflect.DeepEqual(urls, expectedURLs) {
		t.Errorf("GetStringArray(\"proxy.url\") = %v, want %v", urls, expectedURLs)
	}

	// Test fallback for single-value key
	name := f.GetStringArray("app.name")
	expectedName := []string{"TestApp"}
	if !reflect.DeepEqual(name, expectedName) {
		t.Errorf("GetStringArray(\"app.name\") = %v, want %v", name, expectedName)
	}

	// Test non-existent key
	nonExistent := f.GetStringArray("app.nonexistent")
	if nonExistent != nil {
		t.Errorf("GetStringArray(\"app.nonexistent\") should be nil, got %v", nonExistent)
	}
}

func TestPutAndSave(t *testing.T) {
	filename, cleanup := setupTestFile(t, "") // Start with an empty file
	defer cleanup()

	f := NewFile(filename)
	defer f.Close()
	f.Put("user.name", "john")
	f.Put("user.email", "john@example.com")

	if err := f.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Read the file back and verify content
	f2, err := ParseFile(filename)
	if err != nil {
		t.Fatalf("ParseFile on saved file failed: %v", err)
	}
	defer f2.Close()

	if got := f2.Get("user.name"); got != "john" {
		t.Errorf(`Get("user.name") on saved file = %q, want "john"`, got)
	}
}

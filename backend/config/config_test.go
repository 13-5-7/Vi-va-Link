package config

import (
	"os"
	"os/exec"
	"testing"
)

// TestLoad_EmptyJWTSecret verifies that Load() causes a fatal error when JWT_SECRET is empty.
// This is the bug condition test: an empty JWT_SECRET must prevent server startup.
// Validates: Requirements 2.3
func TestLoad_EmptyJWTSecret(t *testing.T) {
	if os.Getenv("TEST_FATAL_SUBPROCESS") == "1" {
		// Running as subprocess: clear JWT_SECRET and call Load()
		os.Setenv("JWT_SECRET", "")
		Load()
		return
	}

	// Re-run this test as a subprocess to capture the fatal exit
	cmd := exec.Command(os.Args[0], "-test.run=TestLoad_EmptyJWTSecret")
	cmd.Env = append(os.Environ(), "TEST_FATAL_SUBPROCESS=1", "JWT_SECRET=")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected Load() to call log.Fatal when JWT_SECRET is empty, but it returned normally")
	}
	// Any non-zero exit (including log.Fatal's os.Exit(1)) means the fatal was triggered
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 0 {
			t.Fatal("expected non-zero exit code when JWT_SECRET is empty")
		}
		// Non-zero exit confirms log.Fatal was called — test passes
	} else {
		t.Fatalf("unexpected error type: %v", err)
	}
}

// TestLoad_ValidJWTSecret verifies that Load() returns a Config with the correct JWTSecret
// when JWT_SECRET is properly set.
// Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5
func TestLoad_ValidJWTSecret(t *testing.T) {
	const testSecret = "test-jwt-secret-value"

	original := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", original)

	os.Setenv("JWT_SECRET", testSecret)

	cfg := Load()

	if cfg == nil {
		t.Fatal("expected non-nil Config, got nil")
	}
	if cfg.JWTSecret != testSecret {
		t.Errorf("expected JWTSecret=%q, got %q", testSecret, cfg.JWTSecret)
	}
}

// TestLoad_DefaultValues verifies that Load() sets default values for PORT and OSRM_BASE_URL
// when they are not set, while still requiring JWT_SECRET.
func TestLoad_DefaultValues(t *testing.T) {
	const testSecret = "another-test-secret"

	origJWT := os.Getenv("JWT_SECRET")
	origPort := os.Getenv("PORT")
	origOSRM := os.Getenv("OSRM_BASE_URL")
	defer func() {
		os.Setenv("JWT_SECRET", origJWT)
		os.Setenv("PORT", origPort)
		os.Setenv("OSRM_BASE_URL", origOSRM)
	}()

	os.Setenv("JWT_SECRET", testSecret)
	os.Unsetenv("PORT")
	os.Unsetenv("OSRM_BASE_URL")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default Port=8080, got %q", cfg.Port)
	}
	if cfg.OSRMBaseURL != "https://router.project-osrm.org" {
		t.Errorf("expected default OSRMBaseURL, got %q", cfg.OSRMBaseURL)
	}
	if cfg.JWTSecret != testSecret {
		t.Errorf("expected JWTSecret=%q, got %q", testSecret, cfg.JWTSecret)
	}
}

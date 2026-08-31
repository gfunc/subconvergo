package config

import "testing"

// TestStartup_RefusesInsecurePublicDefault: startup validation must refuse a
// non-loopback bind when the API token is empty or still the legacy shipped
// default ("password"); loopback binds and token-secured public binds pass.
func TestStartup_RefusesInsecurePublicDefault(t *testing.T) {
	newSettings := func() *Settings {
		s := &Settings{}
		s.init()
		return s
	}

	// Public bind + empty token: refuse.
	s := newSettings()
	s.Server.Listen = "0.0.0.0"
	s.Common.APIAccessToken = ""
	if err := s.ValidateStartupSecurity(); err == nil {
		t.Fatal("public bind with empty token must be refused")
	}

	// Public bind + the legacy shipped default: refuse.
	s = newSettings()
	s.Server.Listen = "0.0.0.0"
	s.Common.APIAccessToken = "password"
	if err := s.ValidateStartupSecurity(); err == nil {
		t.Fatal("public bind with the legacy default token must be refused")
	}

	// Wildcard v6 bind + empty token: refuse.
	s = newSettings()
	s.Server.Listen = "::"
	s.Common.APIAccessToken = ""
	if err := s.ValidateStartupSecurity(); err == nil {
		t.Fatal("wildcard v6 bind with empty token must be refused")
	}

	// Loopback bind + empty token: allowed (protected endpoints stay
	// disabled via fail-closed token checks).
	s = newSettings()
	s.Server.Listen = "127.0.0.1"
	s.Common.APIAccessToken = ""
	if err := s.ValidateStartupSecurity(); err != nil {
		t.Fatalf("loopback bind with empty token must be allowed, got %v", err)
	}

	// Public bind + a real token: allowed.
	s = newSettings()
	s.Server.Listen = "0.0.0.0"
	s.Common.APIAccessToken = "a-real-token"
	if err := s.ValidateStartupSecurity(); err != nil {
		t.Fatalf("public bind with configured token must be allowed, got %v", err)
	}
}

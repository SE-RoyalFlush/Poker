package main

import "testing"

func TestServerAddrFromEnvDefaultsToLocalPort(t *testing.T) {
	t.Setenv("PORT", "")

	if got := serverAddrFromEnv(); got != ":8080" {
		t.Fatalf("serverAddrFromEnv() = %q, want %q", got, ":8080")
	}
}

func TestServerAddrFromEnvUsesConfiguredPort(t *testing.T) {
	t.Setenv("PORT", "9090")

	if got := serverAddrFromEnv(); got != ":9090" {
		t.Fatalf("serverAddrFromEnv() = %q, want %q", got, ":9090")
	}
}

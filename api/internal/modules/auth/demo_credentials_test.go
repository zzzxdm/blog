package auth

import "testing"

func TestIsDefaultDemoCredential(t *testing.T) {
	if !isDefaultDemoCredential("admin@example.com", "password") {
		t.Fatal("expected admin demo credential to match")
	}
	if !isDefaultDemoCredential("Admin@Example.com", "password") {
		t.Fatal("expected email normalization")
	}
	if isDefaultDemoCredential("admin@example.com", "strong-password") {
		t.Fatal("expected non-default password to pass")
	}
	if isDefaultDemoCredential("someone@example.com", "password") {
		t.Fatal("expected non-demo email to pass")
	}
}

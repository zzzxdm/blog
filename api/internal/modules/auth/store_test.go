package auth

import (
	"strings"
	"testing"
)

func TestNewUsersGetOpaqueIDs(t *testing.T) {
	store := NewMemoryStore()

	registered, _, err := store.Register(RegisterRequest{
		Email:       "pretty.id@example.com",
		Password:    "password",
		DisplayName: "Pretty ID",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	assertOpaqueUserID(t, registered.ID, "pretty")

	invited, _, err := store.InviteUser(InviteUserRequest{
		Email:       "invited.person@example.com",
		DisplayName: "Invited Person",
		Role:        "author",
	})
	if err != nil {
		t.Fatalf("InviteUser returned error: %v", err)
	}
	assertOpaqueUserID(t, invited.ID, "invited")
}

func TestInvitePasswordSetupMarksEmailVerified(t *testing.T) {
	store := NewMemoryStore()

	invited, secrets, err := store.InviteUser(InviteUserRequest{
		Email:       "invited-verify@example.com",
		DisplayName: "Invited Verify",
		Role:        "author",
	})
	if err != nil {
		t.Fatalf("InviteUser returned error: %v", err)
	}
	if invited.EmailVerified {
		t.Fatal("expected invited user to start unverified")
	}
	if secrets.InitialPassword == "" {
		t.Fatal("expected initial password")
	}
	if _, _, err := store.Authenticate("invited-verify@example.com", secrets.InitialPassword); err != nil {
		t.Fatalf("expected initial password to authenticate: %v", err)
	}

	if err := store.ResetPassword(secrets.ResetToken, "new-password"); err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}
	if _, _, err := store.Authenticate("invited-verify@example.com", secrets.InitialPassword); err == nil {
		t.Fatal("expected initial password to stop working after reset")
	}

	user, _, err := store.Authenticate("invited-verify@example.com", "new-password")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if !user.EmailVerified {
		t.Fatal("expected password setup through emailed token to verify email")
	}
}

func assertOpaqueUserID(t *testing.T, id string, emailLocalPart string) {
	t.Helper()

	if !strings.HasPrefix(id, "usr_") {
		t.Fatalf("id = %q, want usr_ prefix", id)
	}
	if strings.Contains(id, emailLocalPart) || strings.Contains(id, ".") || strings.Contains(id, "@") {
		t.Fatalf("id = %q should not expose email-derived text", id)
	}
}

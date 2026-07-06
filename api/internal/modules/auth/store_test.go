package auth

import "testing"

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

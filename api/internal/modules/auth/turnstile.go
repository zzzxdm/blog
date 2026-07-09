package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type TurnstileVerifier interface {
	Verify(ctx context.Context, secret string, token string, remoteIP string) (bool, error)
}

type HTTPTurnstileVerifier struct {
	client *http.Client
}

func NewHTTPTurnstileVerifier() *HTTPTurnstileVerifier {
	return &HTTPTurnstileVerifier{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (verifier *HTTPTurnstileVerifier) Verify(ctx context.Context, secret string, token string, remoteIP string) (bool, error) {
	secret = strings.TrimSpace(secret)
	token = strings.TrimSpace(token)
	if secret == "" || token == "" {
		return false, nil
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if strings.TrimSpace(remoteIP) != "" {
		form.Set("remoteip", strings.TrimSpace(remoteIP))
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, turnstileVerifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := verifier.client.Do(request)
	if err != nil {
		return false, fmt.Errorf("verify turnstile: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return false, fmt.Errorf("verify turnstile status: %d", response.StatusCode)
	}

	var payload struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return false, fmt.Errorf("decode turnstile response: %w", err)
	}

	return payload.Success, nil
}

package github

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// AppTokens mints installation tokens for a GitHub App and caches them until
// shortly before they expire.
type AppTokens struct {
	AppID          int64
	InstallationID int64
	Key            *rsa.PrivateKey
	BaseURL        string
	HTTP           *http.Client

	mu      sync.Mutex
	token   string
	expires time.Time
}

// ErrPrivateKey is returned for an unreadable App private key.
var ErrPrivateKey = errors.New("invalid GitHub App private key")

// ParsePrivateKey parses a PEM-encoded PKCS#1 or PKCS#8 RSA key.
func ParsePrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrPrivateKey
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPrivateKey, err)
	}
	rk, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%w: not an RSA key", ErrPrivateKey)
	}
	return rk, nil
}

// JWT returns an App JWT valid for about nine minutes.
func (a *AppTokens) JWT(now time.Time) (string, error) {
	enc := base64.RawURLEncoding
	header := enc.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims, err := json.Marshal(map[string]any{
		// Backdated to tolerate clock drift, as GitHub recommends.
		"iat": now.Add(-60 * time.Second).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": strconv.FormatInt(a.AppID, 10),
	})
	if err != nil {
		return "", err
	}
	signing := header + "." + enc.EncodeToString(claims)
	sum := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, a.Key, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return signing + "." + enc.EncodeToString(sig), nil
}

// Token implements TokenSource.
func (a *AppTokens) Token(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && time.Until(a.expires) > 5*time.Minute {
		return a.token, nil
	}
	jwt, err := a.JWT(time.Now())
	if err != nil {
		return "", err
	}
	c := &Client{BaseURL: a.BaseURL, HTTP: a.HTTP, Tokens: StaticToken(jwt)}
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 30 * time.Second}
	}
	var out struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	path := fmt.Sprintf("/app/installations/%d/access_tokens", a.InstallationID)
	if err := c.do(ctx, http.MethodPost, path, nil, &out); err != nil {
		return "", err
	}
	a.token, a.expires = out.Token, out.ExpiresAt
	return a.token, nil
}

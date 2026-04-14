// Package cookies builds the httpOnly session cookies used by AuthService.
//
// ConnectRPC handlers can set response headers via connect.Response.Header().
// This package only produces *http.Cookie values — callers emit them with
// resp.Header().Add("Set-Cookie", c.String()).
package cookies

import (
	"net/http"
	"time"
)

const (
	// AccessName is the name of the short-lived JWT cookie.
	AccessName = "gp_access"
	// RefreshName is the name of the opaque refresh-token cookie.
	RefreshName = "gp_refresh"
)

// Config holds session cookie settings.
type Config struct {
	// Domain is the cookie Domain attribute (empty = host-only).
	Domain string
	// Secure sets the Secure flag. Must be true in production.
	Secure bool
	// SameSite controls cross-site behavior. Default: SameSiteLaxMode.
	SameSite http.SameSite
}

// Access returns the access-token cookie.
func Access(token string, ttl time.Duration, cfg Config) *http.Cookie {
	return base(AccessName, token, ttl, cfg)
}

// Refresh returns the refresh-token cookie. When longLived is false the
// cookie is a session cookie (no Max-Age) so it clears when the browser exits.
func Refresh(token string, ttl time.Duration, longLived bool, cfg Config) *http.Cookie {
	if !longLived {
		c := base(RefreshName, token, 0, cfg)
		// Session cookie: no MaxAge, no Expires.
		return c
	}
	return base(RefreshName, token, ttl, cfg)
}

// ClearAccess returns a cookie that clears the access-token cookie.
func ClearAccess(cfg Config) *http.Cookie {
	return clear(AccessName, cfg)
}

// ClearRefresh returns a cookie that clears the refresh-token cookie.
func ClearRefresh(cfg Config) *http.Cookie {
	return clear(RefreshName, cfg)
}

// Get extracts a cookie value from an http.Header. Empty string + false when
// the cookie is absent.
func Get(h http.Header, name string) (string, bool) {
	r := http.Request{Header: h}
	c, err := r.Cookie(name)
	if err != nil || c == nil {
		return "", false
	}
	return c.Value, true
}

func base(name, value string, ttl time.Duration, cfg Config) *http.Cookie {
	sameSite := cfg.SameSite
	if sameSite == 0 {
		sameSite = http.SameSiteLaxMode
	}
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   cfg.Domain,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: sameSite,
	}
	if ttl > 0 {
		c.MaxAge = int(ttl.Seconds())
		c.Expires = time.Now().Add(ttl)
	}
	return c
}

func clear(name string, cfg Config) *http.Cookie {
	sameSite := cfg.SameSite
	if sameSite == 0 {
		sameSite = http.SameSiteLaxMode
	}
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   cfg.Domain,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: sameSite,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
}

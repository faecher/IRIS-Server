// SPDX-License-Identifier: EUPL-1.2

// Package traccar provides utilities for processing messages from Traccar
package traccar

import (
	"IRIS-Server/internal/config"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var (
	protocolRegex = regexp.MustCompile(`^https?:\/\/`) // matches http:// or https://

	// ErrNoAuthMethod is returned when neither auth token nor email/password are provided for Traccar authentication
	ErrNoAuthMethod = errors.New("no valid authentication method provided, " +
		"please set either TRACCAR_AUTH_TOKEN or both TRACCAR_EMAIL and TRACCAR_PASSWORD")
	// ErrNoCookies is returned when no cookies are received from the authentication response, which are required for websocket authentication
	ErrNoCookies = errors.New("no cookies received from authentication response")
	// ErrAuthFailed is returned when the authentication request to Traccar does not return a 200 OK status
	ErrAuthFailed = errors.New("authentication failed with status code")
)

const (
	// API paths always include base path + specific endpoint paths
	traccarAPIBasePath = "/api"
	sessionPath        = "/api/session"
	socketPath         = "/socket"

	initialBackoff = 1 * time.Second
	maxBackoff     = 1 * time.Minute

	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second // must be < pongWait
	writeWait  = 10 * time.Second
)

// RunTraccarWebsocketListener connects to the Traccar websocket and handles incoming messages.
// This function will run until the given context is Done
func RunTraccarWebsocketListener(ctx context.Context, cfg config.TraccarConfig) {
	backoff := initialBackoff

	socketURL := getSocketURL(cfg)
	slog.Info("Connecting to Traccar websocket at " + socketURL)

	for ctx.Err() == nil {
		cookies, err := getSessionCookie(ctx, cfg)
		if err != nil {
			slog.Error("Traccar auth failed", "error", err, "retry_in", backoff)

			backoff, err = waitWithContext(ctx, backoff)
			if err != nil {
				return
			}

			continue
		}

		conn, resp, err := dialWebsocket(ctx, socketURL, cookies)
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		if err != nil {
			slog.Error("Error connecting to Traccar websocket", "error", err, "retry_in", backoff)

			backoff, err = waitWithContext(ctx, backoff)
			if err != nil {
				return
			}
			continue
		}

		// Connection successfully estabilshed
		backoff = initialBackoff
		err = runSocketSession(ctx, conn)
		_ = conn.Close()

		if ctx.Err() != nil {
			return
		}

		slog.Warn("Traccar websocket ended, reconnecting", "error", err, "retry_in", backoff)

		backoff, err = waitWithContext(ctx, backoff)
		if err != nil {
			return
		}
	}
}

func getAPIBaseURL(cfg config.TraccarConfig) string {
	if protocolRegex.MatchString(cfg.Host) {
		return cfg.Host + traccarAPIBasePath
	}

	return "https://" + cfg.Host + traccarAPIBasePath
}

func getSocketURL(cfg config.TraccarConfig) string {
	baseURL := getAPIBaseURL(cfg)

	url, hashttps := strings.CutPrefix(baseURL, "https://")
	if hashttps {
		return "wss://" + url + socketPath
	}

	url, hashttp := strings.CutPrefix(baseURL, "http://")
	if hashttp {
		return "ws://" + url + socketPath
	}

	// Default to wss if no protocol is specified, but this should not happen due to getAPIBaseURL logic
	return "wss://" + baseURL + socketPath
}

func getSessionCookie(ctx context.Context, cfg config.TraccarConfig) ([]*http.Cookie, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	var req *http.Request
	var err error

	switch {
	case cfg.AuthToken != "":
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, getAPIBaseURL(cfg)+sessionPath+"?token="+url.QueryEscape(cfg.AuthToken), nil)
	case cfg.Email != "" && cfg.Password != "":
		form := url.Values{}
		form.Set("email", cfg.Email)
		form.Set("password", cfg.Password)
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, getAPIBaseURL(cfg)+sessionPath, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	default:
		return nil, ErrNoAuthMethod
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrAuthFailed, resp.StatusCode)
	}

	cookies := jar.Cookies(req.URL)
	if len(cookies) == 0 {
		return nil, ErrNoCookies
	}

	return cookies, nil
}

func dialWebsocket(ctx context.Context, socketURL string, cookies []*http.Cookie) (*websocket.Conn, *http.Response, error) {
	header := http.Header{}
	for _, cookie := range cookies {
		header.Add("Cookie", cookie.String())
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, resp, err := dialer.DialContext(ctx, socketURL, header)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to websocket: %w", err)
	}

	return conn, resp, nil
}

func nextBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > maxBackoff {
		return maxBackoff
	}
	return next
}

// waitWithContext waits for the specified duration or until the context is done, whichever comes first.
// It returns an error if the context is cancelled, or nil if the wait completed successfully.
// The returned duration is the suggested next backoff duration, which is the next power of 2 up to maxBackoff.
// In case of context cancellation, the original duration is returned to avoid increasing backoff on cancellation.
func waitWithContext(ctx context.Context, duration time.Duration) (time.Duration, error) {
	waitCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	<-waitCtx.Done()
	if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
		return nextBackoff(duration), nil
	}
	return duration, fmt.Errorf("context cancelled: %w", waitCtx.Err())
}

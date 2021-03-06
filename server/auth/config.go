// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/luci/luci-go/server/auth/authdb"
	"github.com/luci/luci-go/server/auth/signing"
)

// Config contains global configuration of the auth library.
//
// This configuration adjusts the library to the particular execution
// environment (GAE, Flex, whatever). It contains concrete implementations of
// various interfaces used by the library.
//
// It lives in the context and must be installed there by some root middleware
// (via ModifyContext call).
type Config struct {
	// DBProvider is a callback that returns most recent DB instance.
	//
	// DB represents a snapshot of user groups used for authorization checks.
	DBProvider func(c context.Context) (authdb.DB, error)

	// Signer possesses the service's private key and can sign blobs with it.
	//
	// It provides the bundle with corresponding public keys and information about
	// the service account they belong too (the service's own identity).
	//
	// Used to implement '/auth/api/v1/server/(certificates|info)' routes.
	Signer signing.Signer

	// AccessTokenProvider knows how to generate OAuth2 access token for the
	// service account belonging to the server itself.
	//
	// Should implement caching itself, if appropriate. Returned tokens are
	// expected to live for at least 1 min.
	AccessTokenProvider func(c context.Context, scopes []string) (*oauth2.Token, error)

	// AnonymousTransport returns http.RoundTriper that can make unauthenticated
	// HTTP requests.
	//
	// The returned round tripper is assumed to be bound to the context and won't
	// outlive it.
	AnonymousTransport func(c context.Context) http.RoundTripper

	// Cache implements a strongly consistent cache.
	//
	// Usually backed by memcache. Should do namespacing itself (i.e. the auth
	// library assumes full ownership of the keyspace).
	Cache Cache

	// IsDevMode is true when running the server locally during development.
	//
	// Setting this to true changes default deadlines. For instance, GAE dev
	// server is known to be very slow and deadlines tuned for production
	// environment are too limiting.
	IsDevMode bool
}

// ModifyConfig makes a context with a derived configuration.
//
// It grabs current configuration from the context (if any), passes it to the
// callback, and puts whatever callback returns into a derived context.
func ModifyConfig(c context.Context, cb func(Config) Config) context.Context {
	var cfg Config
	if cur := getConfig(c); cur != nil {
		cfg = *cur
	}
	cfg = cb(cfg)
	return setConfig(c, &cfg)
}

// adjustedTimeout returns `t` if IsDevMode is false or >=1 min if true.
func (cfg *Config) adjustedTimeout(t time.Duration) time.Duration {
	if !cfg.IsDevMode {
		return t
	}
	if t > time.Minute {
		return t
	}
	return time.Minute
}

var cfgContextKey = "auth.Config context key"

// setConfig replaces the configuration in the context.
func setConfig(c context.Context, cfg *Config) context.Context {
	return context.WithValue(c, &cfgContextKey, cfg)
}

// getConfig returns the config stored in the context (or nil if not there).
func getConfig(c context.Context) *Config {
	val, _ := c.Value(&cfgContextKey).(*Config)
	return val
}

////////////////////////////////////////////////////////////////////////////////
// Helpers that extract stuff from the config.

// GetDB returns most recent snapshot of authorization database using DBProvider
// installed in the context via 'ModifyConfig'.
//
// If no factory is installed, returns DB that forbids everything and logs
// errors. It is often good enough for unit tests that do not care about
// authorization, and still not horribly bad if accidentally used in production.
func GetDB(c context.Context) (authdb.DB, error) {
	if cfg := getConfig(c); cfg != nil && cfg.DBProvider != nil {
		return cfg.DBProvider(c)
	}
	return authdb.ErroringDB{Error: ErrNotConfigured}, nil
}

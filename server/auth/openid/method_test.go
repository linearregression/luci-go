// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package openid

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/luci/luci-go/common/clock/testclock"
	"github.com/luci/luci-go/server/auth"
	"github.com/luci/luci-go/server/auth/authtest"
	"github.com/luci/luci-go/server/secrets/testsecrets"
	"github.com/luci/luci-go/server/settings"
	"golang.org/x/net/context"

	. "github.com/luci/luci-go/common/testing/assertions"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFullFlow(t *testing.T) {
	Convey("with test context", t, func(c C) {
		ctx := context.Background()
		ctx = settings.Use(ctx, settings.New(&settings.MemoryStorage{}))
		ctx, _ = testclock.UseTime(ctx, time.Unix(1442540000, 0))
		ctx = testsecrets.Use(ctx)

		var ts *httptest.Server
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {

			case "/discovery":
				w.Write([]byte(fmt.Sprintf(`{
					"authorization_endpoint": "%s/authorization",
					"token_endpoint": "%s/token",
					"userinfo_endpoint": "%s/userinfo"
				}`, ts.URL, ts.URL, ts.URL)))

			case "/token":
				c.So(r.ParseForm(), ShouldBeNil)
				c.So(r.Form, ShouldResemble, url.Values{
					"redirect_uri":  {"http://fake/redirect"},
					"client_id":     {"client_id"},
					"client_secret": {"client_secret"},
					"code":          {"omg_auth_code"},
					"grant_type":    {"authorization_code"},
				})
				w.Write([]byte(`{"token_type": "Bearer", "access_token": "token_blah"}`))

			case "/userinfo":
				c.So(r.Header.Get("Authorization"), ShouldEqual, "Bearer token_blah")
				w.Write([]byte(`{
					"sub": "user_id_sub",
					"email": "user@example.com",
					"name": "Some Dude",
					"picture": "https://picture/url/photo.jpg"
				}`))

			default:
				http.Error(w, "Not found", http.StatusNotFound)
			}
		}))

		cfg := Settings{
			DiscoveryURL: ts.URL + "/discovery",
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			RedirectURI:  "http://fake/redirect",
		}
		So(StoreSettings(ctx, &cfg, "who", "why"), ShouldBeNil)

		method := AuthMethod{
			SessionStore:        &authtest.MemorySessionStore{},
			Insecure:            true,
			IncompatibleCookies: []string{"wrong_cookie"},
		}

		Convey("Full flow", func() {
			So(method.Warmup(ctx), ShouldBeNil)

			// Generate login URL.
			loginURL, err := method.LoginURL(ctx, "/destination")
			So(err, ShouldBeNil)
			So(loginURL, ShouldEqual, "/auth/openid/login?r=%2Fdestination")

			// "Visit" login URL.
			req, err := http.NewRequest("GET", "http://fake"+loginURL, nil)
			So(err, ShouldBeNil)
			rec := httptest.NewRecorder()
			method.loginHandler(ctx, rec, req, nil)

			// It asks us to visit authorizarion endpoint.
			So(rec.Code, ShouldEqual, http.StatusFound)
			parsed, err := url.Parse(rec.Header().Get("Location"))
			So(err, ShouldBeNil)
			So(parsed.Host, ShouldEqual, ts.URL[len("http://"):])
			So(parsed.Path, ShouldEqual, "/authorization")
			So(parsed.Query(), ShouldResemble, url.Values{
				"client_id":     {"client_id"},
				"redirect_uri":  {"http://fake/redirect"},
				"response_type": {"code"},
				"scope":         {"openid email profile"},
				"state": {
					"AXsiX2kiOiIxNDQyNTQwMDAwMDAwIiwiZGVzdF91cmwiOiIvZGVzdGluYXRpb24iLC" +
						"Job3N0X3VybCI6ImZha2UifUFtzG6wPbuvHG2mY_Wf6eQ_Eiu7n3_Tf6GmRcse1g" +
						"YE",
				},
			})

			// Pretend we've done it. OpenID redirects user's browser to callback URI.
			// `callbackHandler` will call /token and /userinfo fake endpoints exposed
			// by testserver.
			callbackParams := url.Values{}
			callbackParams.Set("code", "omg_auth_code")
			callbackParams.Set("state", parsed.Query().Get("state"))
			req, err = http.NewRequest("GET", "http://fake/redirect?"+callbackParams.Encode(), nil)
			So(err, ShouldBeNil)
			rec = httptest.NewRecorder()
			method.callbackHandler(ctx, rec, req, nil)

			// We should be redirected to the login page, with session cookie set.
			expectedCookie := "oid_session=AXsiX2kiOiIxNDQyNTQwMDAwMDAwIiwic2lkIjoi" +
				"dXNlcl9pZF9zdWIvMSJ9PmRzaOv-mS0PMHkve897iiELNmpiLi_j3ICG1VKuNCs"
			So(rec.Code, ShouldEqual, http.StatusFound)
			So(rec.Header().Get("Location"), ShouldEqual, "/destination")
			So(rec.Header().Get("Set-Cookie"), ShouldEqual,
				expectedCookie+"; Path=/; Expires=Sun, 18 Oct 2015 01:18:20 UTC; Max-Age=2591100; HttpOnly")

			// Use the cookie to authenticate some call.
			req, err = http.NewRequest("GET", "http://fake/something", nil)
			So(err, ShouldBeNil)
			req.Header.Add("Cookie", expectedCookie)
			user, err := method.Authenticate(ctx, req)
			So(err, ShouldBeNil)
			So(user, ShouldResemble, &auth.User{
				Identity: "user:user@example.com",
				Email:    "user@example.com",
				Name:     "Some Dude",
				Picture:  "https://picture/url/s64/photo.jpg",
			})

			// Now generate URL to and visit logout page.
			logoutURL, err := method.LogoutURL(ctx, "/another_destination")
			So(err, ShouldBeNil)
			So(logoutURL, ShouldEqual, "/auth/openid/logout?r=%2Fanother_destination")
			req, err = http.NewRequest("GET", "http://fake"+logoutURL, nil)
			So(err, ShouldBeNil)
			req.Header.Add("Cookie", expectedCookie)
			rec = httptest.NewRecorder()
			method.logoutHandler(ctx, rec, req, nil)

			// Should be redirected to destination with the cookie killed.
			So(rec.Code, ShouldEqual, http.StatusFound)
			So(rec.Header().Get("Location"), ShouldEqual, "/another_destination")
			So(rec.Header().Get("Set-Cookie"), ShouldEqual,
				"oid_session=deleted; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 UTC; Max-Age=0")
		})
	})
}

func TestCallbackHandleEdgeCases(t *testing.T) {
	Convey("with test context", t, func(c C) {
		ctx := context.Background()
		ctx = settings.Use(ctx, settings.New(&settings.MemoryStorage{}))
		ctx, _ = testclock.UseTime(ctx, time.Unix(1442540000, 0))
		ctx = testsecrets.Use(ctx)

		method := AuthMethod{SessionStore: &authtest.MemorySessionStore{}}

		call := func(query map[string]string) *httptest.ResponseRecorder {
			q := url.Values{}
			for k, v := range query {
				q.Add(k, v)
			}
			req, err := http.NewRequest("GET", "/auth/openid/callback?"+q.Encode(), nil)
			c.So(err, ShouldBeNil)
			req.Host = "fake.com"
			rec := httptest.NewRecorder()
			method.callbackHandler(ctx, rec, req, nil)
			return rec
		}

		Convey("handles 'error'", func() {
			rec := call(map[string]string{"error": "Omg, error"})
			So(rec.Code, ShouldEqual, 400)
			So(rec.Body.String(), ShouldEqual, "OpenID login error: Omg, error\n")
		})

		Convey("handles no 'code'", func() {
			rec := call(map[string]string{})
			So(rec.Code, ShouldEqual, 400)
			So(rec.Body.String(), ShouldEqual, "Missing 'code' parameter\n")
		})

		Convey("handles no 'state'", func() {
			rec := call(map[string]string{"code": "123"})
			So(rec.Code, ShouldEqual, 400)
			So(rec.Body.String(), ShouldEqual, "Missing 'state' parameter\n")
		})

		Convey("handles bad 'state'", func() {
			rec := call(map[string]string{"code": "123", "state": "garbage"})
			So(rec.Code, ShouldEqual, 400)
			So(rec.Body.String(), ShouldEqual, "Failed to validate 'state' token\n")
		})

		Convey("handles redirect to another host", func() {
			state := map[string]string{
				"dest_url": "/",
				"host_url": "non-default.fake.com",
			}
			stateTok, err := openIDStateToken.Generate(ctx, nil, state, 0)
			So(err, ShouldBeNil)

			rec := call(map[string]string{"code": "123", "state": stateTok})
			So(rec.Code, ShouldEqual, 302)
			So(rec.Header().Get("Location"), ShouldEqual,
				"https://non-default.fake.com/auth/openid/callback?"+
					"code=123&state=AXsiX2kiOiIxNDQyNTQwMDAwMDAwIiwiZGVzdF91cmwiOiIvIiw"+
					"iaG9zdF91cmwiOiJub24tZGVmYXVsdC5mYWtlLmNvbSJ92y0UJtCrN2qGYbcbCiZsV"+
					"9OdFEa3zAauzz4lmwPJLwI")
		})
	})
}

func TestNotConfigured(t *testing.T) {
	Convey("Returns ErrNotConfigured is on SessionStore", t, func() {
		ctx := context.Background()
		method := AuthMethod{}

		_, err := method.LoginURL(ctx, "/")
		So(err, ShouldEqual, ErrNotConfigured)

		_, err = method.LogoutURL(ctx, "/")
		So(err, ShouldEqual, ErrNotConfigured)

		_, err = method.Authenticate(ctx, &http.Request{})
		So(err, ShouldEqual, ErrNotConfigured)
	})
}

func TestBadDestinationURLs(t *testing.T) {
	Convey("Rejects bad destination URLs", t, func() {
		ctx := context.Background()
		method := AuthMethod{SessionStore: &authtest.MemorySessionStore{}}

		_, err := method.LoginURL(ctx, "http://somesite")
		So(err, ShouldErrLike, "openid: dest URL in LoginURL or LogoutURL must be relative")

		_, err = method.LogoutURL(ctx, "http://somesite")
		So(err, ShouldErrLike, "openid: dest URL in LoginURL or LogoutURL must be relative")
	})
}

func TestDisabled(t *testing.T) {
	Convey("Disabled == true works for API", t, func() {
		ctx := context.Background()
		method := AuthMethod{
			SessionStore: &authtest.MemorySessionStore{},
			Disabled:     true,
		}

		_, err := method.LoginURL(ctx, "/")
		So(err, ShouldEqual, ErrDisabled)

		_, err = method.LogoutURL(ctx, "/")
		So(err, ShouldEqual, ErrDisabled)

		_, err = method.Authenticate(ctx, &http.Request{})
		So(err, ShouldEqual, ErrDisabled)
	})

	Convey("Disabled == true works for URLs", t, func() {
		ctx := context.Background()
		method := AuthMethod{
			SessionStore: &authtest.MemorySessionStore{},
			Disabled:     true,
		}

		req, err := http.NewRequest("GET", "http://fake/auth/openid/login", nil)
		So(err, ShouldBeNil)
		rec := httptest.NewRecorder()
		method.loginHandler(ctx, rec, req, nil)
		So(rec.Code, ShouldEqual, 400)
		So(rec.Body.String(), ShouldEqual, "OpenID authentication is disabled\n")

		req, err = http.NewRequest("GET", "http://fake/auth/openid/logout", nil)
		So(err, ShouldBeNil)
		rec = httptest.NewRecorder()
		method.logoutHandler(ctx, rec, req, nil)
		So(rec.Code, ShouldEqual, 400)
		So(rec.Body.String(), ShouldEqual, "OpenID authentication is disabled\n")

		req, err = http.NewRequest("GET", "http://fake/auth/openid/callback", nil)
		So(err, ShouldBeNil)
		rec = httptest.NewRecorder()
		method.callbackHandler(ctx, rec, req, nil)
		So(rec.Code, ShouldEqual, 400)
		So(rec.Body.String(), ShouldEqual, "OpenID authentication is disabled\n")
	})
}
// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package googleoauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"

	"github.com/luci/luci-go/common/clock"
	"github.com/luci/luci-go/common/clock/testclock"
	"github.com/luci/luci-go/common/gcloud/iam"

	. "github.com/smartystreets/goconvey/convey"
)

type token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func TestGetAccessToken(t *testing.T) {
	ctx, _ := testclock.UseTime(context.Background(), testclock.TestRecentTimeLocal)
	issuedAt := testclock.TestRecentTimeLocal.Add(-15 * time.Second).Unix()

	Convey("Happy path", t, func() {
		req, tok, url, err := call(ctx, JwtFlowParams{
			ServiceAccount: "account@example.com",
			Scopes:         []string{"a", "b"},
		}, 200, token{"abc", "Bearer", 3600})

		So(err, ShouldBeNil)

		// Request parameters are valid.
		So(req.Get("grant_type"), ShouldEqual, "urn:ietf:params:oauth:grant-type:jwt-bearer")
		claims := deconstructJWT(req.Get("assertion"))
		So(claims, ShouldResemble, iam.ClaimSet{
			Iss:   "account@example.com",
			Scope: "a b",
			Aud:   url,
			Iat:   issuedAt,
			Exp:   issuedAt + 3600,
		})

		// Response is understood.
		So(tok, ShouldResemble, &oauth2.Token{
			AccessToken: "abc",
			TokenType:   "Bearer",
			Expiry:      clock.Now(ctx).Add(time.Hour).UTC(),
		})
	})

	Convey("Uses Bearer as default", t, func() {
		_, tok, _, err := call(ctx, JwtFlowParams{
			ServiceAccount: "account@example.com",
			Scopes:         []string{"a", "b"},
		}, 200, token{"def", "", 3600})

		So(err, ShouldBeNil)
		So(tok, ShouldResemble, &oauth2.Token{
			AccessToken: "def",
			TokenType:   "Bearer",
			Expiry:      clock.Now(ctx).Add(time.Hour).UTC(),
		})
	})

	Convey("Bad HTTP code", t, func() {
		_, _, _, err := call(ctx, JwtFlowParams{
			ServiceAccount: "account@example.com",
			Scopes:         []string{"a", "b"},
		}, 403, nil)
		So(err, ShouldHaveSameTypeAs, &googleapi.Error{})
		So(err.(*googleapi.Error).Code, ShouldEqual, 403)
	})

	Convey("Zero 'expires_in'", t, func() {
		_, _, _, err := call(ctx, JwtFlowParams{
			ServiceAccount: "account@example.com",
			Scopes:         []string{"a", "b"},
		}, 200, token{"zzz", "", 0})
		So(err, ShouldNotBeNil)
	})

	Convey("Negative 'expires_in'", t, func() {
		_, _, _, err := call(ctx, JwtFlowParams{
			ServiceAccount: "account@example.com",
			Scopes:         []string{"a", "b"},
		}, 200, token{"zzz", "", -100})
		So(err, ShouldNotBeNil)
	})

	Convey("Not valid JSON", t, func() {
		_, _, _, err := call(ctx, JwtFlowParams{
			ServiceAccount: "account@example.com",
			Scopes:         []string{"a", "b"},
		}, 200, "zzzzzz")
		So(err, ShouldNotBeNil)
	})
}

func call(ctx context.Context, params JwtFlowParams, status int, resp interface{}) (url.Values, *oauth2.Token, string, error) {
	values := make(chan url.Values, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			panic("not a POST")
		}
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		values <- r.Form
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	params.tokenEndpoint = ts.URL
	params.Signer = fakeSigner{}

	tok, err := GetAccessToken(ctx, params)
	req := <-values
	return req, tok, ts.URL, err
}

func deconstructJWT(token string) (claims iam.ClaimSet) {
	parts := strings.Split(token, ".") // <header>.<claims>.<signature>
	So(len(parts), ShouldEqual, 3)

	// We are interested only in claim set. The headers and signature are mocked
	// by fakeSigner, no sense it checking them.
	claimsBin, err := base64.RawURLEncoding.DecodeString(parts[1])
	So(err, ShouldBeNil)
	So(json.Unmarshal(claimsBin, &claims), ShouldBeNil)

	return
}

type fakeSigner struct{}

func (fakeSigner) SignJWT(c context.Context, serviceAccount string, cs *iam.ClaimSet) (keyName, signedJwt string, err error) {
	blob, err := json.Marshal(cs)
	if err != nil {
		return "", "", err
	}
	claimsB64 := base64.RawURLEncoding.EncodeToString(blob)
	return "unused key id", "fake_hdr." + claimsB64 + ".fake_sig", nil
}

package webui

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHTTPLoginSurvivesLegacySecureCookie(t *testing.T) {
	useConfigFile(t, Config{Password: "correct-horse"})
	s := loginServer()
	w := postLogin(t, s, url.Values{"password": {"correct-horse"}})
	c := cookieNamed(w, sessionCookie)
	if c == nil || c.Secure || c.Name == "backpack_session" {
		t.Fatalf("HTTP login cookie: %+v", c)
	}
	r := httptest.NewRequest(http.MethodGet, "http://panel/", nil)
	r.AddCookie(&http.Cookie{Name: "backpack_session", Value: "old-secure-session"})
	r.AddCookie(c)
	if _, ok := s.identify(r); !ok {
		t.Fatal("new HTTP session rejected")
	}
}
func TestHTTPAndHTTPSUseDifferentCookieNames(t *testing.T) {
	plain := httptest.NewRequest("GET", "http://panel/", nil)
	secure := httptest.NewRequest("GET", "https://panel/", nil)
	secure.TLS = &tls.ConnectionState{}
	for _, name := range []string{sessionCookie, twoFactorCookie} {
		a := authCookie(plain, name, "http", sessionTTL)
		b := authCookie(secure, name, "https", sessionTTL)
		if a.Name == b.Name || a.Secure || !b.Secure {
			t.Fatal("cookie transports overlap")
		}
		if clearedCookie(secure, name).Name != b.Name {
			t.Fatal("logout clears wrong cookie")
		}
	}
}

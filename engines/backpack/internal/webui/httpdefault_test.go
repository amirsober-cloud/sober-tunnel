package webui

import (
	"net/http/httptest"
	"testing"
)

func TestLegacyAutomaticTLSUsesHTTPWithoutCertificate(t *testing.T) {
	useConfigFile(t, Config{Password: "keep-password", BasePath: "keep-path", HTTPS: true})
	c := Load()
	if c.HTTPS || c.Scheme() != "http" || c.Password != "keep-password" || c.BasePath != "keep-path" {
		t.Fatalf("unexpected configuration: %+v", c)
	}
	if secureRequest(httptest.NewRequest("GET", "http://panel/", nil)) {
		t.Fatal("HTTP session would receive a Secure cookie")
	}
}
func TestConfiguredACMEKeepsHTTPS(t *testing.T) {
	useConfigFile(t, Config{HTTPS: true, TLSDomain: "panel.example.com"})
	if !Load().HTTPS {
		t.Fatal("configured HTTPS disabled")
	}
}

package config

import "testing"

func TestIsProxyURI_ShadowSocksLongScheme(t *testing.T) {
	if !IsProxyURI("shadowsocks://aes-256-gcm:password@example.com:8388#Name") {
		t.Fatalf("expected shadowsocks:// URI to be recognized")
	}
}

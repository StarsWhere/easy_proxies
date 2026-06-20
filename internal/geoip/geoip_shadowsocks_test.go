package geoip

import "testing"

func TestExtractSSHost_LegacyWholePayloadBase64(t *testing.T) {
	got := extractSSHost("ss://YWVzLTI1Ni1nY206dGVzdC1wYXNzd29yZEBleGFtcGxlLmNvbTo4Mzg4#Synthetic-Node@example.net:8388")
	if got != "example.com" {
		t.Fatalf("extractSSHost() = %q, want example.com", got)
	}
}

func TestExtractSSHost_SIP002Base64(t *testing.T) {
	got := extractSSHost("ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8389#Test")
	if got != "example.com" {
		t.Fatalf("extractSSHost() = %q, want example.com", got)
	}
}

func TestExtractHostFromURI_ShadowSocksLongScheme(t *testing.T) {
	got := extractHostFromURI("shadowsocks://aes-256-gcm:password@example.com:8388#Test")
	if got != "example.com" {
		t.Fatalf("extractHostFromURI() = %q, want example.com", got)
	}
}

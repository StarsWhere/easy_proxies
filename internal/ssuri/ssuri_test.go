package ssuri

import "testing"

func TestParseSupportedFormats(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		method   string
		password string
		server   string
		port     int
		fragment string
	}{
		{
			name:     "legacy whole payload base64",
			raw:      "ss://YWVzLTI1Ni1nY206dGVzdC1wYXNzd29yZEBleGFtcGxlLmNvbTo4Mzg4#Synthetic-Node@example.net:8388",
			method:   "aes-256-gcm",
			password: "test-password",
			server:   "example.com",
			port:     8388,
			fragment: "Synthetic-Node@example.net:8388",
		},
		{
			name:     "sip002 base64 userinfo",
			raw:      "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8389#Test",
			method:   "aes-256-gcm",
			password: "password",
			server:   "example.com",
			port:     8389,
			fragment: "Test",
		},
		{
			name:     "plain userinfo",
			raw:      "ss://aes-256-gcm:password@example.com:8388#Plain",
			method:   "aes-256-gcm",
			password: "password",
			server:   "example.com",
			port:     8388,
			fragment: "Plain",
		},
		{
			name:     "url safe base64 without padding",
			raw:      "ss://MjAyMi1ibGFrZTMtYWVzLTEyOC1nY206cC9hc3M_d29yZA@example.com:8443#URLSafe",
			method:   "2022-blake3-aes-128-gcm",
			password: "p/ass?word",
			server:   "example.com",
			port:     8443,
			fragment: "URLSafe",
		},
		{
			name:     "fragment containing at sign and port",
			raw:      "ss://aes-256-gcm:password@example.com:9000#node@example.org:9000",
			method:   "aes-256-gcm",
			password: "password",
			server:   "example.com",
			port:     9000,
			fragment: "node@example.org:9000",
		},
		{
			name:     "ipv6 host",
			raw:      "ss://aes-256-gcm:password@[2001:db8::1]:8388#IPv6",
			method:   "aes-256-gcm",
			password: "password",
			server:   "2001:db8::1",
			port:     8388,
			fragment: "IPv6",
		},
		{
			name:     "default port",
			raw:      "ss://aes-256-gcm:password@example.com#DefaultPort",
			method:   "aes-256-gcm",
			password: "password",
			server:   "example.com",
			port:     8388,
			fragment: "DefaultPort",
		},
		{
			name:     "shadowsocks scheme",
			raw:      "shadowsocks://aes-256-gcm:password@example.com:8388#LongScheme",
			method:   "aes-256-gcm",
			password: "password",
			server:   "example.com",
			port:     8388,
			fragment: "LongScheme",
		},
		{
			name:     "legacy password containing at sign",
			raw:      "ss://YWVzLTI1Ni1nY206cEBzc0BleGFtcGxlLmNvbTo4Mzg4#PasswordAt",
			method:   "aes-256-gcm",
			password: "p@ss",
			server:   "example.com",
			port:     8388,
			fragment: "PasswordAt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.raw)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got.Method != tt.method || got.Password != tt.password || got.Server != tt.server || got.Port != tt.port || got.Fragment != tt.fragment {
				t.Fatalf("Parse() = %+v, want method=%q password=%q server=%q port=%d fragment=%q", got, tt.method, tt.password, tt.server, tt.port, tt.fragment)
			}
		})
	}
}

func TestParseQuery(t *testing.T) {
	got, err := Parse("ss://aes-256-gcm:password@example.com:8388?plugin=v2ray-plugin%3Btls#Query")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got.Query.Get("plugin") != "v2ray-plugin;tls" {
		t.Fatalf("plugin query = %q", got.Query.Get("plugin"))
	}
}

func TestParseInvalidURIs(t *testing.T) {
	tests := []string{
		"ss://@@@",
		"ss://YWVzLTI1Ni1nY20tcGFzc3dvcmQ@example.com:8388",
		"ss://:password@example.com:8388",
		"ss://aes-256-gcm:@example.com:8388",
		"ss://aes-256-gcm:password@example.com:notaport",
		"ss://aes-256-gcm:password@[2001:db8::1:8388",
		"ss://aes-256-gcm:password@2001:db8::1:8388",
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			if _, err := Parse(raw); err == nil {
				t.Fatalf("Parse() error = nil, want error")
			}
		})
	}
}

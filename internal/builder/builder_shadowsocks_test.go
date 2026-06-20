package builder

import (
	"strings"
	"testing"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func TestBuildNodeOutbound_ShadowsocksLegacyWholePayloadBase64(t *testing.T) {
	outbound, err := buildNodeOutbound("test-ss", "ss://YWVzLTI1Ni1nY206dGVzdC1wYXNzd29yZEBleGFtcGxlLmNvbTo4Mzg4#Synthetic-Node@example.net:8388", false)
	if err != nil {
		t.Fatalf("build node outbound failed: %v", err)
	}

	if outbound.Type != C.TypeShadowsocks {
		t.Fatalf("expected type %q, got %q", C.TypeShadowsocks, outbound.Type)
	}
	opts, ok := outbound.Options.(*option.ShadowsocksOutboundOptions)
	if !ok {
		t.Fatalf("expected *option.ShadowsocksOutboundOptions, got %T", outbound.Options)
	}
	if opts.Method != "aes-256-gcm" {
		t.Fatalf("expected method aes-256-gcm, got %q", opts.Method)
	}
	if opts.Password != "test-password" {
		t.Fatalf("expected decoded password, got %q", opts.Password)
	}
	if opts.Server != "example.com" {
		t.Fatalf("expected server example.com, got %q", opts.Server)
	}
	if opts.ServerPort != 8388 {
		t.Fatalf("expected port 8388, got %d", opts.ServerPort)
	}
}

func TestBuildNodeOutbound_ShadowsocksPluginStillUnsupported(t *testing.T) {
	_, err := buildNodeOutbound("test-ss", "ss://aes-256-gcm:password@example.com:8388?plugin=v2ray-plugin", false)
	if err == nil {
		t.Fatalf("expected plugin error, got nil")
	}
	if !strings.Contains(err.Error(), "shadowsocks plugin not supported") {
		t.Fatalf("expected plugin unsupported error, got %v", err)
	}
}

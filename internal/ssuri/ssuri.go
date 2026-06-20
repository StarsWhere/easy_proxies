package ssuri

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const defaultPort = 8388

// URI is the parsed form of a Shadowsocks URI.
type URI struct {
	Method   string
	Password string
	Server   string
	Port     int
	Query    url.Values
	Fragment string
}

// Parse accepts SIP002 Shadowsocks URIs and the legacy whole-payload base64
// format: ss://base64(method:password@host:port)#name.
func Parse(raw string) (URI, error) {
	schemeEnd := strings.Index(raw, "://")
	if schemeEnd < 0 {
		return URI{}, errors.New("shadowsocks uri missing scheme")
	}
	scheme := strings.ToLower(raw[:schemeEnd])
	if scheme != "ss" && scheme != "shadowsocks" {
		return URI{}, fmt.Errorf("unsupported shadowsocks scheme %q", raw[:schemeEnd])
	}

	rest := raw[schemeEnd+3:]
	fragment := ""
	if hashIdx := strings.Index(rest, "#"); hashIdx >= 0 {
		fragment, _ = url.PathUnescape(rest[hashIdx+1:])
		rest = rest[:hashIdx]
	}

	query := url.Values{}
	if queryIdx := strings.Index(rest, "?"); queryIdx >= 0 {
		parsedQuery, err := url.ParseQuery(rest[queryIdx+1:])
		if err != nil {
			return URI{}, fmt.Errorf("parse shadowsocks query: %w", err)
		}
		query = parsedQuery
		rest = rest[:queryIdx]
	}
	if rest == "" {
		return URI{}, errors.New("shadowsocks uri missing payload")
	}

	if strings.Contains(rest, "@") {
		return parseSIP002(rest, query, fragment)
	}
	return parseLegacy(rest, query, fragment)
}

func parseSIP002(rest string, query url.Values, fragment string) (URI, error) {
	atIdx := strings.LastIndex(rest, "@")
	if atIdx <= 0 || atIdx == len(rest)-1 {
		return URI{}, errors.New("shadowsocks uri must include userinfo and host")
	}

	userInfo := rest[:atIdx]
	hostPort := rest[atIdx+1:]
	method, password, err := parseUserInfo(userInfo)
	if err != nil {
		return URI{}, err
	}
	server, port, err := parseHostPort(hostPort)
	if err != nil {
		return URI{}, err
	}
	return URI{Method: method, Password: password, Server: server, Port: port, Query: query, Fragment: fragment}, nil
}

func parseLegacy(encoded string, query url.Values, fragment string) (URI, error) {
	decoded, err := decodeBase64(encoded)
	if err != nil {
		return URI{}, fmt.Errorf("decode legacy shadowsocks payload: %w", err)
	}

	decodedText := string(decoded)
	atIdx := strings.LastIndex(decodedText, "@")
	if atIdx <= 0 || atIdx == len(decodedText)-1 {
		return URI{}, errors.New("legacy shadowsocks payload must be method:password@host:port")
	}
	methodPart := decodedText[:atIdx]
	hostPort := decodedText[atIdx+1:]
	method, password, err := parsePlainUserInfo(methodPart)
	if err != nil {
		return URI{}, err
	}
	server, port, err := parseHostPort(hostPort)
	if err != nil {
		return URI{}, err
	}
	return URI{Method: method, Password: password, Server: server, Port: port, Query: query, Fragment: fragment}, nil
}

func parseUserInfo(userInfo string) (string, string, error) {
	if decoded, err := decodeBase64(userInfo); err == nil {
		if method, password, plainErr := parsePlainUserInfo(string(decoded)); plainErr == nil {
			return method, password, nil
		}
	}

	unescaped, err := url.PathUnescape(userInfo)
	if err != nil {
		return "", "", fmt.Errorf("decode shadowsocks userinfo: %w", err)
	}
	return parsePlainUserInfo(unescaped)
}

func parsePlainUserInfo(userInfo string) (string, string, error) {
	method, password, ok := strings.Cut(userInfo, ":")
	if !ok {
		return "", "", errors.New("shadowsocks userinfo format must be method:password")
	}
	if method == "" {
		return "", "", errors.New("shadowsocks method is required")
	}
	if password == "" {
		return "", "", errors.New("shadowsocks password is required")
	}
	return method, password, nil
}

func parseHostPort(hostPort string) (string, int, error) {
	if hostPort == "" {
		return "", 0, errors.New("shadowsocks host is required")
	}

	host := ""
	port := defaultPort
	if strings.HasPrefix(hostPort, "[") {
		end := strings.Index(hostPort, "]")
		if end < 0 {
			return "", 0, errors.New("invalid IPv6 host")
		}
		host = hostPort[1:end]
		remainder := hostPort[end+1:]
		if remainder != "" {
			if !strings.HasPrefix(remainder, ":") {
				return "", 0, errors.New("invalid IPv6 host port")
			}
			parsedPort, err := parsePort(remainder[1:])
			if err != nil {
				return "", 0, err
			}
			port = parsedPort
		}
	} else if strings.Count(hostPort, ":") > 1 {
		return "", 0, errors.New("IPv6 shadowsocks host must use [addr]:port")
	} else if h, p, err := net.SplitHostPort(hostPort); err == nil {
		host = h
		parsedPort, err := parsePort(p)
		if err != nil {
			return "", 0, err
		}
		port = parsedPort
	} else if strings.Contains(hostPort, ":") {
		h, p, _ := strings.Cut(hostPort, ":")
		host = h
		parsedPort, err := parsePort(p)
		if err != nil {
			return "", 0, err
		}
		port = parsedPort
	} else {
		host = hostPort
	}

	if host == "" {
		return "", 0, errors.New("shadowsocks host is required")
	}
	unescapedHost, err := url.PathUnescape(host)
	if err != nil {
		return "", 0, fmt.Errorf("decode shadowsocks host: %w", err)
	}
	return unescapedHost, port, nil
}

func parsePort(portText string) (int, error) {
	if portText == "" {
		return 0, errors.New("shadowsocks port is required")
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return 0, fmt.Errorf("invalid shadowsocks port %q", portText)
	}
	if port <= 0 || port > 65535 {
		return 0, fmt.Errorf("invalid shadowsocks port %q", portText)
	}
	return port, nil
}

func decodeBase64(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty base64 value")
	}

	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	var lastErr error
	for _, enc := range encodings {
		decoded, err := enc.DecodeString(s)
		if err == nil {
			return decoded, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

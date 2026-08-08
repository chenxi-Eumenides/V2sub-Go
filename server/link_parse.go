package server

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/Ericwyn/v2sub/conf"
)

// ParseSsLink parses a shadowsocks ss:// link into a VmessJson.
// Supports:
//   - ss://base64(method:password@host:port)#tag  (SIP002)
//   - ss://base64(method:password)@host:port#tag  (legacy)
//   - ss://method:password@host:port#tag          (plain)
func ParseSsLink(ssStr string) *conf.VmessJson {
	s := strings.TrimPrefix(ssStr, "ss://")

	var tag string
	if i := strings.Index(s, "#"); i >= 0 {
		tag = decodeTag(s[i+1:])
		s = s[:i]
	}

	var userInfo, hostPart string
	if at := strings.Index(s, "@"); at >= 0 {
		userInfo = s[:at]
		hostPart = s[at+1:]
		// 老格式: base64(method:password)@host:port
		if decoded := decodeBase64Silent(userInfo); decoded != "" && !strings.Contains(decoded, "@") {
			userInfo = decoded
		}
	} else {
		// SIP002: 整体 base64 编码的 method:password@host:port
		decoded := decodeBase64Silent(s)
		if decoded == "" {
			return nil
		}
		method, password, host, port := splitSsInfo(decoded)
		if host == "" || port == "" {
			return nil
		}
		return newSsVmess(method, password, host, port, tag)
	}

	method, password, host, port := splitSsInfo(userInfo + "@" + hostPart)
	if host == "" || port == "" {
		return nil
	}
	return newSsVmess(method, password, host, port, tag)
}

// ParseVlessLink parses a vless:// link into a VmessJson.
// Format: vless://uuid@host:port?query#tag
func ParseVlessLink(vlessStr string) *conf.VmessJson {
	s := strings.TrimPrefix(vlessStr, "vless://")

	var tag string
	if i := strings.Index(s, "#"); i >= 0 {
		tag = decodeTag(s[i+1:])
		s = s[:i]
	}

	var query string
	if i := strings.Index(s, "?"); i >= 0 {
		query = s[i+1:]
		s = s[:i]
	}

	at := strings.Index(s, "@")
	if at < 0 {
		return nil
	}
	uuid := s[:at]
	host, port := splitHostPort(s[at+1:])
	if host == "" || port == "" {
		return nil
	}

	v := &conf.VmessJson{
		ID:     uuid,
		Addr:   host,
		Port:   portToRawMessage(port),
		Ps:     tag,
		Net:    "tcp",
		Remark: tag,
	}
	for _, kv := range strings.Split(query, "&") {
		if kv == "" {
			continue
		}
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		if unescaped, err := url.QueryUnescape(val); err == nil {
			val = unescaped
		}
		switch key {
		case "type":
			v.Net = val
		case "headerType":
			v.Type = val
		case "host":
			v.Host = val
		case "path":
			v.Path = val
		case "security":
			v.Security = val
		case "flow":
			v.Flow = val
		case "sni":
			v.Sni = val
		case "fp":
			v.Fp = val
		case "pbk":
			v.Pbk = val
		case "sid":
			v.Sid = val
		case "spx":
			v.Spx = val
		}
	}
	return v
}

func newSsVmess(method, password, host, port, tag string) *conf.VmessJson {
	return &conf.VmessJson{
		Method:   method,
		Password: password,
		Addr:     host,
		Port:     portToRawMessage(port),
		Ps:       tag,
		Remark:   tag,
	}
}

// decodeTag decodes URL-encoded link tags (e.g. "%40" -> "@").
func decodeTag(tag string) string {
	if decoded, err := url.QueryUnescape(tag); err == nil {
		return decoded
	}
	return tag
}

// splitSsInfo splits "method:password@host:port" into its parts.
func splitSsInfo(s string) (method, password, host, port string) {
	userInfo, hostPart, ok := strings.Cut(s, "@")
	if !ok {
		return "", "", "", ""
	}
	method, password, _ = strings.Cut(userInfo, ":")
	host, port = splitHostPort(hostPart)
	return method, password, host, port
}

// splitHostPort splits "host:port", supporting IPv6 like "[::1]:8080".
func splitHostPort(s string) (host, port string) {
	if strings.HasPrefix(s, "[") {
		if i := strings.Index(s, "]:"); i >= 0 {
			return s[1:i], s[i+2:]
		}
		return strings.Trim(s, "[]"), ""
	}
	host, port, _ = strings.Cut(s, ":")
	return host, port
}

func portToRawMessage(port string) json.RawMessage {
	return json.RawMessage(strconv.Quote(port))
}

// decodeBase64Silent decodes standard/url-safe base64 without logging.
// Returns empty string on failure.
func decodeBase64Silent(s string) string {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(decoded)
}

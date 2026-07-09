package parser

import (
	"strings"

	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/server"
	"github.com/Ericwyn/v2sub/utils/decode"
)

// VmessBase64Parser handles the standard V2Ray subscription format:
// base64-encoded body containing newline-separated vmess:// links.
type VmessBase64Parser struct{}

func (p *VmessBase64Parser) CanParse(body string) bool {
	// Try base64 decode and check if result contains vmess:// and/or ss:// links
	decoded := decode.Base64Decode(body)
	return decoded != "" && (strings.Contains(decoded, "vmess://") ||
		strings.Contains(decoded, "ss://") ||
		strings.Contains(decoded, "trojan://") ||
		strings.Contains(decoded, "vless://"))
}

func (p *VmessBase64Parser) Parse(body string, subName string) []conf.SerSubEntry {
	decoded := decode.Base64Decode(body)
	if decoded == "" {
		return nil
	}
	entries := make([]conf.SerSubEntry, 0)
	for _, line := range strings.Split(decoded, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "vmess://") {
			vmess := server.ParseVmessLink(line)
			if vmess != nil {
				entries = append(entries, conf.SerSubEntry{
					SubName: subName,
					Source:  line,
					Vmess:   *vmess,
				})
			}
		}
		// Future: handle ss://, trojan://, vless:// links
	}
	return entries
}

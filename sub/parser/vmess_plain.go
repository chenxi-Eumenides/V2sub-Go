package parser

import (
	"strings"

	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/server"
	"github.com/Ericwyn/v2sub/utils/decode"
)

// VmessPlainParser handles plain vmess:// links (no base64 wrapping).
// Some subscription providers return the links directly.
type VmessPlainParser struct{}

func (p *VmessPlainParser) CanParse(body string) bool {
	// Body contains at least one vmess:// link and is NOT base64-encoded
	return strings.Contains(body, "vmess://") && !isBase64Body(body)
}

func (p *VmessPlainParser) Parse(body string, subName string) []conf.SerSubEntry {
	entries := make([]conf.SerSubEntry, 0)
	for _, line := range strings.Split(body, "\n") {
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
	}
	return entries
}

// isBase64Body checks if the body is likely base64-encoded
// (already handled by VmessBase64Parser, so this parser skips it)
func isBase64Body(body string) bool {
	decoded := decode.Base64Decode(body)
	return decoded != "" && strings.Contains(decoded, "vmess://")
}

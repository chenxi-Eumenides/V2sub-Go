package parser

import (
	"strings"

	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/utils/decode"
)

// Base64LinkParser handles subscriptions wrapped in base64, containing
// newline-separated vmess://, ss://, trojan:// or vless:// links.
type Base64LinkParser struct{}

func (p *Base64LinkParser) CanParse(body string) bool {
	decoded := decode.Base64Decode(body)
	if decoded == "" {
		return false
	}
	return containsLink(decoded)
}

func (p *Base64LinkParser) Parse(body string, subName string) []conf.SerSubEntry {
	decoded := decode.Base64Decode(body)
	if decoded == "" {
		return nil
	}
	return parseLines(decoded, subName)
}

// containsLink checks if the text contains any known subscription link prefix.
func containsLink(text string) bool {
	for _, prefix := range []string{"vmess://", "ss://", "trojan://", "vless://"} {
		if strings.Contains(text, prefix) {
			return true
		}
	}
	return false
}

// parseLines splits the text into lines and parses each supported link.
func parseLines(text string, subName string) []conf.SerSubEntry {
	entries := make([]conf.SerSubEntry, 0)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if entry := parseLinkLine(line, subName); entry != nil {
			entries = append(entries, *entry)
		}
	}
	return entries
}

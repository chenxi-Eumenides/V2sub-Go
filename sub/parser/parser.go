package parser

import (
	"strings"

	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/server"
)

// SubParser parses a raw subscription response body into server entries.
// Each implementation handles a specific subscription format.
type SubParser interface {
	// CanParse returns true if this parser can handle the raw body.
	// It should be a lightweight check (e.g., check for format markers).
	CanParse(body string) bool

	// Parse extracts server entries from the body for the given subscription name.
	Parse(body string, subName string) []conf.SerSubEntry
}

// Parse tries each registered parser in order, returning results from the
// first parser that can handle the body.
// Returns nil if no parser can handle the format.
func Parse(body string, subName string) []conf.SerSubEntry {
	parsers := []SubParser{
		&Base64LinkParser{},
		&PlainLinkParser{},
	}
	for _, p := range parsers {
		if p.CanParse(body) {
			return p.Parse(body, subName)
		}
	}
	return nil
}

// parseLinkLine parses a single subscription link (vmess/ss/vless/trojan)
// into a server entry. Returns nil if the line is not a supported link.
func parseLinkLine(line string, subName string) *conf.SerSubEntry {
	var protocol string
	var vmess *conf.VmessJson
	switch {
	case strings.HasPrefix(line, "vmess://"):
		protocol = "vmess"
		vmess = server.ParseVmessLink(line)
	case strings.HasPrefix(line, "vless://"):
		protocol = "vless"
		vmess = server.ParseVlessLink(line)
	case strings.HasPrefix(line, "ss://"):
		protocol = "ss"
		vmess = server.ParseSsLink(line)
	}
	if vmess == nil {
		return nil
	}
	return &conf.SerSubEntry{
		SubName:  subName,
		Source:   line,
		Protocol: protocol,
		Vmess:    *vmess,
	}
}

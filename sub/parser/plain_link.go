package parser

import (
	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/utils/decode"
)

// PlainLinkParser handles plain links (no base64 wrapping).
// Some subscription providers return the links directly.
type PlainLinkParser struct{}

func (p *PlainLinkParser) CanParse(body string) bool {
	// Body contains at least one link and is NOT base64-encoded
	return containsLink(body) && !isBase64Body(body)
}

func (p *PlainLinkParser) Parse(body string, subName string) []conf.SerSubEntry {
	return parseLines(body, subName)
}

// isBase64Body checks if the body is likely base64-encoded
// (already handled by Base64LinkParser, so this parser skips it)
func isBase64Body(body string) bool {
	decoded := decode.Base64Decode(body)
	return decoded != "" && containsLink(decoded)
}

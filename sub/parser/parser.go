package parser

import "github.com/Ericwyn/v2sub/conf"

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
	// Register parsers here (ordered by specificity)
	parsers := []SubParser{
		&VmessBase64Parser{},
		&VmessPlainParser{},
	}
	for _, p := range parsers {
		if p.CanParse(body) {
			return p.Parse(body, subName)
		}
	}
	return nil
}

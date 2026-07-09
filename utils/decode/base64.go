package decode

import (
	"encoding/base64"
	"strings"

	"github.com/Ericwyn/v2sub/utils/log"
)

func Base64Decode(base64Str string) string {
	decoded, err := base64.StdEncoding.DecodeString(base64Str)
	decodestr := string(decoded)
	if err == nil {
		return decodestr
	} else {
		log.E("base64 decode fail")
		log.E(base64Str)
		return ""
	}
}

func VmessBase64Decode(vmessBase64Str string) string {
	// Normalize URL-safe chars to standard base64
	s := strings.ReplaceAll(vmessBase64Str, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")

	// Add padding if needed (standard base64 length must be multiple of 4)
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.E("vmess base64 decode fail")
		log.E(vmessBase64Str)
		return ""
	}
	return string(decoded)
}

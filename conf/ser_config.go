package conf

import (
	"encoding/json"
	"strconv"
)

type V2SubJson struct {
	Conf ConfConfig `json:"conf"`
	Rule RuleConfig `json:"rule"`
	Ser  SerConfig  `json:"ser"`
}

type ConfConfig struct {
	SocksPort         int  `json:"socksPort"`
	HttpPort          int  `json:"httpPort"`
	AllowLocalConnect bool `json:"allowLocalConnect"`
	BypassLan         bool `json:"bypassLan"`
	CopyDatToOfficial bool `json:"copyDatToOfficial"`
}

type RuleConfig struct {
	Proxy  []string `json:"proxy"`
	Direct []string `json:"direct"`
}

type SerConfig struct {
	Current SerCurrent       `json:"current"`
	Sub     []SerSubEntry    `json:"sub"`
	SubUrls map[string]string `json:"subUrls,omitempty"`
}

type SerCurrent struct {
	SubName string `json:"subName"`
	Index   int    `json:"index"`
}

type SerSubEntry struct {
	SubName  string    `json:"subName"`
	Source   string    `json:"source"`
	Protocol string    `json:"protocol"` // vmess / ss / vless
	Vmess    VmessJson `json:"vmess"`
}

type VmessJson struct {
	Ps     string          `json:"ps"`
	Addr   string          `json:"add"`
	Port   json.RawMessage `json:"port"`
	ID     string          `json:"id"`
	Aid    int             `json:"aid"`
	Net    string          `json:"net"`
	Type   string          `json:"type"`
	TLS    string          `json:"tls"`
	Remark string          `json:"remark,omitempty"`
	Host   string          `json:"host,omitempty"`
	Path   string          `json:"path,omitempty"`
	Sni    string          `json:"sni,omitempty"`
	Fp     string          `json:"fp,omitempty"`
	Alpn   string          `json:"alpn,omitempty"`

	// shadowsocks (ss://)
	Method   string `json:"method,omitempty"`
	Password string `json:"password,omitempty"`

	// vless (vless://)
	Flow     string `json:"flow,omitempty"`
	Security string `json:"security,omitempty"` // tls / reality / none
	Pbk      string `json:"pbk,omitempty"`      // publicKey (reality)
	Sid      string `json:"sid,omitempty"`      // shortId (reality)
	Spx      string `json:"spx,omitempty"`      // spiderX (reality)
}

// GetPort returns the port as a string, handling both JSON int and string formats.
// Some providers return "port":10086 (int), others return "port":"52623" (string).
func (v *VmessJson) GetPort() string {
	var portInt int
	if err := json.Unmarshal(v.Port, &portInt); err == nil {
		return strconv.Itoa(portInt)
	}
	var portStr string
	json.Unmarshal(v.Port, &portStr)
	return portStr
}

// GetPs returns the preferred node name, using Remark as fallback when Ps is empty.
// Some providers use "remark" instead of "ps" for the node display name.
func (v *VmessJson) GetPs() string {
	if v.Ps != "" {
		return v.Ps
	}
	return v.Remark
}

// ProtocolName returns the node protocol, defaulting to vmess
// for entries saved before the protocol field existed.
func (e *SerSubEntry) ProtocolName() string {
	if e.Protocol == "" {
		return "vmess"
	}
	return e.Protocol
}

var GlobalConf ConfConfig
var GlobalRule RuleConfig
var GlobalSer SerConfig

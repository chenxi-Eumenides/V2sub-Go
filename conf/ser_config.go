package conf

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
	Current SerCurrent    `json:"current"`
	Sub     []SerSubEntry `json:"sub"`
	SubUrls map[string]string `json:"subUrls,omitempty"`
}

type SerCurrent struct {
	SubName string `json:"subName"`
	Index   int    `json:"index"`
}

type SerSubEntry struct {
	SubName string    `json:"subName"`
	Source  string    `json:"source"`
	Vmess   VmessJson `json:"vmess"`
}

type VmessJson struct {
	Ps   string `json:"ps"`
	Addr string `json:"add"`
	Port string `json:"port"`
	ID   string `json:"id"`
	Aid  int    `json:"aid"`
	Net  string `json:"net"`
	Type string `json:"type"`
	TLS  string `json:"tls"`
}

var GlobalConf ConfConfig
var GlobalRule RuleConfig
var GlobalSer SerConfig

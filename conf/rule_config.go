package conf

// DomainRule 自定义域名规则配置
type DomainRule struct {
	Proxy  []string `json:"proxy"`  // 走代理的域名列表
	Direct []string `json:"direct"` // 直连的域名列表
}

// 默认配置
var RuleConfigNow DomainRule

// RuleConfigName 规则配置文件名
const RuleConfigName = "rules.json"

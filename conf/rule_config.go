package conf

var RuleConfigNow RuleConfig

func init() {
	RuleConfigNow.Proxy = make([]string, 0)
	RuleConfigNow.Direct = make([]string, 0)
}

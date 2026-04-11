package rule

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Ericwyn/v2sub/ajax"
	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/utils/log"
	"github.com/Ericwyn/v2sub/utils/param"
	"github.com/Ericwyn/v2sub/utils/storage"
)

const (
	geoSiteURL  = "https://raw.githubusercontent.com/Loyalsoldier/v2ray-rules-dat/release/geosite.dat"
	geoSiteFile = "geosite.dat"
)

func ParseArgs(args []string) {
	param.AssistParamLength(args, 1)

	conf.LoadLocalConfig()

	switch args[0] {
	case "update", "u":
		UpdateGeoSite()
	case "proxy":
		HandleProxy(args[1:])
	case "direct":
		HandleDirect(args[1:])
	case "list", "l":
		ListRules()
	default:
		log.E("rule args error, use -h to get help")
	}
}

func HandleProxy(args []string) {
	if len(args) < 1 {
		ListProxyRules()
		return
	}
	switch args[0] {
	case "add", "a":
		param.AssistParamLength(args, 2)
		AddDomain("proxy", args[1])
	case "remove", "r":
		param.AssistParamLength(args, 2)
		RemoveDomain("proxy", args[1])
	default:
		log.E("unknown proxy sub command: ", args[0])
	}
}

func HandleDirect(args []string) {
	if len(args) < 1 {
		ListDirectRules()
		return
	}
	switch args[0] {
	case "add", "a":
		param.AssistParamLength(args, 2)
		AddDomain("direct", args[1])
	case "remove", "r":
		param.AssistParamLength(args, 2)
		RemoveDomain("direct", args[1])
	default:
		log.E("unknown direct sub command: ", args[0])
	}
}

func UpdateGeoSite() {
	log.I("start download geosite data from Loyalsoldier/v2ray-rules-dat...")

	ajax.Send(ajax.Request{
		Url:    geoSiteURL,
		Method: ajax.GET,
		Success: func(response *ajax.Response) {
			storageDir := storage.GetConfigDirPath()
			err := ioutil.WriteFile(storageDir+"/"+geoSiteFile, []byte(response.Body), 0644)
			if err != nil {
				log.E("save geosite.dat failed: ", err.Error())
				return
			}
			log.I("geosite.dat updated successfully")
			log.I("saved to: ", storageDir+"/"+geoSiteFile)
		},
		Fail: func(status int, errMsg string) {
			log.E("download geosite.dat failed, status: ", status)
		},
	})
}

func AddDomain(ruleType string, domain string) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		log.E("domain cannot be empty")
		return
	}

	var list *[]string
	if ruleType == "proxy" {
		list = &conf.RuleConfigNow.Proxy
	} else {
		list = &conf.RuleConfigNow.Direct
	}

	for _, d := range *list {
		if d == domain {
			log.I("domain '", domain, "' already exists in ", ruleType, " rules")
			return
		}
	}

	*list = append(*list, domain)
	conf.FlushRuleConfig()
	log.I("added '", domain, "' to ", ruleType, " rules")
}

func RemoveDomain(ruleType string, domain string) {
	domain = strings.ToLower(strings.TrimSpace(domain))

	var list *[]string
	if ruleType == "proxy" {
		list = &conf.RuleConfigNow.Proxy
	} else {
		list = &conf.RuleConfigNow.Direct
	}

	found := false
	newList := make([]string, 0, len(*list))
	for _, d := range *list {
		if d == domain {
			found = true
		} else {
			newList = append(newList, d)
		}
	}

	if !found {
		log.I("domain '", domain, "' not found in ", ruleType, " rules")
		return
	}

	*list = newList
	conf.FlushRuleConfig()
	log.I("removed '", domain, "' from ", ruleType, " rules")
}

func ListProxyRules() {
	fmt.Println("=======================================================")
	fmt.Println("Proxy rules:")
	if len(conf.RuleConfigNow.Proxy) == 0 {
		fmt.Println("  (empty)")
	} else {
		for i, domain := range conf.RuleConfigNow.Proxy {
			fmt.Printf("  [%d] %s\n", i+1, domain)
		}
	}
	fmt.Println("=======================================================")
}

func ListDirectRules() {
	fmt.Println("=======================================================")
	fmt.Println("Direct rules:")
	if len(conf.RuleConfigNow.Direct) == 0 {
		fmt.Println("  (empty)")
	} else {
		for i, domain := range conf.RuleConfigNow.Direct {
			fmt.Printf("  [%d] %s\n", i+1, domain)
		}
	}
	fmt.Println("=======================================================")
}

func ListRules() {
	ListProxyRules()
	ListDirectRules()
}

func FetchFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("http status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

package rule

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/utils/log"
	"github.com/Ericwyn/v2sub/utils/storage"
)

const (
	geoSiteURL = "https://raw.githubusercontent.com/Loyalsoldier/v2ray-rules-dat/release/geosite.dat"
	geoipURL   = "https://raw.githubusercontent.com/Loyalsoldier/v2ray-rules-dat/release/geoip.dat"
)

func ParseArgs(args []string) {
	if len(args) < 1 {
		return
	}
	switch args[0] {
	case "update", "u":
		updateGeoSite()
	case "proxy":
		handleRule("proxy", args[1:])
	case "direct":
		handleRule("direct", args[1:])
	case "list", "l":
		listRules()
	default:
		log.E("rule 参数错误, 使用 -rule 查看帮助")
	}
}

func updateGeoSite() {
	log.I("开始下载 geo 数据...")
	ok1 := downloadFile("geosite.dat", geoSiteURL)
	ok2 := downloadFile("geoip.dat", geoipURL)

	if conf.GlobalConf.CopyDatToOfficial && (ok1 || ok2) {
		fmt.Println("正在复制到官方位置 ...")
		failed := make([]string, 0)
		if ok1 && !copyToOfficial("geosite.dat") {
			failed = append(failed, "geosite.dat")
		}
		if ok2 && !copyToOfficial("geoip.dat") {
			failed = append(failed, "geoip.dat")
		}
		if len(failed) > 0 {
			fmt.Println("可手动复制:")
			for _, name := range failed {
				src := storage.GetConfigDirPath() + "/" + name
				dest := "/usr/local/bin/" + name
				fmt.Println("  sudo cp " + src + " " + dest)
			}
			fmt.Println("或设置 V2RAY_LOCATION_ASSET 环境变量指定 V2Ray 读取 dat 文件的目录")
		}
	}
}

func downloadFile(name, url string) bool {
	dest := storage.GetConfigDirPath() + "/" + name
	log.I("正在下载:", url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("下载 " + name + " 失败: " + err.Error())
		fmt.Println("可手动执行: wget -O", dest, url)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("下载 %s 失败: HTTP %d\n", name, resp.StatusCode)
		return false
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取 " + name + " 响应失败: " + err.Error())
		return false
	}

	if err := ioutil.WriteFile(dest, data, 0644); err != nil {
		fmt.Println("写入 " + name + " 失败: " + err.Error())
		fmt.Println("解决方法: sudo chown $USER:$USER " + storage.GetConfigDirPath())
		return false
	}

	fmt.Println("  ✓ " + name + " 下载完成")
	return true
}

func copyToOfficial(name string) bool {
	officialDir := "/usr/local/bin"
	if env := os.Getenv("V2RAY_LOCATION_ASSET"); env != "" {
		officialDir = env
	}
	src := storage.GetConfigDirPath() + "/" + name
	dest := officialDir + "/" + name
	data, err := ioutil.ReadFile(src)
	if err != nil {
		fmt.Println("读取 " + src + " 失败: " + err.Error())
		return false
	}
	if err := ioutil.WriteFile(dest, data, 0644); err != nil {
		fmt.Println("复制 " + name + " 失败: " + err.Error())
		return false
	}
	fmt.Println("  ✓ " + name + " 已复制到 " + dest)
	return true
}

func handleRule(ruleType string, args []string) {
	if len(args) < 1 {
		listRuleType(ruleType)
		return
	}
	switch args[0] {
	case "add", "a":
		if len(args) < 2 {
			fmt.Println("用法: -rule", ruleType, "add {域名}")
			return
		}
		addDomain(ruleType, args[1])
	case "remove", "r":
		if len(args) < 2 {
			fmt.Println("用法: -rule", ruleType, "remove {域名}")
			return
		}
		removeDomain(ruleType, args[1])
	default:
		log.E("未知命令:", args[0])
	}
}

func addDomain(ruleType, domain string) {
	domain = normalizeDomain(domain)
	if domain == "" {
		return
	}
	list := getRuleList(ruleType)
	for _, d := range *list {
		if d == domain {
			fmt.Println("域名已存在:", domain)
			return
		}
	}
	*list = append(*list, domain)
	conf.SaveConfig()
	fmt.Println("已添加:", domain)
}

func removeDomain(ruleType, domain string) {
	domain = normalizeDomain(domain)
	list := getRuleList(ruleType)
	newList := make([]string, 0)
	found := false
	for _, d := range *list {
		if d == domain {
			found = true
		} else {
			newList = append(newList, d)
		}
	}
	if !found {
		fmt.Println("域名不存在:", domain)
		return
	}
	*list = newList
	conf.SaveConfig()
	fmt.Println("已移除:", domain)
}

func getRuleList(ruleType string) *[]string {
	if ruleType == "proxy" {
		return &conf.GlobalRule.Proxy
	}
	return &conf.GlobalRule.Direct
}

func listRules() {
	listRuleType("proxy")
	listRuleType("direct")
}

func listRuleType(ruleType string) {
	list := getRuleList(ruleType)
	fmt.Println("=======================================================")
	if ruleType == "proxy" {
		fmt.Println("Proxy 规则:")
	} else {
		fmt.Println("Direct 规则:")
	}
	if len(*list) == 0 {
		fmt.Println("  (空)")
	} else {
		for i, d := range *list {
			fmt.Printf("  [%d] %s\n", i+1, d)
		}
	}
	fmt.Println("=======================================================")
}

func normalizeDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return ""
	}
	if strings.HasPrefix(domain, "http://") {
		domain = domain[7:]
	} else if strings.HasPrefix(domain, "https://") {
		domain = domain[8:]
	}
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}
	domain = strings.TrimPrefix(domain, "www.")
	if strings.Contains(domain, ":") {
		return domain
	}
	return "domain:" + domain
}

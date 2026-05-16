package sub

import (
	"fmt"
	"strings"

	"github.com/Ericwyn/v2sub/ajax"
	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/server"
	"github.com/Ericwyn/v2sub/utils/decode"
	"github.com/Ericwyn/v2sub/utils/log"
	"github.com/Ericwyn/v2sub/utils/putil"
)

func ParseArgs(args []string) {
	if len(args) < 1 {
		return
	}
	switch args[0] {
	case "add":
		if len(args) < 3 {
			fmt.Println("用法: -sub add {name} {url}")
			return
		}
		addSub(args[1], args[2])
	case "update":
		if len(args) < 2 {
			fmt.Println("用法: -sub update {name}")
			return
		}
		if args[1] == "all" {
			for name := range conf.GlobalSer.SubUrls {
				addSub(name, conf.GlobalSer.SubUrls[name])
			}
		} else {
			addSub(args[1], conf.GlobalSer.SubUrls[args[1]])
		}
	case "remove":
		if len(args) < 2 {
			fmt.Println("用法: -sub remove {name}")
			return
		}
		removeSub(args[1])
	case "list":
		listSubs()
	default:
		log.E("sub 参数错误")
	}
}

func addSub(name, url string) {
	log.I("添加订阅:", name)
	if _, exists := conf.GlobalSer.SubUrls[name]; exists {
		// 已存在,删除旧数据
		if name != "" {
			removeSubFromList(name)
		}
	}
	if conf.GlobalSer.SubUrls == nil {
		conf.GlobalSer.SubUrls = make(map[string]string)
	}

	// 请求订阅
	ajax.Send(ajax.Request{
		Url:    url,
		Method: ajax.GET,
		Success: func(resp *ajax.Response) {
			entries := parseSubResult(resp.Body, name)
			if len(entries) == 0 {
				return
			}
			conf.GlobalSer.Sub = append(conf.GlobalSer.Sub, entries...)
			conf.GlobalSer.SubUrls[name] = url
			conf.SaveConfig()
			fmt.Println("订阅成功:", name, "共", len(entries), "个节点")
		},
		Fail: func(status int, msg string) {
			log.E("订阅请求失败:", status, msg)
		},
	})
}

func removeSub(name string) {
	if _, exists := conf.GlobalSer.SubUrls[name]; !exists {
		fmt.Println("订阅不存在:", name)
		return
	}
	removeSubFromList(name)
	delete(conf.GlobalSer.SubUrls, name)
	if conf.GlobalSer.Current.SubName == name {
		conf.GlobalSer.Current.SubName = ""
		conf.GlobalSer.Current.Index = 0
	}
	conf.SaveConfig()
	fmt.Println("已删除订阅:", name)
}

func removeSubFromList(name string) {
	newSub := make([]conf.SerSubEntry, 0)
	for _, entry := range conf.GlobalSer.Sub {
		if entry.SubName != name {
			newSub = append(newSub, entry)
		}
	}
	conf.GlobalSer.Sub = newSub
}

func listSubs() {
	fmt.Println("=======================================================")
	fmt.Println(putil.F("名称", 10), "URL")
	for name, url := range conf.GlobalSer.SubUrls {
		fmt.Println(putil.F(name, 10), url)
	}
	fmt.Println("=======================================================")
}

func parseSubResult(body, subName string) []conf.SerSubEntry {
	decoded := decode.Base64Decode(body)
	if decoded == "" {
		return nil
	}
	entries := make([]conf.SerSubEntry, 0)
	for _, line := range strings.Split(decoded, "\n") {
		if !strings.HasPrefix(line, "vmess://") {
			continue
		}
		vmess := server.ParseVmessLink(line)
		if vmess == nil {
			continue
		}
		entries = append(entries, conf.SerSubEntry{
			SubName: subName,
			Source:  line,
			Vmess:   *vmess,
		})
	}
	return entries
}

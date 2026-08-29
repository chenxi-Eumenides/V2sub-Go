package sub

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Ericwyn/v2sub/ajax"
	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/sub/parser"
	"github.com/Ericwyn/v2sub/utils/log"
	"github.com/Ericwyn/v2sub/utils/putil"
	"github.com/Ericwyn/v2sub/utils/storage"
)

const hooksDir = "hooks"

// SubHook 订阅更新成功后自动运行的脚本
type SubHook struct {
	Script string   // hooks 目录下的脚本文件名
	Args   []string // 用户传入的额外参数
}

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
		addSub(args[1], args[2], nil)
	case "update":
		if len(args) < 2 {
			fmt.Println("用法: -sub update {name} [-hook {文件名} {参数...}]")
			return
		}
		hook := parseHook(args[2:])
		if hook != nil {
			// 订阅更新前先校验 hook 参数, 错误则中止
			if msg := validateHook(hook); msg != "" {
				log.E("hook 参数错误:", msg)
				os.Exit(1)
				return
			}
		}
		if args[1] == "all" {
			for name := range conf.GlobalSer.SubUrls {
				addSub(name, conf.GlobalSer.SubUrls[name], hook)
			}
		} else {
			addSub(args[1], conf.GlobalSer.SubUrls[args[1]], hook)
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

// parseHook 从参数中提取 -hook 标记后的脚本文件名和参数, 未找到则返回 nil
func parseHook(args []string) *SubHook {
	for i, a := range args {
		if a == "-hook" && i+1 < len(args) {
			return &SubHook{
				Script: args[i+1],
				Args:   args[i+2:],
			}
		}
	}
	return nil
}

// validateHook 校验 hook 参数合法性, 返回错误描述; 合法则返回空字符串
func validateHook(hook *SubHook) string {
	if hook.Script == "" {
		return "缺少脚本文件名"
	}
	// 只允许 hooks 目录下的文件名, 拒绝任何包含路径分隔符的输入
	if strings.ContainsAny(hook.Script, `/\`) || hook.Script == "." || hook.Script == ".." {
		return "仅允许 hooks 目录下的文件名, 不支持路径: " + hook.Script
	}
	scriptPath := filepath.Join(hooksPath(), hook.Script)
	if _, err := os.Stat(scriptPath); err != nil {
		return "hooks 目录中不存在脚本: " + hook.Script
	}
	return ""
}

// hooksPath 返回 hooks 目录的完整路径
func hooksPath() string {
	return filepath.Join(storage.GetConfigDirPath(), hooksDir)
}

// runHook 在订阅更新成功后执行 hook 脚本, 订阅名自动作为第一个参数传入
func runHook(hook *SubHook, subName string) {
	if hook == nil || hook.Script == "" {
		return
	}
	if msg := validateHook(hook); msg != "" {
		log.E("hook 参数错误:", msg)
		return
	}
	scriptPath := filepath.Join(hooksPath(), hook.Script)
	fullArgs := append([]string{subName}, hook.Args...)
	cmd := exec.Command(scriptPath, fullArgs...)
	cmd.Dir = storage.GetConfigDirPath()
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		fmt.Println("hook 输出:", string(out))
	}
	if err != nil {
		log.E("hook 脚本执行失败:", err)
	}
}

func addSub(name, url string, hook *SubHook) {
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
			entries := parser.Parse(resp.Body, name)
			if len(entries) == 0 {
				fmt.Println("订阅解析失败: 无法识别的订阅格式")
				return
			}
			conf.GlobalSer.Sub = append(conf.GlobalSer.Sub, entries...)
			conf.GlobalSer.SubUrls[name] = url
			conf.SaveConfig()
			fmt.Println("订阅成功:", name, "共", len(entries), "个节点")
			runHook(hook, name)
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

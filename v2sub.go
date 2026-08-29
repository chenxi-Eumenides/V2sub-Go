package main

import (
	"fmt"
	"os"

	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/conn"
	"github.com/Ericwyn/v2sub/rule"
	"github.com/Ericwyn/v2sub/server"
	"github.com/Ericwyn/v2sub/sub"
	"github.com/Ericwyn/v2sub/utils/log"
)

const versionMsg = "Release 1.4.0"

func main() {
	if len(os.Args) < 2 {
		printArgsHelp()
		os.Exit(0)
	}

	// -v/--version、-h/--help 不检查环境
	switch os.Args[1] {
	case "-v", "--version", "--v":
		fmt.Println("v2sub", versionMsg)
		fmt.Println("https://github.com/chenxi-Eumenides/V2sub-Go")
		return
	case "-h", "--help":
		printArgsHelp()
		return
	}

	conf.InitEnvironment()

	bootArgs := os.Args[1:]
	parseArg(bootArgs)
}

func parseArg(args []string) {
	switch args[0] {
	case "-sub":
		if len(args) < 2 {
			printSubHelp()
			return
		}
		sub.ParseArgs(args[1:])
	case "-status":
		server.ShowInfo()
	case "-ser":
		if len(args) < 2 {
			printSerHelp()
			return
		}
		server.ParseArgs(args[1:])
	case "-rule":
		if len(args) < 2 {
			printRuleHelp()
			return
		}
		rule.ParseArgs(args[1:])
	case "-conf":
		if len(args) < 2 {
			printConfHelp()
			return
		}
		conf.ParseArgs(args[1:])
	case "-conn":
		if len(args) < 2 {
			printConnHelp()
			return
		}
		conn.ParseArgs(args[1:])
	default:
		log.E("参数错误，使用 -h 查看帮助")
		os.Exit(-1)
	}
}

func printArgsHelp() {
	fmt.Println(`v2sub - Linux V2Ray 订阅管理工具
版本 ` + versionMsg + `

用法: v2sub <命令> [子命令] [参数]

  订阅管理 (-sub):
    -sub add {name} {url}         添加订阅
    -sub update {name}            更新订阅
    -sub update {name} -hook {文件名} {参数...}   更新订阅后运行 /etc/v2sub/hooks/ 下的脚本(订阅名自动作为第一个参数)
    -sub update all               更新全部订阅
    -sub remove {name}            删除订阅
    -sub list                     查看订阅列表

  节点管理 (-ser):
    -ser list                     查看所有节点
    -ser set {id}                 设置默认节点
    -ser setx                     测速并设置最快节点
    -ser speedtest                测试节点连接速度

  规则管理 (-rule):
    -rule update                  下载 geosite.dat
    -rule proxy                   查看 Proxy 规则
    -rule proxy add {domain}      添加域名到 Proxy
    -rule proxy remove {domain}   从 Proxy 移除域名
    -rule direct                  查看 Direct 规则
    -rule direct add {domain}     添加域名到 Direct
    -rule direct remove {domain}  从 Direct 移除域名
    -rule list                    查看所有规则

  连接配置 (-conf):
    -conf list                    查看当前配置
    -conf sport {port}            设置 SOCKS 端口 (默认 1080)
    -conf hport {port}            设置 HTTP 端口 (默认 1081)
    -conf lconn {true|false}      允许局域网连接
    -conf bypasslan {true|false}  绕过局域网直连
    -conf copydat {true|false}    下载 geosite.dat 后复制到 V2Ray 默认位置

  连接 (-conn):
    -conn start                   启动 V2Ray
    -conn start-pac               启动 V2Ray + PAC 服务
    -conn kill                    停止 V2Ray

   -status                         查看当前状态信息
   -v, --version                   查看版本号
   -h, --help                      查看帮助`)
}

func printSubHelp() {
	fmt.Println(`用法: v2sub -sub <命令> [参数]

  add {name} {url}      添加订阅
  update {name}         更新指定订阅
  update {name} -hook {文件名} {参数...}   更新订阅后自动运行 /etc/v2sub/hooks/ 下的脚本, 订阅名自动作为第一个参数
  update all            更新全部订阅
  remove {name}         删除指定订阅
  list                  查看所有订阅`)
}

func printSerHelp() {
	fmt.Println(`用法: v2sub -ser <命令> [参数]

  list                  查看所有节点
  set {id}              设置默认节点
  setx                  测速并设置最快节点
  speedtest             测试所有节点连接速度`)
}

func printRuleHelp() {
	fmt.Println(`用法: v2sub -rule <命令> [参数]

  update                下载 geosite.dat
  proxy                 查看 Proxy 规则
  proxy add {domain}    添加域名到 Proxy
  proxy remove {domain} 从 Proxy 移除域名
  direct                查看 Direct 规则
  direct add {domain}   添加域名到 Direct
  direct remove {domain}从 Direct 移除域名
  list                  查看所有规则`)
}

func printConfHelp() {
	fmt.Println(`用法: v2sub -conf <命令> [参数]

  list                  查看当前配置
  sport {port}          设置 SOCKS 端口 (默认 1080)
  hport {port}          设置 HTTP 端口 (默认 1081)
  lconn {true|false}    允许局域网连接
  bypasslan {true|false}绕过局域网直连
  copydat {true|false}  下载 geosite.dat 后复制到 V2Ray 默认位置`)
}

func printConnHelp() {
	fmt.Println(`用法: v2sub -conn <命令> [参数]

  start                 启动 V2Ray 连接
  start-pac             启动 V2Ray + PAC 服务
  kill                  停止 V2Ray`)
}

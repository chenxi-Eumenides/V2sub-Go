package conf

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseArgs(args []string) {
	if len(args) < 1 {
		return
	}
	switch args[0] {
	case "sport":
		if len(args) < 2 {
			fmt.Println("用法: -conf sport {端口号}")
			return
		}
		setInt(args[1], &GlobalConf.SocksPort, "SOCKS 端口")
	case "hport":
		if len(args) < 2 {
			fmt.Println("用法: -conf hport {端口号}")
			return
		}
		setInt(args[1], &GlobalConf.HttpPort, "HTTP 端口")
	case "lconn":
		if len(args) < 2 {
			fmt.Println("用法: -conf lconn {true|false}")
			return
		}
		setBool(args[1], &GlobalConf.AllowLocalConnect, "允许局域网连接")
	case "bypasslan":
		if len(args) < 2 {
			fmt.Println("用法: -conf bypasslan {true|false}")
			return
		}
		setBool(args[1], &GlobalConf.BypassLan, "绕过局域网直连")
	case "copydat":
		if len(args) < 2 {
			fmt.Println("用法: -conf copydat {true|false}")
			return
		}
		setBool(args[1], &GlobalConf.CopyDatToOfficial, "复制 geosite.dat 到官方位置")
	case "list":
		fmt.Println("SOCKS 端口:", GlobalConf.SocksPort)
		fmt.Println("HTTP 端口:", GlobalConf.HttpPort)
		fmt.Println("允许局域网连接:", GlobalConf.AllowLocalConnect)
		fmt.Println("绕过局域网直连:", GlobalConf.BypassLan)
		fmt.Println("复制 geosite.dat 到官方位置:", GlobalConf.CopyDatToOfficial)
	default:
		fmt.Println("未知命令:", args[0])
	}
}

func setInt(val string, target *int, name string) {
	n, err := strconv.Atoi(val)
	if err != nil || n <= 0 || n >= 65534 {
		fmt.Println(name, "设置失败, 无效值:", val)
		return
	}
	*target = n
	fmt.Println("设置", name, "为:", n)
	SaveConfig()
}

func setBool(val string, target *bool, name string) {
	val = strings.ToLower(val)
	var b bool
	switch val {
	case "true", "t", "1":
		b = true
	case "false", "f", "0":
		b = false
	default:
		fmt.Println(name, "设置失败, 请输入 true 或 false")
		return
	}
	*target = b
	fmt.Println("设置", name, "为:", b)
	SaveConfig()
}

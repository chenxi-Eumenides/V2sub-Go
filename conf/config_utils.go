package conf

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Ericwyn/v2sub/utils/log"
	"github.com/Ericwyn/v2sub/utils/storage"
)

// 载入 config
// 写 config
// 创建 config 文件夹
// 底层对接 storage

var initFlag = false
var serverConfigName = "server.json"
var subConfigName = "sub.json"
var ruleConfigName = "rules.json"

// 从本地读取配置文件
func LoadLocalConfig() {
	if !initFlag {
		if err := storage.InitConfigDir(); err != nil {
			fmt.Println("=======================================================")
			fmt.Println("错误: 无法访问配置目录 /etc/v2sub/")
			fmt.Println("")
			fmt.Println("v2sub 需要在 " + storage.GetConfigDirPath() + " 目录读写配置文件。")
			fmt.Println("当前用户没有该目录的写入权限。")
			fmt.Println("")
			fmt.Println("解决方法: 使用 sudo 运行")
			fmt.Println("  sudo v2sub [参数]")
			fmt.Println("")
			fmt.Println("也可以手动创建目录并更改所有者为当前用户:")
			fmt.Println("  sudo mkdir -p " + storage.GetConfigDirPath())
			fmt.Println("  sudo chown $USER:$USER " + storage.GetConfigDirPath())
			fmt.Println("=======================================================")
			os.Exit(1)
		}
		subConfigBytes := storage.ReadConfigFileLocal(subConfigName)
		if string(subConfigBytes) != "" {
			err := json.Unmarshal(subConfigBytes, &SubConfigNow)
			if err != nil {
				log.E("parse sub config file to json error")
			}
		}

		serverConfigBytes := storage.ReadConfigFileLocal(serverConfigName)
		if string(serverConfigBytes) != "" {
			err := json.Unmarshal(serverConfigBytes, &ServerConfigNow)
			if err != nil {
				log.E("parse server config file to json error")
			}
			if ServerConfigNow.SocksPort == 0 {
				ServerConfigNow.SocksPort = 1080
			}
			if ServerConfigNow.HttpPort == 0 {
				ServerConfigNow.HttpPort = 1081
			}
		}

		if ServerConfigNow.ServerList == nil {
			ServerConfigNow.ServerList = make([]VServer, 0)
		}

		loadRuleConfig()

		initFlag = true
		log.I("load config msg success")
	}
}

// 将配置文件保存到本地
func FlushConfig() {
	writeLocalConfig(SubConfigNow, ServerConfigNow)
}

// 将配置输出到本地文件中
func writeLocalConfig(subMap map[string]VSub, serverList ServerConfig) {
	mapJson, err := json.MarshalIndent(subMap, "", "    ")
	if err != nil {
		log.E("general sub map json error")
		panic(err)
	} else {
		storage.WriteConfigFileLocal(string(mapJson), subConfigName)
	}

	serversJson, err := json.MarshalIndent(serverList, "", "    ")
	if err != nil {
		log.E("general sub map json error")
		panic(err)
	} else {
		storage.WriteConfigFileLocal(string(serversJson), serverConfigName)
	}
}

func loadRuleConfig() {
	ruleConfigBytes := storage.ReadConfigFileLocal(ruleConfigName)
	if string(ruleConfigBytes) != "" {
		err := json.Unmarshal(ruleConfigBytes, &RuleConfigNow)
		if err != nil {
			log.E("parse rule config file to json error")
		}
	}
	if RuleConfigNow.Proxy == nil {
		RuleConfigNow.Proxy = make([]string, 0)
	}
	if RuleConfigNow.Direct == nil {
		RuleConfigNow.Direct = make([]string, 0)
	}
}

func FlushRuleConfig() {
	ruleJson, err := json.MarshalIndent(RuleConfigNow, "", "    ")
	if err != nil {
		log.E("general rule config json error")
		panic(err)
	} else {
		storage.WriteConfigFileLocal(string(ruleJson), ruleConfigName)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/conn"
	"github.com/Ericwyn/v2sub/utils/log"
	"github.com/Ericwyn/v2sub/utils/storage"
)

const geoSiteURL = "https://raw.githubusercontent.com/Loyalsoldier/v2ray-rules-dat/release/geosite.dat"

func runInit() {
	fmt.Println("======================================")
	fmt.Println("  v2sub 初始化")
	fmt.Println("======================================")
	fmt.Println()

	initConfigDir()
	initDetectV2ray()
	initCreateConfigs()
	initDownloadGeosite()

	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("  初始化完成")
	fmt.Println("  使用 v2sub -sub add 添加订阅")
	fmt.Println("  使用 v2sub -conn start 启动代理")
	fmt.Println("======================================")
}

func initConfigDir() {
	fmt.Print("[1/4] 创建配置目录 /etc/v2sub/ ... ")
	if err := storage.InitConfigDir(); err != nil {
		fmt.Println("✗ 失败")
		fmt.Printf("  原因: %v\n", err)
		fmt.Println("  请手动执行: sudo mkdir -p /etc/v2sub && sudo chown $USER:$USER /etc/v2sub")
	} else {
		fmt.Println("✓ 成功")
	}
}

func initDetectV2ray() {
	fmt.Print("[2/4] 查找 v2ray ... ")
	path, err := conn.FindV2ray()
	if err != nil {
		fmt.Println("✗ 未找到")
		fmt.Printf("  原因: %v\n", err)
		fmt.Println("  请安装 v2ray (https://www.v2fly.org)，或设置环境变量:")
		fmt.Println("    export V2RAY_PATH=/你的/v2ray/路径")
	} else {
		fmt.Printf("✓ 找到 %s\n", path)
	}
}

func initCreateConfigs() {
	fmt.Println("[3/4] 创建默认配置文件 ...")
	allOk := true

	// config_module.json — 使用 storage 内置模板
	if ok := createConfigModuleFile(); ok {
		fmt.Println("  ✓ config_module.json")
	} else {
		fmt.Println("  ✗ config_module.json 创建失败")
		fmt.Println("    请手动将 config_module.json 放到 /etc/v2sub/")
		allOk = false
	}

	if ok := createFileIfNotExists("server.json", defaultServerConfig()); ok {
		fmt.Println("  ✓ server.json (已存在则跳过)")
	} else {
		fmt.Println("  ✗ server.json 创建失败")
		allOk = false
	}

	if ok := createFileIfNotExists("sub.json", defaultSubConfig()); ok {
		fmt.Println("  ✓ sub.json (已存在则跳过)")
	} else {
		fmt.Println("  ✗ sub.json 创建失败")
		allOk = false
	}

	if ok := createFileIfNotExists("rules.json", defaultRuleConfig()); ok {
		fmt.Println("  ✓ rules.json (已存在则跳过)")
	} else {
		fmt.Println("  ✗ rules.json 创建失败")
		allOk = false
	}

	if !allOk {
		fmt.Println("  部分文件创建失败，请检查目录权限")
		fmt.Println("  sudo chown $USER:$USER /etc/v2sub")
	}
}

func initDownloadGeosite() {
	fmt.Print("[4/4] 下载 geosite.dat ... ")
	fmt.Println()
	fmt.Println("  正在从 Loyalsoldier/v2ray-rules-dat 下载 ...")

	resp, err := http.Get(geoSiteURL)
	if err != nil {
		fmt.Printf("  ✗ 下载失败: %v\n", err)
		fmt.Println("  解决方法:")
		fmt.Println("    1. 检查网络连接是否能访问 GitHub")
		fmt.Println("    2. 手动执行: v2sub -rule update")
		fmt.Println("    3. 手动下载: wget -O /etc/v2sub/geosite.dat " + geoSiteURL)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("  ✗ 下载失败: HTTP %d\n", resp.StatusCode)
		fmt.Println("  解决方法: 手动执行 v2sub -rule update")
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("  ✗ 读取数据失败: %v\n", err)
		return
	}

	if err := ioutil.WriteFile("/etc/v2sub/geosite.dat", data, 0644); err != nil {
		fmt.Printf("  ✗ 写入文件失败: %v\n", err)
		fmt.Println("  解决方法: 检查 /etc/v2sub/ 目录写入权限")
		return
	}

	fmt.Println("  ✓ geosite.dat 下载完成")
}

func createConfigModuleFile() bool {
	configDir := storage.GetConfigDirPath()
	filePath := configDir + "/config_module.json"
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	storage.LoadV2ConfigModule()
	return true
}

func createFileIfNotExists(fileName string, jsonData interface{}) bool {
	configDir := storage.GetConfigDirPath()
	filePath := configDir + "/" + fileName
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	bytes, err := json.MarshalIndent(jsonData, "", "    ")
	if err != nil {
		log.E("marshal default config error: ", err.Error())
		return false
	}
	storage.WriteConfigFileLocal(string(bytes), fileName)
	return true
}

func defaultServerConfig() conf.ServerConfig {
	return conf.ServerConfig{
		SocksPort:         1080,
		HttpPort:          1081,
		AllowLocalConnect: false,
		BypassLan:         false,
		ServerList:        []conf.VServer{},
	}
}

func defaultSubConfig() map[string]conf.VSub {
	return map[string]conf.VSub{}
}

func defaultRuleConfig() conf.DomainRule {
	return conf.DomainRule{
		Proxy:  []string{},
		Direct: []string{},
	}
}

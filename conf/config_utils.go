package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/Ericwyn/v2sub/utils/storage"
)

func InitEnvironment() {
	initConfigDir()
	initV2SubConfig()
	initConfigModule()
}

func initConfigDir() {
	configDir := storage.GetConfigDirPath()
	if _, err := os.Stat(configDir); err == nil {
		return
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Println("无法创建配置目录 " + configDir)
		fmt.Println("原因:", err)
		fmt.Println("请手动执行:")
		fmt.Println("  sudo mkdir -p " + configDir)
		fmt.Println("  sudo chown $USER:$USER " + configDir)
		fmt.Println("设置后可以无需 root 权限运行 v2sub")
		os.Exit(1)
	}

	// 创建成功，尝试改为当前用户权限
	chownToUser(configDir)
	fmt.Println("已创建配置目录 " + configDir)
	fmt.Println("  （目录权限已设置为当前用户，后续无需 root 权限）")
}

func chownToUser(path string) {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser == "" {
		return
	}
	u, err := user.Lookup(sudoUser)
	if err != nil {
		return
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	os.Chown(path, uid, gid)
}

func initV2SubConfig() {
	v2subPath := storage.GetConfigDirPath() + "/v2sub.json"
	if _, err := os.Stat(v2subPath); os.IsNotExist(err) {
		cfg := defaultV2SubJson()
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Println("创建默认配置失败:", err)
			os.Exit(1)
		}
		storage.SaveV2SubConfig(string(data))
		chownToUser(v2subPath)
		fmt.Println("已创建默认配置:", v2subPath)
	}

	raw := storage.LoadV2SubConfig()
	if raw == "" {
		fmt.Println("读取 v2sub.json 失败: 文件为空")
		os.Exit(1)
	}
	var full V2SubJson
	if err := json.Unmarshal([]byte(raw), &full); err != nil {
		fmt.Println("解析 v2sub.json 失败:", err)
		os.Exit(1)
	}
	GlobalConf = full.Conf
	GlobalRule = full.Rule
	GlobalSer = full.Ser
	if GlobalConf.SocksPort == 0 {
		GlobalConf.SocksPort = 1080
	}
	if GlobalConf.HttpPort == 0 {
		GlobalConf.HttpPort = 1081
	}
}

func initConfigModule() {
	modulePath := storage.GetConfigDirPath() + "/config_module.json"
	needsChown := false
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		needsChown = true
	}
	storage.LoadV2ConfigModule()
	if needsChown {
		chownToUser(modulePath)
	}
}

func SaveConfig() {
	data, err := json.MarshalIndent(map[string]interface{}{
		"conf": GlobalConf,
		"rule": GlobalRule,
		"ser":  GlobalSer,
	}, "", "  ")
	if err != nil {
		fmt.Println("序列化配置失败:", err)
		os.Exit(1)
	}
	storage.SaveV2SubConfig(string(data))
	chownToUser(storage.GetConfigDirPath() + "/v2sub.json")
}

func defaultV2SubJson() map[string]interface{} {
	return map[string]interface{}{
		"conf": map[string]interface{}{
			"socksPort":         1080,
			"httpPort":          1081,
			"allowLocalConnect": false,
			"bypassLan":         false,
			"copyDatToOfficial": false,
		},
		"rule": map[string]interface{}{
			"proxy":  []string{},
			"direct": []string{},
		},
		"ser": map[string]interface{}{
			"current": map[string]interface{}{
				"subName": "",
				"index":   0,
			},
			"sub":     []interface{}{},
			"subUrls": map[string]interface{}{},
		},
	}
}

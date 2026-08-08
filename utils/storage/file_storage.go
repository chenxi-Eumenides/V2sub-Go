package storage

import (
	"fmt"
	"os"

	"github.com/Ericwyn/GoTools/file"
	"github.com/Ericwyn/v2sub/utils/log"
)

// 本地配置文件 IO 管理

const configDirPath = "/etc/v2sub"

func InitConfigDir() error {
	return os.MkdirAll(configDirPath, 0755)
}

// 写文件
func WriteConfigFileLocal(data string, fileName string) {
	_ = os.MkdirAll(configDirPath, 0755)

	dir := file.OpenFile(configDirPath)
	configFile := file.OpenFile(dir.AbsPath() + "/" + fileName)

	// 直接删除，再重写，解决写文件错乱的问题
	if configFile.Exits() {
		configFile.Delete()
	}

	err := configFile.Write(file.W_NEW, []string{data})
	if err != nil {
		log.E("write config to " + configFile.AbsPath() + " error")
		panic(err)
	}
}

// 读文件
func ReadConfigFileLocal(fileName string) []byte {
	_ = os.MkdirAll(configDirPath, 0755)
	dir := file.OpenFile(configDirPath)
	configFile := file.OpenFile(dir.AbsPath() + "/" + fileName)
	read, err := configFile.Read()
	if err != nil {
		log.E("read config file: " + configFile.AbsPath() + " error")
	}
	return read
}

var moduleFileName string = "config_module.json"
var v2subConfigName string = "v2sub.json"

func LoadV2ConfigModule() string {
	_ = os.MkdirAll(configDirPath, 0755)
	dir := file.OpenFile(configDirPath)
	configFile := file.OpenFile(dir.AbsPath() + "/" + moduleFileName)
	read, err := configFile.Read()
	if err == nil && string(read) != "" {
		return string(read)
	} else {
		createV2ConfigModule()
		fmt.Println("已创建默认模板:", configDirPath+"/"+moduleFileName)
		return module
	}
}

func GetConfigDirPath() string {
	return configDirPath
}

func SaveV2SubConfig(data string) {
	WriteConfigFileLocal(data, v2subConfigName)
}

func LoadV2SubConfig() string {
	raw := ReadConfigFileLocal(v2subConfigName)
	return string(raw)
}

func createV2ConfigModule() {
	_ = os.MkdirAll(configDirPath, 0755)
	dir := file.OpenFile(configDirPath)
	configFile := file.OpenFile(dir.AbsPath() + "/" + moduleFileName)
	err := configFile.Write(file.W_NEW, []string{module})
	if err != nil {
		log.E("write config to " + configFile.AbsPath() + " error")
		panic(err)
	}
}

var module = `{
  "log": {
    "access": "",
    "error": "",
    "loglevel": "warning"
  },
  "inbounds": [
    {
      "tag": "socks",
      "port": {sPort},
      "listen": "{bindAddr}",
      // "listen": "127.0.0.1",
      "protocol": "socks",
      "sniffing": {
        "enabled": true,
        "destOverride": [
          "http",
          "tls"
        ]
      },
      "settings": {
        "auth": "noauth",
        "udp": true,
        "allowTransparent": false
      }
    },
    {
      "tag": "http",
      "port": {hPort},
      "listen": "{bindAddr}",
      // "listen": "127.0.0.1",
      "protocol": "http",
      "sniffing": {
        "enabled": true,
        "destOverride": [
          "http",
          "tls"
        ]
      },
      "settings": {
        "udp": false,
        "allowTransparent": false
      }
    }
  ],
  "outbounds": [
    {ProxyOutbound},
    {
      "tag": "direct",
      "protocol": "freedom",
      "settings": {}
    },
    {
      "tag": "block",
      "protocol": "blackhole",
      "settings": {
        "response": {
          "type": "http"
        }
      }
    }
  ],
  "dns": {
    "servers": [
      {
        "address": "223.5.5.5", //中国大陆域名使用阿里的 DNS
        "port": 53,
        "domains": [
          "geosite:cn",
          "ntp.org"   // NTP 服务器
        ]
      },
      {
        "address": "114.114.114.114", //中国大陆域名使用 114 的 DNS (备用)
        "port": 53,
        "domains": [
          "geosite:cn",
          "ntp.org"   // NTP 服务器
        ]
      },
      {
        "address": "8.8.8.8", //非中国大陆域名使用 Google 的 DNS
        "port": 53,
        "domains": [
          "geosite:geolocation-!cn"
        ]
      },
      {
        "address": "1.1.1.1", //非中国大陆域名使用 Cloudflare 的 DNS
        "port": 53,
        "domains": [
          "geosite:geolocation-!cn"
        ]
      }
    ]
  },
  "routing": {
    "domainStrategy": "IPOnDemand",
    "rules": [
      {
        "type": "field", 
        "ip": [ 
          // 设置 DNS 配置中的国内 DNS 服务器地址直连，以达到 DNS 分流目的
          "223.5.5.5",
          "114.114.114.114"
        ],
        "outboundTag": "direct"
      },
      {
        "type": "field",
        "ip": [ 
          // 设置 DNS 配置中的国外 DNS 服务器地址走代理，以达到 DNS 分流目的
          "8.8.8.8",
          "1.1.1.1"
        ],
        "outboundTag": "proxy"
      },
      { // 广告拦截
        "type": "field", 
        "domain": [
          "geosite:category-ads-all"
        ],
        "outboundTag": "block"
      },
      { // BT 流量直连
        "type": "field",
        "protocol":["bittorrent"], 
        "outboundTag": "direct"
      },
      { // 直连中国大陆主流网站 ip 和 保留 ip
        "type": "field", 
        "ip": [
          "geoip:private",
          "geoip:cn"
        ],
        "outboundTag": "direct"
      },
      { // 直连中国大陆主流网站域名
        "type": "field", 
        "domain": [
          "geosite:cn"
        ],
        "outboundTag": "direct"
      },
      { // 自定义 proxy 域名规则
        "type": "field",
        "domain": [
          "{customProxyDomains}"
        ],
        "outboundTag": "proxy"
      },
      { // 自定义 direct 域名规则
        "type": "field",
        "domain": [
          "{customDirectDomains}"
        ],
        "outboundTag": "direct"
      }{bypassLanRule}
    ]
  }
}
`

package conn

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Ericwyn/GoTools/file"
	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/utils/command"
	"github.com/Ericwyn/v2sub/utils/log"
	"github.com/Ericwyn/v2sub/utils/param"
	"github.com/Ericwyn/v2sub/utils/putil"
	"github.com/Ericwyn/v2sub/utils/storage"
)

var pacFilePath = "/etc/v2sub/v2sub.pac"

var defaultPacText = `
// 默认全部直连模式
function FindProxyForURL(url, host) {
   return 'DIRECT';
}
`

func ParseArgs(args []string) {
	param.AssistParamLength(args, 1)
	switch args[0] {
	case "start":
		startV2ray()
		fmt.Println("v2ray 已停止")
	case "kill":
		KillV2Sub()
	case "start-pac":
		readPacConfigFile()
		go startPacServerOnly()
		startV2ray()
	default:
		log.E("conn 参数错误")
	}
}

func startV2ray() {
	log.I("start v2ray ......")
	checkV2ray()

	// Guard: check if there are any nodes
	if len(conf.GlobalSer.Sub) == 0 {
		log.E("没有可用节点，请先使用 -sub add 添加订阅")
		os.Exit(-1)
	}
	// Guard: check if Current.Index is valid
	if conf.GlobalSer.Current.Index >= len(conf.GlobalSer.Sub) {
		log.E("当前节点索引无效 (", conf.GlobalSer.Current.Index, ">=", len(conf.GlobalSer.Sub), ")，已重置为 0")
		conf.GlobalSer.Current.Index = 0
		conf.GlobalSer.Current.SubName = conf.GlobalSer.Sub[0].SubName
		conf.SaveConfig()
	}

	runConfig := conf.GlobalSer.Sub[conf.GlobalSer.Current.Index]

	generateConfig(runConfig)

	log.I("use config is :   ↓ ↓ ↓ ↓ ↓ ↓ ↓ ↓ ↓ ↓ ")
	log.I("========================================================================")
	log.I(putil.F("ID", 4), putil.F("别名", 50), putil.F("地址", 24), putil.F("端口", 10), putil.F("类型", 5))
	log.I(putil.F(" ["+strconv.Itoa(conf.GlobalSer.Current.Index)+"]", 4),
		putil.F(runConfig.Vmess.GetPs(), 50),
		putil.F(runConfig.Vmess.Addr, 24),
		putil.F(runConfig.Vmess.GetPort(), 10),
		putil.F(runConfig.ProtocolName(), 5))
	log.I("========================================================================")

	configPath := storage.GetConfigDirPath() + "/config.json"
	log.I("v2ray config path : " + configPath)
	fmt.Println()
	fmt.Println()

	v2rayBin, _ := GetV2rayBinPath()
	var err error
	if useNewV2rayVersion() {
		err = command.RunSync(v2rayBin, "run", "-c", configPath)
	} else {
		err = command.RunSync(v2rayBin, "-config", configPath)
	}
	if err != nil {
		log.E("start v2ray error...")
		log.E(err.Error())
		os.Exit(-1)
	}
}

func generateConfig(entry conf.SerSubEntry) {
	module := storage.LoadV2ConfigModule()

	module = strings.Replace(module, "{ProxyOutbound}", buildProxyOutbound(entry), 1)

	module = strings.Replace(module, "{sPort}", strconv.Itoa(conf.GlobalConf.SocksPort), 1)
	module = strings.Replace(module, "{hPort}", strconv.Itoa(conf.GlobalConf.HttpPort), 1)
	bindAddr := "127.0.0.1"
	if conf.GlobalConf.AllowLocalConnect {
		bindAddr = "0.0.0.0"
	}
	module = strings.Replace(module, "{bindAddr}", bindAddr, -1)

	proxyDomains := buildDomainList(conf.GlobalRule.Proxy)
	directDomains := buildDomainList(conf.GlobalRule.Direct)
	module = strings.Replace(module, "{customProxyDomains}", proxyDomains, 1)
	module = strings.Replace(module, "{customDirectDomains}", directDomains, 1)

	module = strings.Replace(module, "{bypassLanRule}", buildBypassLanRule(), 1)

	storage.WriteConfigFileLocal(module, "config.json")
}

// buildProxyOutbound builds the "proxy" outbound JSON for the entry's protocol.
func buildProxyOutbound(entry conf.SerSubEntry) string {
	switch entry.ProtocolName() {
	case "ss":
		return buildSsOutbound(entry)
	case "vless":
		return buildVlessOutbound(entry)
	default:
		return buildVmessOutbound(entry)
	}
}

func buildVmessOutbound(entry conf.SerSubEntry) string {
	v := entry.Vmess
	return `{
      "tag": "proxy",
      "protocol": "vmess",
      "settings": {
        "vnext": [
          {
            "address": ` + jsonStr(v.Addr) + `,
            "port": ` + v.GetPort() + `,
            "users": [
              {
                "id": ` + jsonStr(v.ID) + `,
                "alterId": ` + strconv.Itoa(v.Aid) + `,
                "email": "t@t.tt",
                "security": "auto"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": ` + jsonStr(v.Net) + `
      },
      "mux": {
        "enabled": false,
        "concurrency": -1
      }
    }`
}

func buildSsOutbound(entry conf.SerSubEntry) string {
	v := entry.Vmess
	return `{
      "tag": "proxy",
      "protocol": "shadowsocks",
      "settings": {
        "servers": [
          {
            "address": ` + jsonStr(v.Addr) + `,
            "port": ` + v.GetPort() + `,
            "method": ` + jsonStr(v.Method) + `,
            "password": ` + jsonStr(v.Password) + `,
            "email": "t@t.tt"
          }
        ]
      }
    }`
}

func buildVlessOutbound(entry conf.SerSubEntry) string {
	v := entry.Vmess
	security := v.Security
	if security == "" {
		security = "none"
	}
	out := `{
      "tag": "proxy",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": ` + jsonStr(v.Addr) + `,
            "port": ` + v.GetPort() + `,
            "users": [
              {
                "id": ` + jsonStr(v.ID) + `,
                "encryption": "none",`
	if v.Flow != "" {
		out += `
                "flow": ` + jsonStr(v.Flow) + `,`
	}
	out += `
                "email": "t@t.tt"
              }
            ]
          }
        ]
      },
      "streamSettings": {
        "network": ` + jsonStr(v.Net) + `,
        "security": ` + jsonStr(security)
	switch security {
	case "reality":
		out += `,
        "realitySettings": {
          "serverName": ` + jsonStr(v.Sni) + `,
          "fingerprint": ` + jsonStr(v.Fp) + `,
          "publicKey": ` + jsonStr(v.Pbk)
		if v.Sid != "" {
			out += `,
          "shortId": ` + jsonStr(v.Sid)
		}
		if v.Spx != "" {
			out += `,
          "spiderX": ` + jsonStr(v.Spx)
		}
		out += `
        }`
	case "tls":
		out += `,
        "tlsSettings": {
          "serverName": ` + jsonStr(v.Sni) + `,
          "fingerprint": ` + jsonStr(v.Fp) + `,
          "allowInsecure": false
        }`
	}
	out += `
      }
    }`
	return out
}

// jsonStr quotes a string for safe embedding in JSON.
func jsonStr(s string) string {
	return strconv.Quote(s)
}

func buildDomainList(domains []string) string {
	if len(domains) == 0 {
		return ""
	}
	parts := make([]string, len(domains))
	for i, d := range domains {
		parts[i] = `"` + d + `"`
	}
	return strings.Join(parts, ", ")
}

func buildBypassLanRule() string {
	if !conf.GlobalConf.BypassLan {
		return ""
	}
	return `,
      {
        "type": "field",
        "ip": [
          "geoip:private",
          "geoip:cn",
          "127.0.0.0/8",
          "10.0.0.0/8",
          "172.16.0.0/12",
          "192.168.0.0/16"
        ],
        "outboundTag": "direct"
      }`
}

func checkV2ray() {
	path, err := FindV2ray()
	if err != nil {
		log.E("can't find v2ray: ", err.Error())
		log.E("please install v2ray or set V2RAY_PATH environment variable")
		os.Exit(-1)
	}
	log.I("found v2ray at: ", path)
}

func readPacConfigFile() {
	pacFile := file.OpenFile(pacFilePath)
	if pacFile.Exits() {
		read, err := pacFile.Read()
		if err != nil {
			log.E("read pac config error, use default pac config")
		} else {
			defaultPacText = string(read)
		}
	} else {
		log.E("read pac config error, pacFile in '" + pacFilePath + "' not exits, use default pac config")
	}
}

func startPacServerOnly() {
	http.HandleFunc("/v2sub.pac", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, pacFilePath)
	})
	s := &http.Server{
		Addr:           ":23333",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	_ = s.ListenAndServe()
}

func KillV2Sub() {
	grep := exec.Command("grep", "v2")
	ps := exec.Command("ps", "cax")
	pipe, _ := ps.StdoutPipe()
	defer pipe.Close()
	grep.Stdin = pipe
	ps.Start()
	res, _ := grep.Output()
	for _, line := range strings.Split(string(res), "\n") {
		elem := strings.Split(line, " ")
		if len(elem) < 2 {
			continue
		}
		pid := elem[0]
		name := elem[len(elem)-1]
		if name == "v2ray" || name == "v2sub" {
			_ = command.RunSync("kill", pid)
		}
	}
}

func useNewV2rayVersion() bool {
	v2rayBin, _ := GetV2rayBinPath()
	result, err := command.RunResult(v2rayBin + " -version")
	if err != nil {
		log.I("check new v2ray version, 'v2ray -version' get error msg, use new v2ray version cmd")
		return true
	}
	log.I("check new v2ray version, 'v2ray -version' get result, ", result)
	return false
}

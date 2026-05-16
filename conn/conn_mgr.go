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

	runConfig := conf.GlobalSer.Sub[conf.GlobalSer.Current.Index]

	generateConfig(runConfig)

	log.I("use config is :   ↓ ↓ ↓ ↓ ↓ ↓ ↓ ↓ ↓ ↓ ")
	log.I("========================================================================")
	log.I(putil.F("ID", 4), putil.F("别名", 50), putil.F("地址", 24), putil.F("端口", 10), putil.F("类型", 5))
	log.I(putil.F(" ["+strconv.Itoa(conf.GlobalSer.Current.Index)+"]", 4),
		putil.F(runConfig.Vmess.Ps, 50),
		putil.F(runConfig.Vmess.Addr, 24),
		putil.F(runConfig.Vmess.Port, 10),
		putil.F(runConfig.Vmess.Type, 5))
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

	module = strings.Replace(module, "{Add}", entry.Vmess.Addr, 1)
	module = strings.Replace(module, "{Port}", entry.Vmess.Port, 1)
	module = strings.Replace(module, "{ID}", entry.Vmess.ID, 1)
	module = strings.Replace(module, "{Aid}", strconv.Itoa(entry.Vmess.Aid), 1)
	module = strings.Replace(module, "{Net}", entry.Vmess.Net, 1)

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

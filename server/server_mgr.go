package server

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Ericwyn/v2sub/conf"
	"github.com/Ericwyn/v2sub/utils/command"
	"github.com/Ericwyn/v2sub/utils/decode"
	"github.com/Ericwyn/v2sub/utils/log"
	"github.com/Ericwyn/v2sub/utils/putil"
)

func ParseArgs(args []string) {
	if len(args) < 1 {
		return
	}
	switch args[0] {
	case "set":
		if len(args) < 2 {
			fmt.Println("用法: -ser set {ID}")
			return
		}
		setServer(args[1])
	case "setx":
		SpeedTestAll(true)
	case "speedtest":
		SpeedTestAll(false)
	case "list":
		listServer()
	default:
		log.E("ser 参数错误")
	}
}

func setServer(idStr string) {
	index, err := strconv.Atoi(idStr)
	if err != nil || index < 0 || index >= len(conf.GlobalSer.Sub) {
		fmt.Println("无效 ID:", idStr)
		return
	}
	conf.GlobalSer.Current.Index = index
	conf.GlobalSer.Current.SubName = conf.GlobalSer.Sub[index].SubName
	conf.SaveConfig()
	fmt.Println("已设置默认节点:", index)
}

func listServer() {
	if len(conf.GlobalSer.Sub) == 0 {
		fmt.Println("没有可用节点")
		fmt.Println("请先使用 -sub add 添加订阅")
		return
	}
	fmt.Println("=======================================================")
	fmt.Println(
		putil.F("ID", 4),
		putil.F("别名", 50),
		putil.F("地址", 24),
		putil.F("端口", 10),
		putil.F("类型", 5),
	)
	for i, entry := range conf.GlobalSer.Sub {
		mark := " "
		if i == conf.GlobalSer.Current.Index && entry.SubName == conf.GlobalSer.Current.SubName {
			mark = "["
		}
		fmt.Println(
			putil.F(mark+strconv.Itoa(i)+"]", 4),
			putil.F(entry.Vmess.Ps, 50),
			putil.F(entry.Vmess.Addr, 24),
			putil.F(entry.Vmess.Port, 10),
			putil.F(entry.Vmess.Type, 5),
		)
	}
	fmt.Println("=======================================================")
}

type SpeedSortEntry struct {
	Speed float64
	Index int
	Entry conf.SerSubEntry
}

func SpeedTestAll(setFastest bool) {
	if len(conf.GlobalSer.Sub) == 0 {
		log.E("没有可用节点进行测速")
		return
	}

	fmt.Println("=======================================================")
	fmt.Println(putil.F("ID", 4), putil.F("别名", 50), putil.F("地址", 24), putil.F("端口", 10), putil.F("类型", 5), putil.F("测速", 5))

	results := sortBySpeed(conf.GlobalSer.Sub)
	if len(results) == 0 {
		log.E("测速失败")
		return
	}

	for _, r := range results {
		mark := " "
		if r.Index == conf.GlobalSer.Current.Index && r.Entry.SubName == conf.GlobalSer.Current.SubName {
			mark = "["
		}
		fmt.Println(
			putil.F(mark+strconv.Itoa(r.Index)+"]", 4),
			putil.F(r.Entry.Vmess.Ps, 50),
			putil.F(r.Entry.Vmess.Addr, 24),
			putil.F(r.Entry.Vmess.Port, 10),
			putil.F(r.Entry.Vmess.Net, 5),
			putil.F(fmt.Sprint(r.Speed)+" ms", 5),
		)
	}
	fmt.Println("=======================================================")

	if setFastest && len(results) > 0 {
		best := results[0]
		conf.GlobalSer.Current.Index = best.Index
		conf.GlobalSer.Current.SubName = best.Entry.SubName
		conf.SaveConfig()
		fmt.Println("已设置最快节点为:", best.Entry.Vmess.Ps)
	}
}

func sortBySpeed(list []conf.SerSubEntry) []SpeedSortEntry {
	var wg sync.WaitGroup
	results := make([]SpeedSortEntry, 0)
	for i, entry := range list {
		wg.Add(1)
		i, entry := i, entry
		go func() {
			ms := pingMs(entry.Vmess.Addr, 8)
			results = append(results, SpeedSortEntry{Speed: ms, Index: i, Entry: entry})
			wg.Done()
		}()
	}
	wg.Wait()
	sort.Slice(results, func(i, j int) bool {
		return results[i].Speed-results[j].Speed < 0
	})
	return results
}

func pingMs(host string, timeout int) float64 {
	out, err := command.RunResult(fmt.Sprintf("timeout %d ping -c 3 %s | grep '^rtt' | awk -F\"/\" '{print $5F}'", timeout, host))
	if err != nil {
		return 9999
	}
	out = strings.TrimSpace(out)
	ms, err := strconv.ParseFloat(out, 64)
	if err != nil {
		return 9999
	}
	return ms
}

func ParseVmessLink(vmessStr string) *conf.VmessJson {
	if !strings.HasPrefix(vmessStr, "vmess://") {
		return nil
	}
	b64 := vmessStr[8:]
	jsonStr := decode.VmessBase64Decode(b64)
	if jsonStr == "" {
		return nil
	}
	var vmess conf.VmessJson
	if err := json.Unmarshal([]byte(jsonStr), &vmess); err != nil {
		log.E("parse vmess json fail")
		return nil
	}
	return &vmess
}

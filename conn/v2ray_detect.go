package conn

import (
	"fmt"
	"os"
	"os/exec"
)

// v2rayBinPath 缓存找到的 v2ray 二进制路径
var v2rayBinPath string

// FindV2ray 自动查找 v2ray 二进制文件位置
// 查找顺序：V2RAY_PATH 环境变量 → PATH → 常见安装路径
// 返回绝对路径，或错误原因
func FindV2ray() (string, error) {
	if v2rayBinPath != "" {
		return v2rayBinPath, nil
	}

	// 1. 优先使用 V2RAY_PATH 环境变量
	if envPath := os.Getenv("V2RAY_PATH"); envPath != "" {
		if info, err := os.Stat(envPath); err == nil && !info.IsDir() {
			v2rayBinPath = envPath
			return envPath, nil
		}
	}

	// 2. 在 PATH 环境变量中搜索
	if path, err := exec.LookPath("v2ray"); err == nil {
		v2rayBinPath = path
		return path, nil
	}

	// 3. 回退到常见安装路径
	fallbackPaths := []string{
		"/usr/local/bin/v2ray",
		"/usr/bin/v2ray",
	}

	for _, fp := range fallbackPaths {
		if info, err := os.Stat(fp); err == nil && !info.IsDir() {
			v2rayBinPath = fp
			return fp, nil
		}
	}

	return "", fmt.Errorf("v2ray not found in PATH, V2RAY_PATH env, or common locations (/usr/local/bin/v2ray, /usr/bin/v2ray)")
}

// GetV2rayBinPath 获取已缓存的 v2ray 路径
// 如果之前未查找过，返回空字符串
func GetV2rayBinPath() (string, error) {
	if v2rayBinPath != "" {
		return v2rayBinPath, nil
	}
	return FindV2ray()
}

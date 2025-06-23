package wifi

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// SavedWiFi 表示一个已保存的WiFi网络
type SavedWiFi struct {
	SSID     string
	Password string
}

// GetSavedNetworks 获取已保存的WiFi网络
func GetSavedNetworks() ([]SavedWiFi, error) {
	// 获取所有保存的WiFi配置文件
	cmd := exec.Command("netsh", "wlan", "show", "profiles")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("获取WiFi配置文件失败: %v", err)
	}

	// 解析输出，提取SSID
	ssids := extractSSIDs(string(output))

	// 获取每个SSID的密码
	var savedNetworks []SavedWiFi
	for _, ssid := range ssids {
		password, err := getPasswordForSSID(ssid)
		if err != nil {
			// 如果获取密码失败，记录错误但继续处理其他网络
			fmt.Printf("警告: 获取 %s 的密码失败: %v\n", ssid, err)
			savedNetworks = append(savedNetworks, SavedWiFi{
				SSID:     ssid,
				Password: "获取失败",
			})
		} else {
			savedNetworks = append(savedNetworks, SavedWiFi{
				SSID:     ssid,
				Password: password,
			})
		}
	}

	return savedNetworks, nil
}

// extractSSIDs 从netsh输出中提取SSID
func extractSSIDs(output string) []string {
	var ssids []string

	// 使用正则表达式匹配"所有用户配置文件 : SSID名称"这一行
	re := regexp.MustCompile(`(?i)所有用户配置文件\s*:\s*(.+)`)
	if !re.MatchString(output) {
		// 尝试英文版本
		re = regexp.MustCompile(`(?i)All User Profile\s*:\s*(.+)`)
	}

	matches := re.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			ssid := strings.TrimSpace(match[1])
			ssids = append(ssids, ssid)
		}
	}

	return ssids
}

// getPasswordForSSID 获取指定SSID的密码
func getPasswordForSSID(ssid string) (string, error) {
	// 使用netsh命令获取指定SSID的详细信息，包括密码
	cmd := exec.Command("netsh", "wlan", "show", "profile", "name="+ssid, "key=clear")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("获取WiFi密码失败: %v", err)
	}

	// 解析输出，提取密码
	return extractPassword(string(output)), nil
}

// extractPassword 从netsh输出中提取密码
func extractPassword(output string) string {
	// 使用正则表达式匹配"关键内容 : 密码"这一行
	re := regexp.MustCompile(`(?i)关键内容\s*:\s*(.+)`)
	if !re.MatchString(output) {
		// 尝试英文版本
		re = regexp.MustCompile(`(?i)Key Content\s*:\s*(.+)`)
	}

	match := re.FindStringSubmatch(output)
	if len(match) >= 2 {
		return strings.TrimSpace(match[1])
	}

	return "未找到密码"
}

// FormatSavedNetworksResult 格式化已保存网络的结果
func FormatSavedNetworksResult(networks []SavedWiFi) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("=== 已保存的WiFi网络 - %s ===\n\n", time.Now().Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("发现 %d 个已保存的WiFi网络:\n\n", len(networks)))

	for i, network := range networks {
		result.WriteString(fmt.Sprintf("网络 #%d:\n", i+1))
		result.WriteString(fmt.Sprintf("  SSID: %s\n", network.SSID))
		result.WriteString(fmt.Sprintf("  密码: %s\n", network.Password))
		result.WriteString("\n")
	}

	return result.String()
}

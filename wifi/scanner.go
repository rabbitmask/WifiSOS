package wifi

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// WiFiNetwork 表示一个WiFi网络
type WiFiNetwork struct {
	SSID     string // 网络名称
	BSSID    string // MAC地址
	Signal   string // 信号强度
	Channel  string // 信道
	Security string // 安全类型
}

// String 返回WiFiNetwork的字符串表示
func (w WiFiNetwork) String() string {
	return fmt.Sprintf("SSID: %-20s | 信号: %-10s | 信道: %-8s | 安全类型: %-15s | BSSID: %s",
		w.SSID,
		w.Signal,
		w.Channel,
		w.Security,
		w.BSSID)
}

// IsValid 检查网络信息是否有效
func (w WiFiNetwork) IsValid() bool {
	return w.SSID != "" && w.BSSID != "" &&
		w.Signal != "" && w.Channel != "" && w.Security != ""
}

// FormatSignal 格式化信号强度显示
func (w WiFiNetwork) FormatSignal() string {
	// 移除百分号并尝试解析数字
	signal := strings.TrimSuffix(w.Signal, "%")
	if signalNum, err := strconv.Atoi(signal); err == nil {
		// 将信号强度转换为星号表示
		stars := signalNum / 20 // 每20%一颗星
		return strings.Repeat("*", stars) + fmt.Sprintf(" (%s)", w.Signal)
	}
	return w.Signal
}

// ScanNetworks 扫描附近的WiFi网络
func ScanNetworks() ([]WiFiNetwork, error) {
	// 首先刷新网络列表
	refreshCmd := exec.Command("netsh", "wlan", "show", "networks", "refresh")
	_, err := refreshCmd.CombinedOutput()
	if err != nil {
		//fmt.Printf("刷新网络列表时出错: %v，继续扫描...\n", err)
	}

	// 等待一小段时间让刷新完成
	time.Sleep(500 * time.Millisecond)

	// 使用netsh命令扫描WiFi网络，确保获取所有详细信息
	cmd := exec.Command("netsh", "wlan", "show", "networks", "mode=Bssid")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果失败，尝试不带mode=Bssid参数重试
		fmt.Printf("使用mode=Bssid扫描失败，尝试基本扫描...\n")
		cmd = exec.Command("netsh", "wlan", "show", "networks")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("扫描WiFi网络失败: %v", err)
		}
	}

	// 将输出转换为字符串并检查是否为空
	outputStr := string(output)
	if len(strings.TrimSpace(outputStr)) == 0 {
		return nil, fmt.Errorf("未获取到WiFi网络信息")
	}

	// 打印原始输出以便调试
	//fmt.Println("=== netsh命令原始输出 ===")
	//fmt.Println(outputStr)
	//fmt.Println("========================")

	// 解析输出
	networks := parseNetshOutput(outputStr)

	// 验证解析结果
	if len(networks) == 0 {
		fmt.Println("警告: 未解析到任何网络信息")
	}

	return networks, nil
}

// parseNetshOutput 解析netsh命令的输出
func parseNetshOutput(output string) []WiFiNetwork {
	var networks []WiFiNetwork
	var currentNetwork WiFiNetwork
	var inBssidSection bool = false

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行
		if line == "" {
			continue
		}

		// 检查是否是新的SSID行
		if strings.HasPrefix(line, "SSID") && !strings.Contains(line, "BSSID") {
			// 如果已经有一个网络在处理中，添加到列表
			if currentNetwork.SSID != "" && currentNetwork.BSSID != "" {
				networks = append(networks, currentNetwork)
			}

			// 开始新的网络
			parts := strings.SplitN(line, ":", 2)
			if len(parts) >= 2 {
				currentNetwork = WiFiNetwork{
					SSID: strings.TrimSpace(parts[1]),
				}
			}
			inBssidSection = false
		} else if strings.Contains(line, "BSSID") && strings.Contains(line, ":") {
			// 新的BSSID部分开始
			if currentNetwork.BSSID != "" && currentNetwork.SSID != "" {
				// 保存之前的BSSID信息作为单独的网络
				networks = append(networks, currentNetwork)
				// 创建新的网络条目，保留相同的SSID
				ssid := currentNetwork.SSID
				currentNetwork = WiFiNetwork{SSID: ssid}
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) >= 2 {
				currentNetwork.BSSID = strings.TrimSpace(parts[1])
			}
			inBssidSection = true
		} else if inBssidSection {
			// 在BSSID部分内处理其他属性
			lowerLine := strings.ToLower(line)

			// 处理安全类型 - 检查多种可能的关键词
			if strings.Contains(lowerLine, "authentication") ||
				strings.Contains(lowerLine, "验证") ||
				strings.Contains(lowerLine, "安全类型") ||
				strings.Contains(lowerLine, "security") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) >= 2 {
					currentNetwork.Security = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(lowerLine, "signal") ||
				strings.Contains(lowerLine, "信号") ||
				strings.Contains(lowerLine, "强度") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) >= 2 {
					currentNetwork.Signal = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(lowerLine, "channel") ||
				strings.Contains(lowerLine, "信道") ||
				strings.Contains(lowerLine, "频道") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) >= 2 {
					currentNetwork.Channel = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// 添加最后一个网络
	if currentNetwork.SSID != "" && currentNetwork.BSSID != "" {
		networks = append(networks, currentNetwork)
	}

	// 调试输出
	fmt.Printf("解析到 %d 个WiFi网络\n", len(networks))
	for i, net := range networks {
		fmt.Printf("网络 #%d: SSID=%s, BSSID=%s, Signal=%s, Channel=%s, Security=%s\n",
			i+1, net.SSID, net.BSSID, net.Signal, net.Channel, net.Security)
	}

	return networks
}

// FormatNetworksResult 格式化网络扫描结果
func FormatNetworksResult(networks []WiFiNetwork) string {
	var result strings.Builder

	// 添加标题
	result.WriteString(fmt.Sprintf("=== WiFi扫描结果 - %s ===\n\n",
		time.Now().Format("2006-01-02 15:04:05")))

	if len(networks) == 0 {
		result.WriteString("未发现WiFi网络\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("发现 %d 个WiFi网络:\n\n",
		len(networks)))

	// 计算SSID的最大长度，用于对齐显示
	maxSSIDLen := 20
	for _, network := range networks {
		if len(network.SSID) > maxSSIDLen {
			maxSSIDLen = len(network.SSID)
		}
	}

	// 添加表头
	result.WriteString(fmt.Sprintf("%-4s | %-*s | %-15s | %-10s | %-20s | %s\n",
		"序号",
		maxSSIDLen, "SSID",
		"信号强度",
		"信道",
		"安全类型",
		"BSSID"))

	// 添加分隔线
	separatorLen := 4 + 3 + maxSSIDLen + 3 + 15 + 3 + 10 + 3 + 20 + 3 + 17
	result.WriteString(strings.Repeat("-", separatorLen) + "\n")

	// 添加网络信息
	for i, network := range networks {
		// 处理空值
		signal := network.Signal
		if signal == "" {
			signal = "N/A"
		}

		channel := network.Channel
		if channel == "" {
			channel = "N/A"
		}

		security := network.Security
		if security == "" {
			security = "N/A"
		}

		bssid := network.BSSID
		if bssid == "" {
			bssid = "N/A"
		}

		// 格式化信号强度显示
		signalDisplay := "N/A"
		if signal != "N/A" {
			signalDisplay = network.FormatSignal()
		}

		// 添加网络信息行
		result.WriteString(fmt.Sprintf("%-4d | %-*s | %-15s | %-10s | %-20s | %s\n",
			i+1,
			maxSSIDLen, network.SSID,
			signalDisplay,
			channel,
			security,
			bssid))
	}

	// 添加详细信息
	result.WriteString("\n详细信息:\n\n")
	for i, network := range networks {
		result.WriteString(fmt.Sprintf("网络 #%d:\n", i+1))
		result.WriteString(fmt.Sprintf("  SSID: %s\n", network.SSID))
		result.WriteString(fmt.Sprintf("  BSSID: %s\n", network.BSSID))
		result.WriteString(fmt.Sprintf("  信号强度: %s\n", network.Signal))
		result.WriteString(fmt.Sprintf("  信道: %s\n", network.Channel))
		result.WriteString(fmt.Sprintf("  安全类型: %s\n", network.Security))
		result.WriteString("\n")
	}

	// 添加注释说明
	result.WriteString("\n注意:\n")
	result.WriteString("- 信号强度: * = 20%, ***** = 100%\n")
	result.WriteString("- N/A 表示信息不可用\n")
	result.WriteString("- 某些字段可能因系统限制或权限不足而无法显示\n")

	return result.String()
}

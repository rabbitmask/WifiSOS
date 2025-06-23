package main

import (
	"WifiSOS/utils"
	"WifiSOS/wifi"
	"fmt"
	"github.com/akamensky/argparse"
	"os"
	"strconv"
)

func main() {
	// 创建命令行解析器
	parser := argparse.NewParser("WifiSOS", "WiFi扫描、密码获取和爆破工具")

	// 定义命令
	scanCommand := parser.NewCommand("scan", "扫描附近的WiFi网络")
	savedCommand := parser.NewCommand("saved", "获取已保存的WiFi网络及密码")
	bruteCommand := parser.NewCommand("brute", "对指定WiFi进行密码爆破")

	// 爆破命令的参数
	ssid := bruteCommand.String("s", "ssid", &argparse.Options{
		Required: true,
		Help:     "目标WiFi的SSID",
	})
	dictPath := bruteCommand.String("d", "dict", &argparse.Options{
		Required: false,
		Help:     "自定义密码字典文件路径",
	})
	maxAttempts := bruteCommand.String("m", "max", &argparse.Options{
		Required: false,
		Help:     "最大尝试次数",
		Default:  "0",
	})

	// 解析命令行参数
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	// 根据命令执行相应的功能
	if scanCommand.Happened() {
		scanWiFi()
	} else if savedCommand.Happened() {
		getSavedWiFi()
	} else if bruteCommand.Happened() {
		// 将最大尝试次数转换为整数
		max, err := strconv.Atoi(*maxAttempts)
		if err != nil {
			fmt.Println("错误: 最大尝试次数必须是一个整数")
			return
		}
		bruteForceWiFi(*ssid, *dictPath, max)
	} else {
		// 如果没有指定命令，显示帮助信息
		fmt.Print(parser.Usage("请指定一个命令: scan, saved 或 brute"))
	}
}

// scanWiFi 扫描附近的WiFi网络
func scanWiFi() {
	fmt.Println("正在扫描附近的WiFi网络...")
	
	// 执行扫描
	networks, err := wifi.ScanNetworks()

	if err != nil {
		fmt.Println(fmt.Sprintf("扫描失败: %v", err))
		return
	}

	// 格式化并显示结果
	result := wifi.FormatNetworksResult(networks)
	fmt.Println(result)

	// 保存结果
	filename, err := utils.SaveResult("wifi_scan", result)
	if err != nil {
		fmt.Println(fmt.Sprintf("保存结果失败: %v", err))
	} else {
		fmt.Println(fmt.Sprintf("结果已保存到: %s", filename))
	}
}

// getSavedWiFi 获取已保存的WiFi网络及密码
func getSavedWiFi() {
	fmt.Println("正在获取已保存的WiFi网络及密码...")

	networks, err := wifi.GetSavedNetworks()
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		return
	}

	// 格式化并显示结果
	result := wifi.FormatSavedNetworksResult(networks)
	fmt.Println(result)

	// 保存结果
	filename, err := utils.SaveResult("saved_wifi", result)
	if err != nil {
		fmt.Printf("保存结果失败: %v\n", err)
	} else {
		fmt.Printf("结果已保存到: %s\n", filename)
	}
}

// bruteForceWiFi 对指定WiFi进行密码爆破
func bruteForceWiFi(ssid string, dictPath string, maxAttempts int) {
	fmt.Printf("正在对WiFi '%s' 进行密码爆破...\n", ssid)

	if dictPath != "" {
		fmt.Printf("使用自定义密码字典: %s\n", dictPath)
	} else {
		fmt.Println("使用内置密码字典")
	}

	if maxAttempts > 0 {
		fmt.Printf("最大尝试次数: %d\n", maxAttempts)
	}

	result, err := wifi.BruteForceWiFi(ssid, dictPath, maxAttempts)
	if err != nil {
		fmt.Printf("爆破失败: %v\n", err)
		return
	}

	// 格式化并显示结果
	formattedResult := wifi.FormatBruteForceResult(result)
	fmt.Println(formattedResult)

	// 保存结果
	filename, err := utils.SaveResult("brute_force_"+ssid, formattedResult)
	if err != nil {
		fmt.Printf("保存结果失败: %v\n", err)
	} else {
		fmt.Printf("结果已保存到: %s\n", filename)
	}
}

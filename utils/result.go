package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SaveResult 将结果保存到文件
func SaveResult(prefix string, content string) (string, error) {

	// 生成文件名，格式为: prefix_timestamp.txt
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(fmt.Sprintf("%s_%s.txt", prefix, timestamp))

	// 写入文件
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("写入文件失败: %v", err)
	}

	return filename, nil
}

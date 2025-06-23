# WifiSOS

WifiSOS是一个用Go语言编写的WiFi扫描、密码获取和爆破工具。

## 功能

- 扫描附近的WiFi网络
- 获取已保存的WiFi网络及密码
- 对指定WiFi进行密码爆破

## 安装

1. 确保已安装Go环境（推荐Go 1.16或更高版本）
2. 克隆本仓库
3. 进入项目目录，运行以下命令安装依赖：

```bash
go mod tidy
```

4. 编译项目：

```bash
go build -o wifigos.exe
```

## 使用方法

### 扫描附近的WiFi网络

```bash
wifigos.exe scan
```

### 获取已保存的WiFi网络及密码

```bash
wifigos.exe saved
```

### 对指定WiFi进行密码爆破

```bash
wifigos.exe brute -s "WiFi名称" [-d 密码字典文件路径] [-m 最大尝试次数]
```

参数说明：
- `-s, --ssid`: 目标WiFi的SSID（必需）
- `-d, --dict`: 自定义密码字典文件路径（可选，默认使用内置密码字典）
- `-m, --max`: 最大尝试次数（可选，默认尝试所有密码）

## 结果保存

所有操作的结果会自动保存在当前目录下，文件名格式为`操作类型_时间戳.txt`。

## 注意事项

1. 本工具仅供网络安全学习和研究使用
2. 请勿用于非法用途，如未经授权入侵他人网络
3. 使用本工具造成的任何后果由使用者自行承担

## 系统要求

- Windows操作系统（Windows 7/8/10/11）
- 管理员权限（部分功能需要）
- 支持WiFi的网卡

## 许可证

MIT License

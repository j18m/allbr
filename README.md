# Allbr - 多功能暴力破解工具

## 版本
v0.0.1

## 简介
Allbr是一个功能强大的多功能暴力破解工具，支持多种服务类型的密码破解和主机存活检测。该工具采用Go语言开发，具有高性能、跨平台和易于使用的特点。

## 功能特性

### 1. 暴力破解功能
- **支持多种服务类型**：SSH、MySQL、FTP、RDP、LDAP、Oracle、MongoDB、Redis
- **灵活的目标指定**：支持单个IP、多个IP（逗号分隔）和IP段（CIDR格式）
- **多种扫描策略**：用户名优先或密码优先
- **智能检测**：支持先ping检测存活主机，再进行端口开放检测
- **多线程支持**：可配置线程数提高扫描效率
- **实时结果保存**：扫描结果实时保存到文件

### 2. 主机存活检测
- **快速检测**：高效的主机存活检测算法
- **多线程支持**：可配置线程数加速检测
- **支持多种目标格式**：单个IP、多个IP和IP段
- **结果导出**：自动保存存活主机列表

## 安装

### 1. 编译安装
```bash
go build -o allbr ./cmd/multibrute/main.go
```

### 2. 直接使用
下载已编译的可执行文件即可使用。

## 快速开始

### 命令结构
```bash
allbr [command] [flags]
```

### 可用命令
- `br`：暴力破解模式
- `ping`：主机存活检测模式

## 暴力破解示例

### 1. 单个目标SSH暴力破解
```bash
./allbr br -t 192.168.1.100 -s ssh -u root -w passwords.txt -T 10
```

### 2. 多个目标MySQL暴力破解（带ping和端口检测）
```bash
./allbr br -t 192.168.1.100,192.168.1.101 -s mysql -u admin,root -w passwords.txt --ping-first --port-check
```

### 3. IP段Redis暴力破解（使用密码优先策略）
```bash
./allbr br -t 192.168.1.0/24 -s redis -u default -w passwords.txt --strategy pass-first -T 20
```

## 主机存活检测示例

### 1. 单个IP检测
```bash
./allbr ping -t 192.168.1.1
```

### 2. IP段检测（指定线程数）
```bash
./allbr ping -t 192.168.1.0/24 -n 20
```

### 3. 多个IP检测
```bash
./allbr ping -t 192.168.1.100,192.168.1.101,192.168.1.102 -n 5
```

## 参数说明

### 通用参数
| 参数 | 简写 | 说明 |
|------|------|------|
| --help | -h | 显示帮助信息 |

### 暴力破解参数（br命令）
| 参数 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| --targets | -t | 目标IP、IP段或文件 | - |
| --service | -s | 服务类型（ssh/mysql/ftp/rdp/ldap/oracle/mongodb/redis） | ssh |
| --port | -p | 服务端口 | 服务默认端口 |
| --usernames | -u | 用户名或用户名文件 | 默认用户名 |
| --passwords | -w | 密码或密码文件 | 默认密码 |
| --threads | -T | 线程数 | 5 |
| --timeout | -o | 超时时间（秒） | 5 |
| --output | -O | 结果输出文件 | multibrute.txt |
| --ping-first | - | 先进行ping检测 | false |
| --port-check | - | 进行端口开放检测 | false |
| --strategy | - | 扫描策略（user-first/pass-first） | user-first |

### 存活检测参数（ping命令）
| 参数 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| --targets | -t | 目标IP、IP段或文件 | - |
| --threads | -n | 线程数 | 10 |

## 支持的服务类型

| 服务类型 | 默认端口 |
|----------|----------|
| SSH | 22 |
| MySQL | 3306 |
| FTP | 21 |
| RDP | 3389 |
| LDAP | 389 |
| Oracle | 1521 |
| MongoDB | 27017 |
| Redis | 6379 |

## 注意事项

1. 本工具仅用于合法的安全测试和授权的渗透测试，请勿用于非法用途。
2. 使用本工具造成的任何后果，由使用者自行承担。
3. 建议在测试前获得目标系统的明确授权。
4. 大规模扫描可能会对网络造成影响，请合理设置线程数和扫描范围。

## 许可证

MIT License

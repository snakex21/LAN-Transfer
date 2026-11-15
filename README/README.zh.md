**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#下载和运行)

一款小巧的便携工具，用于在同一局域网内的两台计算机之间直接发送文件和文件夹。无需云端、无需账户，也没有中转服务器：数据通过加密连接直接从一台电脑传到另一台电脑。核心使用 Go 编写，提供 WebView2 窗口界面，以及适用于脚本的命令行版本。

![LAN Transfer 主窗口：已配对的接收方和三个待发送的项目](../docs/screenshots/main.png)

## 目录

- [功能](#功能)
- [界面语言](#界面语言)
- [截图](#截图)
- [下载和运行](#下载和运行)
- [如何配对](#如何配对)
- [发送文件](#发送文件)
- [命令行](#命令行)
- [端口和防火墙](#端口和防火墙)
- [安全说明](#安全说明)
- [环境变量](#环境变量)
- [构建和测试](#构建和测试)
- [许可证](#许可证)

## 功能

- 自动发现运行 LAN Transfer 的其他计算机（UDP 广播，并以限速的 /24 网段扫描作为后备）。
- 使用接收方显示的 6 位代码进行配对。未配对时无法发送任何内容，也无法浏览任何文件夹。
- 可选的无码模式：在接收方的设置中关闭 **要求配对码**（保存在 `data\settings.json`）或使用 `serve --no-code` 启动。此后网络中任何运行 LAN Transfer 的计算机无需配对码即可向其发送，传输仍然加密；期间标题栏显示 **无需配对码**。
- TLS 1.3 加密。每次运行都会生成新的自签名证书；其指纹在配对时被固定，并在每次连接时进行校验。
- 支持文件和整个文件夹，保留文件夹结构，最多可并行发送 4 个文件。
- 可续传中断的传输，并对每个文件进行 SHA-256 校验。
- 未完成的数据保存在隐藏的 `.lantransfer-partial` 文件夹中，因此不完整的文件永远不会以最终名称出现。
- 名称冲突时会变为 `file (1).txt` 或 `Folder (1)`；绝不会覆盖已有文件。
- 拒绝危险的名称（`..`、绝对路径、`file:stream` 形式的备用数据流、`CON`、`NUL`、`COM1` 等保留名称）。接收前会检查可用磁盘空间。
- 支持从资源管理器拖放，提供传输队列，显示每个文件的进度、速度和剩余时间，支持取消、续传，并记录已发送和已接收项目的历史。
- 便携：单个 EXE 文件，设置和历史记录保存在其旁边的 `data` 文件夹中。
- 窗口、窗口标题和命令行输出支持 30 种语言（与本 README 相同），并根据 Windows 显示语言自动选择；阿拉伯语和希伯来语使用从右到左的布局。

## 界面语言

LAN Transfer 支持与本 README 相同的 30 种语言。首次启动时，它使用 Windows 显示语言（Windows 语言列表中第一个受支持的语言）；如果 Windows 使用的语言不在列表中，界面将显示为英语。

要更改语言，请点击右上角的齿轮按钮（**设置**）并选择一种语言。窗口、窗口标题和所有消息会立即切换，所选语言保存在 `data\settings.json` 中。选择 **自动（Windows）** 可重新跟随 Windows 显示语言。数字、大小、速度、日期和复数形式均遵循所选语言的规则；阿拉伯语和希伯来语使用镜像的从右到左布局，而路径、IP 地址、代码和指纹仍保持从左到右。来自另一台计算机的消息（例如“接收方磁盘空间不足”）会以你的语言显示，而不是以接收方的语言显示。

`LANTransferCLI.exe` 使用相同的语言。`LANTRANSFER_LANG`（例如 `LANTRANSFER_LANG=de`）会覆盖已保存的选择，对窗口和命令行均有效。

| English | Deutsch | العربية (RTL) |
|---|---|---|
| ![英语界面的 LAN Transfer](../docs/screenshots/lang-en.png) | ![德语界面的 LAN Transfer](../docs/screenshots/lang-de.png) | ![阿拉伯语界面的 LAN Transfer，从右到左布局](../docs/screenshots/lang-ar.png) |

## 截图

| 配对 | 带进度的发送 |
|---|---|
| ![显示 6 位代码和证书指纹的配对对话框](../docs/screenshots/pairing.png) | ![显示进度、速度和剩余时间的传输队列](../docs/screenshots/sending.png) |
| **取消后续传** | **接收方：历史记录** |
| ![已取消的传输从已接收的偏移量处续传，SHA-256 校验通过](../docs/screenshots/resume.png) | ![历史记录中显示已接收项目的接收方窗口](../docs/screenshots/history.png) |

## 下载和运行

1. 从 [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) 下载 `LANTransfer-2.1.0-windows-amd64.zip`，并使用发布的 `.sha256` 文件进行校验。
2. 在两台计算机上将其解压到任意文件夹。
3. 启动 `LANTransfer.exe`。首次启动时，Windows 会（通过 UAC）请求添加两条防火墙规则的权限；参见[端口和防火墙](#端口和防火墙)。

系统要求：Windows 10 或 11（64 位），并安装 WebView2 运行时（当前版本的 Windows 已自带）。两台计算机都必须运行第 2 版。第 2 版无法与第 1 版通信。

## 如何配对

1. 在两台计算机上启动 LAN Transfer。接收方会在右上角显示其配对代码和证书指纹。
2. 在发送方计算机上，点击 **接收方** 列表中的接收方。如果它没有出现，请使用 **手动连接（IP 地址）** 并输入其 IP 地址。
3. 输入接收方屏幕上显示的 6 位代码，并将对话框中显示的证书指纹与接收方上的指纹进行比对。如果两者不一致，请不要继续。
4. 点击 **配对**。设备随即显示 **✓ 已配对**。

同一台计算机输错 5 次代码后，接收方会阻止其继续尝试。**新配对码** 会生成新代码并取消所有现有配对。接收方每次启动都会生成新的代码和证书，因此接收方重启后需要重新配对。

## 发送文件

1. 与接收方配对。
2. 将文件或文件夹拖到窗口中，使用 **选择文件** / **选择文件夹**，或粘贴路径。
3. 选择目标位置：接收方的默认文件夹（`Downloads\LAN Transfer`）或接收方上另一个已存在的文件夹。
4. 点击 **发送**。每个项目都会显示进度、速度和剩余时间，并且可以取消。已取消或中断的传输可以续传；只会发送缺失的部分。

## 命令行

`LANTransferCLI.exe` 是不带窗口的同一程序，适用于脚本和自动化。

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

代码也可以通过 `LANTRANSFER_CODE` 传入。没有窗口的接收方可以使用来自 `LANTRANSFER_PIN` 的固定代码。再次运行相同的 `send` 命令即可续传中断的传输。

## 端口和防火墙

- UDP 45454：发现。TCP 45455：传输（HTTPS）。
- 首次启动时，应用会（通过 UAC）请求一次，以添加两条入站规则：`LAN Transfer 2 Discovery (UDP)` 和 `LAN Transfer 2 Transfer (TCP)`。
- 这些规则只接受来自本地子网的连接（`remoteip=LocalSubnet`），并且仅在“专用”和“域”网络配置文件上生效（`profile=private,domain`）。在 Windows 标记为“公用”的网络上（咖啡馆、酒店、机场），其他计算机无法连接到 LAN Transfer。
- 如果在家庭网络中看不到接收方，请在 Windows 网络设置中确认该网络已设置为“专用”。
- 第 1 版创建了名为 `LAN Transfer Discovery` 和 `LAN Transfer Transfer` 的规则，适用于所有配置文件，包括“公用”。如果第 2 版发现这些规则，会显示一个横幅，提供替换它们的选项。在你点击 **更新规则** 并确认 Windows 提示之前，不会做任何更改。

## 安全说明

- 通过 UDP 进行的发现未经身份验证；它仅作为提示。信任来自配对：代码加上被固定的证书指纹。
- 在首次配对期间，同一网络中的主动攻击者（中间人）可能会冒充接收方。配对前请比对两个屏幕上的证书指纹。
- 已配对的发送方可以浏览接收方的文件夹，并将文件保存到接收方上任何已存在的文件夹中（自定义目标位置选项）。请只与你信任的计算机配对，并使用 **新配对码** 撤销所有配对。
- 在标记为“公用”的网络上，防火墙规则不会开放这些端口。
- 关闭 **要求配对码** 后，网络中任何运行 LAN Transfer 的人都可以无需配对码向本机发送文件并浏览其文件夹，与版本 1 相同（但已加密）。请仅在可信网络中使用。重新开启配对码会断开未使用配对码连接的发送方。

## 环境变量

这些变量用于测试和自动化；正常使用时无需设置任何一个。

| 变量 | 含义 |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | 其他端口 |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | 接收文件夹、程序数据文件夹 |
| `LANTRANSFER_NAME` | 向其他计算机显示的计算机名称 |
| `LANTRANSFER_LANG` | 界面和命令行的语言（`en`、`de`、`pl` 等）；覆盖已保存的选择 |
| `LANTRANSFER_LOOPBACK=1` | 仅限 127.0.0.1：不广播、不扫描局域网、不创建防火墙规则 |
| `LANTRANSFER_PEER_UDP_PORTS` | 回环模式下其他实例的 UDP 端口 |
| `LANTRANSFER_NO_FIREWALL=1` | 不创建也不检查防火墙规则 |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | 接收方的固定代码、发送方使用的代码 |
| `LANTRANSFER_REQUIRE_CODE=0` | `0` = 此接收方不需要配对码，`1` = 需要配对码；优先于设置 |

## 构建和测试

在 Windows 10/11 上使用 Go 1.23 或更高版本。

```powershell
build.bat
go test ./...
python tools/check_i18n.py
```

`build.bat` 会运行测试，并使用 `VERSION` 中的版本号（默认 `2.1.0`）构建 `LANTransfer.exe`（窗口版）和 `LANTransferCLI.exe`（控制台版）。测试包括通过 127.0.0.1 进行的端到端传输：配对和锁定、拒绝未经身份验证的请求、从 0 B 到 96 MB 的文件、Unicode 文件夹和超过 260 个字符的路径、名称冲突、恶意名称、错误的校验和、取消和续传、磁盘已满以及证书固定。

翻译文件位于 `locales/ui.<code>.json`（波兰语为源语言，英语为后备语言）。`tools/check_i18n.py` 会检查全部 30 个文件是否具有相同的键、占位符和复数形式，并且没有遗留的波兰语；另有一个 Go 测试检查程序使用的每个键在每种语言中都存在。

发布说明：[docs/releases/v2.1.0.md](../docs/releases/v2.1.0.md) · [v2.0.0](../docs/releases/v2.0.0.md)。

## 许可证

MIT，参见 [LICENSE](../LICENSE)。

**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#다운로드-및-실행)

같은 로컬 네트워크에 있는 두 컴퓨터 간에 파일과 폴더를 직접 보내기 위한 작은 포터블 도구입니다. 클라우드도, 계정도, 중계 서버도 필요 없습니다. 데이터는 암호화된 연결을 통해 한 PC에서 다른 PC로 곧바로 전송됩니다. 핵심은 Go로 작성되었으며, WebView2 창과 스크립트용 명령줄 버전을 제공합니다.

![LAN Transfer 메인 창: 페어링된 수신자와 보낼 준비가 된 세 개의 항목](../docs/screenshots/main.png)

## 목차

- [기능](#기능)
- [인터페이스 언어](#인터페이스-언어)
- [스크린샷](#스크린샷)
- [다운로드 및 실행](#다운로드-및-실행)
- [페어링 방법](#페어링-방법)
- [파일 보내기](#파일-보내기)
- [명령줄](#명령줄)
- [포트 및 방화벽](#포트-및-방화벽)
- [보안 참고 사항](#보안-참고-사항)
- [환경 변수](#환경-변수)
- [빌드 및 테스트](#빌드-및-테스트)
- [라이선스](#라이선스)

## 기능

- LAN Transfer를 실행 중인 다른 컴퓨터를 자동으로 검색합니다(UDP 브로드캐스트, 대체 수단으로 속도를 제한한 /24 스캔 사용).
- 수신자에 표시되는 6자리 코드로 페어링합니다. 페어링하지 않으면 아무것도 보낼 수 없고 폴더를 탐색할 수도 없습니다.
- 코드 없는 선택 모드: 수신 측 설정에서 **페어링 코드 요구**을(를) 끄거나(`data\settings.json`에 저장됨) `serve --no-code`로 시작하세요. 그러면 네트워크에서 LAN Transfer를 실행하는 모든 컴퓨터가 코드 없이 보낼 수 있으며 전송은 계속 암호화됩니다. 그동안 머리글에 **코드 없음** 배지가 표시됩니다.
- TLS 1.3 암호화. 실행할 때마다 새 자체 서명 인증서가 생성되며, 그 지문은 페어링 시 고정(pinning)되고 연결할 때마다 확인됩니다.
- 파일과 폴더 전체를 지원하며, 폴더 구조를 유지하고 최대 4개 파일을 병렬로 전송합니다.
- 중단된 전송 이어받기 및 모든 파일에 대한 SHA-256 검사.
- 완료되지 않은 데이터는 숨겨진 `.lantransfer-partial` 폴더에 보관되므로, 잘린 파일이 최종 이름으로 나타나는 일은 없습니다.
- 이름이 충돌하면 `file (1).txt` 또는 `Folder (1)`이 되며, 기존 파일은 절대 덮어쓰지 않습니다.
- 위험한 이름은 거부됩니다(`..`, 절대 경로, `file:stream` 형식의 대체 데이터 스트림, `CON`, `NUL`, `COM1` 같은 예약된 이름). 수신하기 전에 디스크 여유 공간을 확인합니다.
- 탐색기에서 끌어서 놓기, 파일별 진행률·속도·남은 시간을 보여 주는 대기열, 취소, 이어받기, 보내고 받은 항목의 기록.
- 포터블: EXE 파일 하나만 있으면 되며, 설정과 기록은 그 옆의 `data` 폴더에 저장됩니다.
- 창, 창 제목, 명령줄 출력이 30개 언어(이 README와 동일)로 제공되며 Windows 표시 언어에 따라 자동으로 선택됩니다. 아랍어와 히브리어는 오른쪽에서 왼쪽으로 쓰는 레이아웃을 사용합니다.

## 인터페이스 언어

LAN Transfer는 이 README와 같은 30개 언어를 지원합니다. 처음 시작할 때는 Windows 표시 언어(Windows 언어 목록에서 지원되는 첫 번째 언어)를 사용하며, Windows 언어가 목록에 없으면 인터페이스가 영어로 표시됩니다.

언어를 바꾸려면 오른쪽 위 모서리의 톱니바퀴 단추(**설정**)를 클릭하고 언어를 선택하세요. 창, 창 제목, 모든 메시지가 즉시 바뀌며 선택한 언어는 `data\settings.json`에 저장됩니다. **자동(Windows)** 항목을 선택하면 다시 Windows 표시 언어를 따릅니다. 숫자, 크기, 속도, 날짜, 복수형은 선택한 언어의 규칙을 따릅니다. 아랍어와 히브리어는 좌우가 반전된 오른쪽에서 왼쪽 레이아웃을 사용하지만, 경로, IP 주소, 코드, 지문은 왼쪽에서 오른쪽 방향을 유지합니다. 상대 컴퓨터에서 온 메시지(예: “수신자의 디스크 공간이 부족합니다”)도 수신자의 언어가 아니라 사용자의 언어로 표시됩니다.

`LANTransferCLI.exe`도 같은 언어를 사용합니다. `LANTRANSFER_LANG`(예: `LANTRANSFER_LANG=de`)은 창과 명령줄 모두에서 저장된 선택보다 우선합니다.

| English | Deutsch | العربية (RTL) |
|---|---|---|
| ![영어로 표시된 LAN Transfer](../docs/screenshots/lang-en.png) | ![독일어로 표시된 LAN Transfer](../docs/screenshots/lang-de.png) | ![아랍어로 표시된 LAN Transfer, 오른쪽에서 왼쪽 레이아웃](../docs/screenshots/lang-ar.png) |

## 스크린샷

| 페어링 | 진행률과 함께 보내기 |
|---|---|
| ![6자리 코드와 인증서 지문이 표시된 페어링 대화 상자](../docs/screenshots/pairing.png) | ![진행률, 속도, 남은 시간이 표시된 전송 대기열](../docs/screenshots/sending.png) |
| **취소 후 이어받기** | **수신자: 기록** |
| ![취소된 전송이 이미 받은 오프셋부터 재개되고 SHA-256 검증을 통과한 모습](../docs/screenshots/resume.png) | ![받은 항목이 기록에 표시된 수신자 창](../docs/screenshots/history.png) |

## 다운로드 및 실행

1. [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest)에서 `LANTransfer-2.1.0-windows-amd64.zip`을 다운로드하고, 게시된 `.sha256` 파일과 대조하여 확인합니다.
2. 두 컴퓨터 모두에서 원하는 폴더에 압축을 풉니다.
3. `LANTransfer.exe`를 실행합니다. 처음 실행할 때 Windows가 방화벽 규칙 두 개를 추가할 수 있도록 (UAC를 통해) 권한을 요청합니다. [포트 및 방화벽](#포트-및-방화벽)을 참고하세요.

요구 사항: WebView2 런타임이 있는 Windows 10 또는 11(64비트). WebView2는 최신 Windows에 포함되어 있습니다. 두 컴퓨터 모두 버전 2를 실행해야 합니다. 버전 2는 버전 1과 통신할 수 없습니다.

## 페어링 방법

1. 두 컴퓨터 모두에서 LAN Transfer를 실행합니다. 수신자는 오른쪽 위 모서리에 페어링 코드와 인증서 지문을 표시합니다.
2. 보내는 컴퓨터에서 **수신자** 목록의 수신자를 클릭합니다. 목록에 나타나지 않으면 **수동으로 연결(IP 주소)** 항목을 사용하여 해당 IP 주소를 입력합니다.
3. 수신자 화면에 표시된 6자리 코드를 입력하고, 대화 상자에 표시된 인증서 지문을 수신자의 지문과 비교합니다. 서로 다르면 계속 진행하지 마세요.
4. **페어링**을 클릭합니다. 이제 장치에 **✓ 페어링됨**이 표시됩니다.

한 컴퓨터에서 잘못된 코드를 5번 입력하면 수신자는 해당 컴퓨터의 추가 시도를 차단합니다. **새 코드**는 새 코드를 만들고 기존 페어링을 모두 취소합니다. 수신자는 시작할 때마다 새 코드와 인증서를 만들므로, 수신자를 다시 시작한 후에는 다시 페어링해야 합니다.

## 파일 보내기

1. 수신자와 페어링합니다.
2. 파일이나 폴더를 창으로 끌어 놓거나, **파일 선택** / **폴더 선택**을 사용하거나, 경로를 붙여 넣습니다.
3. 대상을 선택합니다. 수신자의 기본 폴더(`Downloads\LAN Transfer`) 또는 수신자에 이미 있는 다른 폴더를 지정할 수 있습니다.
4. **보내기**를 클릭합니다. 각 항목에는 진행률, 속도, 남은 시간이 표시되며 취소할 수 있습니다. 취소되거나 중단된 전송은 이어받을 수 있으며, 누락된 부분만 전송됩니다.

## 명령줄

`LANTransferCLI.exe`는 창이 없는 동일한 프로그램으로, 스크립트와 자동화에 사용합니다.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

코드는 `LANTRANSFER_CODE`로 전달할 수도 있습니다. 창이 없는 수신자는 `LANTRANSFER_PIN`의 고정 코드를 사용할 수 있습니다. 같은 `send` 명령을 다시 실행하면 중단된 전송이 이어집니다.

## 포트 및 방화벽

- UDP 45454: 검색. TCP 45455: 전송(HTTPS).
- 처음 실행할 때 앱은 두 개의 인바운드 규칙 `LAN Transfer 2 Discovery (UDP)` 및 `LAN Transfer 2 Transfer (TCP)`를 추가하기 위해 (UAC를 통해) 한 번 확인을 요청합니다.
- 이 규칙은 로컬 서브넷에서 오는 연결만(`remoteip=LocalSubnet`), 그리고 개인 및 도메인 네트워크 프로필에서만(`profile=private,domain`) 허용합니다. Windows가 공용으로 표시한 네트워크(카페, 호텔, 공항)에서는 다른 컴퓨터가 LAN Transfer에 연결할 수 없습니다.
- 집 네트워크에서 수신자가 보이지 않으면, Windows 네트워크 설정에서 해당 네트워크가 개인으로 설정되어 있는지 확인하세요.
- 버전 1은 공용을 포함한 모든 프로필에 대해 `LAN Transfer Discovery` 및 `LAN Transfer Transfer`라는 이름의 규칙을 만들었습니다. 버전 2가 이 규칙을 발견하면 교체를 제안하는 배너를 표시합니다. **규칙 업데이트**를 클릭하고 Windows 확인 창을 승인하기 전에는 아무것도 변경되지 않습니다.

## 보안 참고 사항

- UDP를 통한 검색은 인증되지 않으며, 단지 힌트일 뿐입니다. 신뢰는 페어링, 즉 코드와 고정된 인증서 지문에서 나옵니다.
- 첫 페어링 중에는 같은 네트워크에 있는 능동적 공격자(중간자)가 수신자로 가장할 수 있습니다. 페어링하기 전에 두 화면의 인증서 지문을 비교하세요.
- 페어링된 발신자는 수신자의 폴더를 탐색하고 그곳의 기존 폴더 어디에나 파일을 저장할 수 있습니다(사용자 지정 대상 옵션). 신뢰하는 컴퓨터와만 페어링하고, 모든 페어링을 취소하려면 **새 코드**를 사용하세요.
- 공용으로 표시된 네트워크에서는 방화벽 규칙이 포트를 열지 않습니다.
- **페어링 코드 요구**이(가) 꺼져 있으면 네트워크에서 LAN Transfer를 실행하는 누구나 코드 없이 이 컴퓨터로 파일을 보내고 폴더를 탐색할 수 있습니다(버전 1과 같지만 암호화됨). 신뢰할 수 있는 네트워크에서만 사용하세요. 코드를 다시 켜면 코드 없이 연결된 발신 측의 연결이 끊어집니다.

## 환경 변수

이 변수들은 테스트와 자동화를 위한 것이며, 일반적인 사용에는 어느 것도 필요하지 않습니다.

| 변수 | 의미 |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | 다른 포트 |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | 수신 폴더, 프로그램 데이터 폴더 |
| `LANTRANSFER_NAME` | 다른 컴퓨터에 표시되는 컴퓨터 이름 |
| `LANTRANSFER_LANG` | 인터페이스 및 CLI 언어(`en`, `de`, `pl` 등). 저장된 선택보다 우선함 |
| `LANTRANSFER_LOOPBACK=1` | 127.0.0.1 전용: 브로드캐스트 없음, LAN 스캔 없음, 방화벽 규칙 없음 |
| `LANTRANSFER_PEER_UDP_PORTS` | 루프백 모드에서 다른 인스턴스의 UDP 포트 |
| `LANTRANSFER_NO_FIREWALL=1` | 방화벽 규칙을 만들거나 확인하지 않음 |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | 수신자의 고정 코드, 발신자가 사용하는 코드 |
| `LANTRANSFER_REQUIRE_CODE=0` | `0` = 이 수신 측은 페어링 코드가 필요 없음, `1` = 코드 필요. 설정보다 우선합니다 |

## 빌드 및 테스트

Windows 10/11에서 Go 1.23 이상.

```powershell
build.bat
go test ./...
python tools/check_i18n.py
```

`build.bat`은 테스트를 실행하고 `VERSION`의 버전(기본값 `2.1.0`)으로 `LANTransfer.exe`(창 버전)와 `LANTransferCLI.exe`(콘솔 버전)를 빌드합니다. 테스트에는 127.0.0.1을 통한 엔드투엔드 전송이 포함됩니다: 페어링과 잠금, 인증되지 않은 요청 거부, 0 B부터 96 MB까지의 파일, 유니코드 폴더와 260자를 넘는 경로, 이름 충돌, 악의적인 이름, 잘못된 체크섬, 취소와 이어받기, 디스크 가득 참, 인증서 고정.

번역은 `locales/ui.<code>.json`에 있습니다(폴란드어가 원본, 영어가 대체 언어). `tools/check_i18n.py`는 30개 파일 모두의 키, 자리 표시자, 복수형이 같고 폴란드어가 남아 있지 않은지 확인하며, Go 테스트는 프로그램이 사용하는 모든 키가 모든 언어에 있는지 확인합니다.

릴리스 노트: [docs/releases/v2.1.0.md](../docs/releases/v2.1.0.md) · [v2.0.0](../docs/releases/v2.0.0.md).

## 라이선스

MIT, [LICENSE](../LICENSE)를 참고하세요.

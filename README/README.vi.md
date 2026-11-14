**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#tải-về-và-chạy)

Một công cụ di động nhỏ gọn để gửi tệp và thư mục trực tiếp giữa hai máy tính trong cùng một mạng cục bộ. Không có đám mây, không cần tài khoản và không có máy chủ chuyển tiếp: dữ liệu đi thẳng từ máy này sang máy kia qua một kết nối được mã hóa. Lõi viết bằng Go, một cửa sổ WebView2 và một phiên bản dòng lệnh dành cho script.

![Cửa sổ chính của LAN Transfer: một máy nhận đã ghép nối và ba mục sẵn sàng để gửi](../docs/screenshots/main.png)

## Mục lục

- [Tính năng](#tính-năng)
- [Ngôn ngữ giao diện](#ngôn-ngữ-giao-diện)
- [Ảnh chụp màn hình](#ảnh-chụp-màn-hình)
- [Tải về và chạy](#tải-về-và-chạy)
- [Cách ghép nối](#cách-ghép-nối)
- [Gửi tệp](#gửi-tệp)
- [Dòng lệnh](#dòng-lệnh)
- [Cổng và tường lửa](#cổng-và-tường-lửa)
- [Lưu ý bảo mật](#lưu-ý-bảo-mật)
- [Biến môi trường](#biến-môi-trường)
- [Biên dịch và kiểm thử](#biên-dịch-và-kiểm-thử)
- [Giấy phép](#giấy-phép)

## Tính năng

- Tự động phát hiện các máy tính khác đang chạy LAN Transfer (UDP broadcast, kèm cơ chế quét dải /24 có giới hạn tốc độ làm phương án dự phòng).
- Ghép nối bằng mã 6 chữ số hiển thị trên máy nhận. Khi chưa ghép nối thì không thể gửi bất cứ thứ gì và không thể duyệt bất kỳ thư mục nào.
- Mã hóa TLS 1.3. Mỗi lần chạy, chương trình tạo một chứng chỉ tự ký mới; dấu vân tay của chứng chỉ được ghim khi ghép nối và được kiểm tra ở mọi kết nối.
- Gửi tệp và cả thư mục, giữ nguyên cấu trúc thư mục, tối đa 4 tệp được gửi song song.
- Tiếp tục các lần truyền bị gián đoạn và kiểm tra SHA-256 cho từng tệp.
- Dữ liệu chưa hoàn tất được lưu trong thư mục ẩn `.lantransfer-partial`, nên một tệp bị cắt dở không bao giờ xuất hiện dưới tên cuối cùng của nó.
- Khi trùng tên, tệp mới sẽ có tên `file (1).txt` hoặc `Folder (1)`; tệp hiện có không bao giờ bị ghi đè.
- Tên nguy hiểm bị từ chối (`..`, đường dẫn tuyệt đối, luồng dữ liệu thay thế `file:stream`, các tên dành riêng như `CON`, `NUL`, `COM1`). Dung lượng đĩa trống được kiểm tra trước khi nhận.
- Kéo và thả từ File Explorer, hàng đợi với tiến độ của từng tệp, tốc độ, thời gian còn lại, hủy, tiếp tục và lịch sử các mục đã gửi và đã nhận.
- Di động: một tệp EXE duy nhất, cài đặt và lịch sử nằm trong thư mục `data` bên cạnh nó.
- Cửa sổ, tiêu đề cửa sổ và đầu ra dòng lệnh có 30 ngôn ngữ (giống như README này), được chọn tự động theo ngôn ngữ hiển thị của Windows; bố cục từ phải sang trái cho tiếng Ả Rập và tiếng Do Thái.

## Ngôn ngữ giao diện

LAN Transfer hỗ trợ cùng 30 ngôn ngữ như README này. Ở lần khởi động đầu tiên, chương trình dùng ngôn ngữ hiển thị của Windows (ngôn ngữ được hỗ trợ đầu tiên trong danh sách ngôn ngữ của Windows); nếu Windows dùng một ngôn ngữ không có trong danh sách, giao diện sẽ hiển thị bằng tiếng Anh.

Để thay đổi, hãy nhấp vào nút bánh răng (**Cài đặt**) ở góc trên bên phải và chọn một ngôn ngữ. Cửa sổ, tiêu đề cửa sổ và mọi thông báo chuyển đổi ngay lập tức, và lựa chọn được lưu trong `data\settings.json`. **Tự động (Windows)** sẽ lại theo ngôn ngữ hiển thị của Windows. Số, kích thước, tốc độ, ngày tháng và dạng số nhiều tuân theo quy tắc của ngôn ngữ đã chọn; tiếng Ả Rập và tiếng Do Thái dùng bố cục từ phải sang trái được lật ngược, còn đường dẫn, địa chỉ IP, mã và dấu vân tay vẫn hiển thị từ trái sang phải. Thông báo từ máy tính kia (ví dụ “không đủ dung lượng trên ổ đĩa của máy nhận”) được hiển thị bằng ngôn ngữ của bạn, không phải ngôn ngữ của máy nhận.

`LANTransferCLI.exe` dùng cùng ngôn ngữ đó. `LANTRANSFER_LANG` (ví dụ `LANTRANSFER_LANG=de`) ghi đè lựa chọn đã lưu cho cả cửa sổ lẫn dòng lệnh.

| English | Deutsch | العربية (RTL) |
|---|---|---|
| ![LAN Transfer bằng tiếng Anh](../docs/screenshots/lang-en.png) | ![LAN Transfer bằng tiếng Đức](../docs/screenshots/lang-de.png) | ![LAN Transfer bằng tiếng Ả Rập, bố cục từ phải sang trái](../docs/screenshots/lang-ar.png) |

## Ảnh chụp màn hình

| Ghép nối | Gửi kèm tiến độ |
|---|---|
| ![Hộp thoại ghép nối với mã 6 chữ số và dấu vân tay chứng chỉ](../docs/screenshots/pairing.png) | ![Hàng đợi truyền với tiến độ, tốc độ và thời gian còn lại](../docs/screenshots/sending.png) |
| **Tiếp tục sau khi hủy** | **Máy nhận: lịch sử** |
| ![Một lần truyền đã hủy được tiếp tục từ vị trí đã nhận, SHA-256 đã được xác minh](../docs/screenshots/resume.png) | ![Cửa sổ máy nhận với các mục đã nhận trong lịch sử](../docs/screenshots/history.png) |

## Tải về và chạy

1. Tải `LANTransfer-2.1.0-windows-amd64.zip` từ [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) và đối chiếu nó với tệp `.sha256` được công bố.
2. Giải nén vào một thư mục bất kỳ trên cả hai máy tính.
3. Chạy `LANTransfer.exe`. Ở lần chạy đầu tiên, Windows sẽ hỏi (UAC) quyền thêm hai quy tắc tường lửa; xem [Cổng và tường lửa](#cổng-và-tường-lửa).

Yêu cầu: Windows 10 hoặc 11 (64-bit) có WebView2 runtime, vốn là một phần của các phiên bản Windows hiện tại. Cả hai máy tính phải chạy phiên bản 2. Phiên bản 2 không thể giao tiếp với phiên bản 1.

## Cách ghép nối

1. Khởi động LAN Transfer trên cả hai máy tính. Máy nhận hiển thị mã ghép nối và dấu vân tay chứng chỉ ở góc trên bên phải.
2. Trên máy gửi, nhấp vào máy nhận trong danh sách **Máy nhận**. Nếu máy nhận không xuất hiện, hãy dùng **Kết nối thủ công (địa chỉ IP)** và nhập địa chỉ IP của nó.
3. Nhập mã 6 chữ số từ màn hình máy nhận và so sánh dấu vân tay chứng chỉ hiển thị trong hộp thoại với dấu vân tay trên máy nhận. Nếu chúng khác nhau, đừng tiếp tục.
4. Nhấp **Ghép nối**. Thiết bị lúc này hiển thị **✓ đã ghép nối**.

Sau 5 lần nhập sai mã từ cùng một máy tính, máy nhận sẽ chặn các lần thử tiếp theo từ máy đó. **Mã mới** tạo một mã mới và hủy tất cả các ghép nối hiện có. Máy nhận tạo mã và chứng chỉ mới mỗi khi khởi động, vì vậy sau khi máy nhận khởi động lại, hãy ghép nối lại.

## Gửi tệp

1. Ghép nối với máy nhận.
2. Kéo tệp hoặc thư mục vào cửa sổ, dùng **Chọn tệp** / **Chọn thư mục**, hoặc dán một đường dẫn.
3. Chọn nơi đến: thư mục mặc định của máy nhận (`Downloads\LAN Transfer`) hoặc một thư mục khác đã tồn tại trên máy nhận.
4. Nhấp **Gửi**. Mỗi mục hiển thị tiến độ, tốc độ, thời gian còn lại và có thể bị hủy. Một lần truyền đã bị hủy hoặc bị gián đoạn có thể được tiếp tục; chỉ phần còn thiếu được gửi.

## Dòng lệnh

`LANTransferCLI.exe` là cùng một chương trình nhưng không có cửa sổ, dành cho script và tự động hóa.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Mã cũng có thể được truyền qua `LANTRANSFER_CODE`. Máy nhận không có cửa sổ có thể dùng một mã cố định từ `LANTRANSFER_PIN`. Chạy lại cùng lệnh `send` sẽ tiếp tục lần truyền bị gián đoạn.

## Cổng và tường lửa

- UDP 45454: phát hiện. TCP 45455: truyền dữ liệu (HTTPS).
- Ở lần chạy đầu tiên, ứng dụng hỏi một lần (UAC) để thêm hai quy tắc chiều vào: `LAN Transfer 2 Discovery (UDP)` và `LAN Transfer 2 Transfer (TCP)`.
- Các quy tắc chỉ chấp nhận kết nối từ mạng con cục bộ (`remoteip=LocalSubnet`) và chỉ trên các cấu hình mạng Riêng tư (Private) và Miền (Domain) (`profile=private,domain`). Trên mạng mà Windows đánh dấu là Công cộng (Public) (quán cà phê, khách sạn, sân bay), các máy tính khác không thể kết nối tới LAN Transfer.
- Nếu không thấy máy nhận trong mạng gia đình, hãy kiểm tra xem mạng đã được đặt là Riêng tư (Private) trong cài đặt mạng của Windows hay chưa.
- Phiên bản 1 đã tạo các quy tắc có tên `LAN Transfer Discovery` và `LAN Transfer Transfer` cho mọi cấu hình, kể cả Công cộng. Nếu phiên bản 2 tìm thấy chúng, nó sẽ hiển thị một biểu ngữ đề nghị thay thế chúng. Không có gì bị thay đổi cho đến khi bạn nhấp **Cập nhật quy tắc** và xác nhận hộp thoại của Windows.

## Lưu ý bảo mật

- Việc phát hiện qua UDP không được xác thực; nó chỉ là một gợi ý. Sự tin cậy đến từ việc ghép nối: mã cùng với dấu vân tay chứng chỉ đã được ghim.
- Trong lần ghép nối đầu tiên, một kẻ tấn công chủ động trên cùng mạng (man in the middle) có thể giả làm máy nhận. Hãy so sánh dấu vân tay chứng chỉ trên cả hai màn hình trước khi ghép nối.
- Một máy gửi đã ghép nối có thể duyệt các thư mục của máy nhận và lưu tệp vào bất kỳ thư mục hiện có nào ở đó (tùy chọn nơi đến tùy chỉnh). Chỉ ghép nối các máy tính bạn tin tưởng, và dùng **Mã mới** để thu hồi tất cả các ghép nối.
- Các quy tắc tường lửa không mở cổng trên các mạng được đánh dấu là Công cộng.

## Biến môi trường

Các biến này dành cho kiểm thử và tự động hóa; khi sử dụng bình thường thì không cần biến nào.

| Biến | Ý nghĩa |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | cổng khác |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | thư mục nhận, thư mục dữ liệu của chương trình |
| `LANTRANSFER_NAME` | tên máy tính hiển thị cho người khác |
| `LANTRANSFER_LANG` | ngôn ngữ giao diện và CLI (`en`, `de`, `pl`, …); ghi đè lựa chọn đã lưu |
| `LANTRANSFER_LOOPBACK=1` | chỉ 127.0.0.1: không broadcast, không quét LAN, không có quy tắc tường lửa |
| `LANTRANSFER_PEER_UDP_PORTS` | cổng UDP của các instance khác trong chế độ loopback |
| `LANTRANSFER_NO_FIREWALL=1` | không tạo và không kiểm tra quy tắc tường lửa |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | mã cố định của máy nhận, mã được máy gửi sử dụng |

## Biên dịch và kiểm thử

Go 1.23 trở lên trên Windows 10/11.

```powershell
build.bat
go test ./...
python tools/check_i18n.py
```

`build.bat` chạy các bài kiểm thử và biên dịch `LANTransfer.exe` (có cửa sổ) và `LANTransferCLI.exe` (console) với số phiên bản lấy từ `VERSION` (mặc định `2.1.0`). Các bài kiểm thử bao gồm truyền đầu-cuối qua 127.0.0.1: ghép nối và khóa sau nhiều lần sai, từ chối yêu cầu chưa xác thực, tệp từ 0 B đến 96 MB, thư mục Unicode và đường dẫn dài hơn 260 ký tự, trùng tên, tên độc hại, checksum sai, hủy và tiếp tục, đĩa đầy và ghim chứng chỉ.

Các bản dịch nằm trong `locales/ui.<code>.json` (tiếng Ba Lan là nguồn, tiếng Anh là ngôn ngữ dự phòng). `tools/check_i18n.py` kiểm tra rằng cả 30 tệp có cùng khóa, phần giữ chỗ và dạng số nhiều, và không còn sót tiếng Ba Lan; một bài kiểm thử Go kiểm tra rằng mọi khóa mà chương trình dùng đều có trong mọi ngôn ngữ.

Ghi chú phát hành: [docs/releases/v2.1.0.md](../docs/releases/v2.1.0.md) · [v2.0.0](../docs/releases/v2.0.0.md).

## Giấy phép

MIT, xem [LICENSE](../LICENSE).

**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#programı-indirme-ve-çalıştırma)

Aynı yerel ağdaki iki bilgisayar arasında dosya ve klasörleri doğrudan göndermek için küçük, taşınabilir bir araç. Bulut yok, hesap yok, aracı sunucu yok: veriler şifreli bir bağlantı üzerinden doğrudan bir bilgisayardan diğerine gider. Go ile yazılmış çekirdek, bir WebView2 penceresi ve betikler için bir komut satırı sürümü.

![LAN Transfer ana penceresi: eşleştirilmiş bir alıcı ve gönderilmeye hazır üç öğe](../docs/screenshots/main.png)

## İçindekiler

- [Özellikler](#özellikler)
- [Ekran görüntüleri](#ekran-görüntüleri)
- [Programı indirme ve çalıştırma](#programı-indirme-ve-çalıştırma)
- [Eşleştirme nasıl yapılır](#eşleştirme-nasıl-yapılır)
- [Dosya gönderme](#dosya-gönderme)
- [Komut satırı](#komut-satırı)
- [Portlar ve güvenlik duvarı](#portlar-ve-güvenlik-duvarı)
- [Güvenlik notları](#güvenlik-notları)
- [Ortam değişkenleri](#ortam-değişkenleri)
- [Derleme ve testler](#derleme-ve-testler)
- [Lisans](#lisans)

## Özellikler

- LAN Transfer çalıştıran diğer bilgisayarların otomatik olarak bulunması (UDP broadcast; yedek olarak hızı sınırlandırılmış bir /24 taraması).
- Alıcıda gösterilen 6 haneli kodla eşleştirme. Eşleştirme olmadan hiçbir şey gönderilemez ve hiçbir klasöre göz atılamaz.
- TLS 1.3 şifreleme. Her çalıştırmada yeni, kendinden imzalı bir sertifika oluşturulur; parmak izi eşleştirme sırasında sabitlenir ve her bağlantıda denetlenir.
- Klasör yapısı korunarak dosyalar ve klasörlerin tamamı; aynı anda en fazla 4 dosya paralel gönderilir.
- Yarıda kalan aktarımların devam ettirilmesi ve her dosyanın SHA-256 ile denetlenmesi.
- Tamamlanmamış veriler gizli bir `.lantransfer-partial` klasöründe tutulur; böylece yarım kalmış bir dosya hiçbir zaman nihai adıyla görünmez.
- Ad çakışmalarında `file (1).txt` veya `Folder (1)` oluşturulur; mevcut dosyaların üzerine asla yazılmaz.
- Tehlikeli adlar reddedilir (`..`, mutlak yollar, `file:stream` alternatif veri akışları, `CON`, `NUL`, `COM1` gibi ayrılmış adlar). Alım öncesinde boş disk alanı denetlenir.
- Gezgin'den sürükle ve bırak; dosya başına ilerleme, hız, kalan süre, iptal ve devam ettirme içeren bir kuyruk ile gönderilen ve alınan öğelerin geçmişi.
- Taşınabilir: tek bir EXE; ayarlar ve geçmiş, yanındaki `data` klasöründe.

Kullanıcı arayüzü şu anda Lehçedir.

## Ekran görüntüleri

| Eşleştirme | İlerlemeyle gönderme |
|---|---|
| ![6 haneli kodu ve sertifika parmak izini gösteren eşleştirme penceresi](../docs/screenshots/pairing.png) | ![İlerleme, hız ve kalan süreyi gösteren aktarım kuyruğu](../docs/screenshots/sending.png) |
| **İptalden sonra devam ettirme** | **Alıcı: geçmiş** |
| ![İptal edilen bir aktarımın alınmış konumdan devam ettirilmesi, SHA-256 ile doğrulanmış](../docs/screenshots/resume.png) | ![Geçmişte alınan öğelerin görüldüğü alıcı penceresi](../docs/screenshots/history.png) |

## Programı indirme ve çalıştırma

1. `LANTransfer-2.0.0-windows-amd64.zip` dosyasını [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) sayfasından indirin ve yayımlanan `.sha256` dosyasıyla doğrulayın.
2. Her iki bilgisayarda da istediğiniz bir klasöre çıkarın.
3. `LANTransfer.exe` dosyasını başlatın. İlk başlatmada Windows, iki güvenlik duvarı kuralı eklemek için (UAC) izin ister; bkz. [Portlar ve güvenlik duvarı](#portlar-ve-güvenlik-duvarı).

Gereksinimler: güncel Windows sürümlerinin bir parçası olan WebView2 çalışma zamanına sahip Windows 10 veya 11 (64 bit). Her iki bilgisayarda da sürüm 2 çalışmalıdır. Sürüm 2, sürüm 1 ile iletişim kuramaz.

## Eşleştirme nasıl yapılır

1. LAN Transfer'i her iki bilgisayarda da başlatın. Alıcı, eşleştirme kodunu ve sertifika parmak izini sağ üst köşede gösterir.
2. Gönderen bilgisayarda **Odbiorca** (Alıcı) listesindeki alıcıya tıklayın. Listede görünmüyorsa **Połącz ręcznie** (Elle bağlan) seçeneğini kullanıp IP adresini girin.
3. Alıcının ekranındaki 6 haneli kodu yazın ve pencerede gösterilen sertifika parmak izini alıcıdakiyle karşılaştırın. Farklıysa devam etmeyin.
4. **Sparuj** (Eşleştir) düğmesine tıklayın. Cihaz artık **sparowano** (eşleştirildi) olarak görünür.

Bir bilgisayardan 5 yanlış kod girildikten sonra alıcı, o bilgisayardan gelen diğer denemeleri engeller. **Nowy kod** (Yeni kod) yeni bir kod oluşturur ve mevcut tüm eşleştirmeleri iptal eder. Alıcı her başlatıldığında yeni bir kod ve sertifika oluşturur; bu nedenle alıcı yeniden başlatıldıktan sonra tekrar eşleştirme yapın.

## Dosya gönderme

1. Alıcıyla eşleştirin.
2. Dosyaları veya klasörleri pencereye sürükleyin, **Wybierz pliki** / **Wybierz folder** (Dosya seç / Klasör seç) düğmelerini kullanın ya da bir yol yapıştırın.
3. Hedefi seçin: alıcının varsayılan klasörü (`Downloads\LAN Transfer`) veya alıcıda bulunan başka bir mevcut klasör.
4. **Wyślij** (Gönder) düğmesine tıklayın. Her öğe ilerlemeyi, hızı ve kalan süreyi gösterir ve iptal edilebilir. İptal edilen veya yarıda kalan bir aktarım devam ettirilebilir; yalnızca eksik kısım gönderilir.

## Komut satırı

`LANTransferCLI.exe`, betikler ve otomasyon için aynı programın penceresiz sürümüdür.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Kod `LANTRANSFER_CODE` ile de verilebilir. Penceresiz bir alıcı, `LANTRANSFER_PIN` içindeki sabit bir kodu kullanabilir. Aynı `send` komutunu yeniden çalıştırmak, yarıda kalan aktarımı devam ettirir.

## Portlar ve güvenlik duvarı

- UDP 45454: bulma. TCP 45455: aktarım (HTTPS).
- İlk başlatmada uygulama, iki gelen bağlantı kuralı eklemek için bir kez (UAC) izin ister: `LAN Transfer 2 Discovery (UDP)` ve `LAN Transfer 2 Transfer (TCP)`.
- Kurallar bağlantıları yalnızca yerel alt ağdan (`remoteip=LocalSubnet`) ve yalnızca Özel ve Etki Alanı ağ profillerinde (`profile=private,domain`) kabul eder. Windows'un Ortak olarak işaretlediği bir ağda (kafe, otel, havalimanı) diğer bilgisayarlar LAN Transfer'e bağlanamaz.
- Alıcı ev ağınızda görünmüyorsa, Windows ağ ayarlarında ağın Özel olarak ayarlandığını kontrol edin.
- Sürüm 1, Ortak dahil tüm profiller için `LAN Transfer Discovery` ve `LAN Transfer Transfer` adlı kurallar oluşturuyordu. Sürüm 2 bunları bulursa, bunların değiştirilmesini öneren bir bildirim şeridi gösterir. **Zaktualizuj reguły** (Kuralları güncelle) düğmesine tıklayıp Windows onay penceresini onaylayana kadar hiçbir şey değiştirilmez.

## Güvenlik notları

- UDP üzerinden bulma kimlik doğrulamalı değildir; yalnızca bir ipucudur. Güven, eşleştirmeden gelir: kod ile sabitlenmiş sertifika parmak izi.
- İlk eşleştirme sırasında aynı ağdaki aktif bir saldırgan (ortadaki adam, man in the middle) kendini alıcı gibi gösterebilir. Eşleştirmeden önce sertifika parmak izini her iki ekranda karşılaştırın.
- Eşleştirilmiş bir gönderici, alıcının klasörlerine göz atabilir ve oradaki mevcut herhangi bir klasöre dosya kaydedebilir (özel hedef seçeneği). Yalnızca güvendiğiniz bilgisayarları eşleştirin ve tüm eşleştirmeleri iptal etmek için **Nowy kod** düğmesini kullanın.
- Güvenlik duvarı kuralları, Ortak olarak işaretlenmiş ağlarda portları açmaz.

## Ortam değişkenleri

Bunlar testler ve otomasyon için tasarlanmıştır; normal kullanımda hiçbirine gerek yoktur.

| Değişken | Anlamı |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | farklı portlar |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | alım klasörü, program veri klasörü |
| `LANTRANSFER_NAME` | diğerlerine gösterilen bilgisayar adı |
| `LANTRANSFER_LOOPBACK=1` | yalnızca 127.0.0.1: broadcast yok, LAN taraması yok, güvenlik duvarı kuralı yok |
| `LANTRANSFER_PEER_UDP_PORTS` | loopback modunda diğer örneklerin UDP portları |
| `LANTRANSFER_NO_FIREWALL=1` | güvenlik duvarı kurallarını oluşturma veya denetleme |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | alıcının sabit kodu, göndericinin kullandığı kod |

## Derleme ve testler

Windows 10/11 üzerinde Go 1.23 veya daha yenisi.

```powershell
build.bat
go test ./...
```

`build.bat` testleri çalıştırır ve `LANTransfer.exe` (pencereli) ile `LANTransferCLI.exe` (konsol) dosyalarını `VERSION` dosyasındaki sürümle (varsayılan `2.0.0`) derler. Testler 127.0.0.1 üzerinden uçtan uca aktarımları kapsar: eşleştirme ve kilitleme, reddedilen kimliği doğrulanmamış istekler, 0 B ile 96 MB arası dosyalar, Unicode klasörler ve 260 karakterden uzun yollar, ad çakışmaları, kötü niyetli adlar, hatalı sağlama toplamı, iptal ve devam ettirme, dolu disk ve sertifika sabitleme.

Sürüm notları: [docs/releases/v2.0.0.md](../docs/releases/v2.0.0.md).

## Lisans

MIT, bkz. [LICENSE](../LICENSE).

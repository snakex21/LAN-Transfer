**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#unduh-dan-jalankan)

Alat portabel kecil untuk mengirim file dan folder secara langsung antara dua komputer dalam jaringan lokal yang sama. Tanpa cloud, tanpa akun, dan tanpa server perantara: data dikirim langsung dari satu PC ke PC lain melalui koneksi terenkripsi. Inti dalam Go, jendela WebView2, dan versi baris perintah untuk skrip.

![Jendela utama LAN Transfer: penerima yang sudah dipasangkan dan tiga item yang siap dikirim](../docs/screenshots/main.png)

## Daftar isi

- [Fitur](#fitur)
- [Bahasa antarmuka](#bahasa-antarmuka)
- [Tangkapan layar](#tangkapan-layar)
- [Unduh dan jalankan](#unduh-dan-jalankan)
- [Cara memasangkan](#cara-memasangkan)
- [Mengirim file](#mengirim-file)
- [Baris perintah](#baris-perintah)
- [Port dan firewall](#port-dan-firewall)
- [Catatan keamanan](#catatan-keamanan)
- [Variabel lingkungan](#variabel-lingkungan)
- [Build dan pengujian](#build-dan-pengujian)
- [Lisensi](#lisensi)

## Fitur

- Penemuan otomatis komputer lain yang menjalankan LAN Transfer (UDP broadcast, dengan pemindaian /24 yang dibatasi kecepatannya sebagai cadangan).
- Pemasangan dengan kode 6 digit yang ditampilkan di penerima. Tanpa pemasangan, tidak ada yang bisa dikirim dan tidak ada folder yang bisa dijelajahi.
- Enkripsi TLS 1.3. Setiap kali dijalankan, sertifikat swa-tanda tangan (self-signed) baru dibuat; sidik jarinya disematkan (pinned) saat pemasangan dan diperiksa pada setiap koneksi.
- File dan seluruh folder, dengan struktur folder dipertahankan dan hingga 4 file dikirim secara paralel.
- Melanjutkan transfer yang terputus dan pemeriksaan SHA-256 untuk setiap file.
- Data yang belum selesai disimpan dalam folder tersembunyi `.lantransfer-partial`, sehingga file yang terpotong tidak pernah muncul dengan nama akhirnya.
- Bentrokan nama menghasilkan `file (1).txt` atau `Folder (1)`; file yang sudah ada tidak pernah ditimpa.
- Nama berbahaya ditolak (`..`, path absolut, alternate data stream `file:stream`, nama khusus seperti `CON`, `NUL`, `COM1`). Ruang disk kosong diperiksa sebelum menerima.
- Seret dan lepas dari Explorer, antrean dengan progres per file, kecepatan, sisa waktu, pembatalan, melanjutkan, serta riwayat item yang dikirim dan diterima.
- Portabel: satu file EXE, pengaturan dan riwayat disimpan di folder `data` di sebelahnya.
- Jendela, judulnya, dan keluaran baris perintah tersedia dalam 30 bahasa (sama dengan README ini), dipilih otomatis berdasarkan bahasa tampilan Windows; tata letak kanan-ke-kiri untuk bahasa Arab dan Ibrani.

## Bahasa antarmuka

LAN Transfer tersedia dalam 30 bahasa yang sama dengan README ini. Saat pertama kali dijalankan, aplikasi menggunakan bahasa tampilan Windows (bahasa pertama yang didukung dalam daftar bahasa Windows); jika Windows menggunakan bahasa yang tidak ada dalam daftar, antarmuka ditampilkan dalam bahasa Inggris.

Untuk mengubahnya, klik tombol roda gigi (**Pengaturan**) di pojok kanan atas, lalu pilih bahasa. Jendela, judulnya, dan semua pesan langsung berganti, dan pilihan Anda disimpan di `data\settings.json`. **Otomatis (Windows)** kembali mengikuti bahasa tampilan Windows. Angka, ukuran, kecepatan, tanggal, dan bentuk jamak mengikuti aturan bahasa yang dipilih; bahasa Arab dan Ibrani menggunakan tata letak kanan-ke-kiri yang dicerminkan, sedangkan path, alamat IP, kode, dan sidik jari sertifikat tetap ditulis dari kiri ke kanan. Pesan dari komputer lain (misalnya “ruang disk penerima tidak cukup”) ditampilkan dalam bahasa Anda, bukan dalam bahasa penerima.

`LANTransferCLI.exe` menggunakan bahasa yang sama. `LANTRANSFER_LANG` (misalnya `LANTRANSFER_LANG=de`) menggantikan pilihan yang disimpan, baik untuk jendela maupun baris perintah.

| English | Deutsch | العربية (RTL) |
|---|---|---|
| ![LAN Transfer dalam bahasa Inggris](../docs/screenshots/lang-en.png) | ![LAN Transfer dalam bahasa Jerman](../docs/screenshots/lang-de.png) | ![LAN Transfer dalam bahasa Arab, tata letak kanan-ke-kiri](../docs/screenshots/lang-ar.png) |

## Tangkapan layar

| Pemasangan | Pengiriman dengan progres |
|---|---|
| ![Dialog pemasangan dengan kode 6 digit dan sidik jari sertifikat](../docs/screenshots/pairing.png) | ![Antrean transfer dengan progres, kecepatan, dan sisa waktu](../docs/screenshots/sending.png) |
| **Melanjutkan setelah dibatalkan** | **Penerima: riwayat** |
| ![Transfer yang dibatalkan dilanjutkan dari posisi yang sudah diterima, terverifikasi SHA-256](../docs/screenshots/resume.png) | ![Jendela penerima dengan item yang diterima di riwayat](../docs/screenshots/history.png) |

## Unduh dan jalankan

1. Unduh `LANTransfer-2.1.0-windows-amd64.zip` dari [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) dan verifikasi dengan file `.sha256` yang dipublikasikan.
2. Ekstrak ke folder mana pun di kedua komputer.
3. Jalankan `LANTransfer.exe`. Saat pertama kali dijalankan, Windows meminta izin (UAC) untuk menambahkan dua aturan firewall; lihat [Port dan firewall](#port-dan-firewall).

Persyaratan: Windows 10 atau 11 (64-bit) dengan runtime WebView2, yang sudah menjadi bagian dari Windows versi terbaru. Kedua komputer harus menjalankan versi 2. Versi 2 tidak dapat berkomunikasi dengan versi 1.

## Cara memasangkan

1. Jalankan LAN Transfer di kedua komputer. Penerima menampilkan kode pemasangan dan sidik jari sertifikatnya di pojok kanan atas.
2. Di komputer pengirim, klik penerima di daftar **Penerima**. Jika tidak muncul, gunakan **Sambungkan secara manual (alamat IP)** dan masukkan alamat IP-nya.
3. Ketik kode 6 digit dari layar penerima dan bandingkan sidik jari sertifikat yang ditampilkan di dialog dengan yang ada di penerima. Jika berbeda, jangan lanjutkan.
4. Klik **Pasangkan**. Perangkat kini menampilkan **✓ dipasangkan**.

Setelah 5 kode salah dari satu komputer, penerima memblokir percobaan berikutnya dari komputer tersebut. **Kode baru** membuat kode baru dan membatalkan semua pemasangan yang ada. Penerima membuat kode dan sertifikat baru setiap kali dijalankan, jadi setelah penerima dimulai ulang, lakukan pemasangan lagi.

## Mengirim file

1. Pasangkan dengan penerima.
2. Seret file atau folder ke jendela, gunakan **Pilih file** / **Pilih folder**, atau tempel sebuah path.
3. Pilih tujuan: folder bawaan penerima (`Downloads\LAN Transfer`) atau folder lain yang sudah ada di penerima.
4. Klik **Kirim**. Setiap item menampilkan progres, kecepatan, dan sisa waktu, serta dapat dibatalkan. Transfer yang dibatalkan atau terputus dapat dilanjutkan; hanya bagian yang belum terkirim yang dikirim.

## Baris perintah

`LANTransferCLI.exe` adalah program yang sama tanpa jendela, untuk skrip dan otomatisasi.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Kode juga dapat diberikan melalui `LANTRANSFER_CODE`. Penerima tanpa jendela dapat menggunakan kode tetap dari `LANTRANSFER_PIN`. Menjalankan perintah `send` yang sama sekali lagi akan melanjutkan transfer yang terputus.

## Port dan firewall

- UDP 45454: penemuan. TCP 45455: transfer (HTTPS).
- Saat pertama kali dijalankan, aplikasi meminta izin satu kali (UAC) untuk menambahkan dua aturan masuk: `LAN Transfer 2 Discovery (UDP)` dan `LAN Transfer 2 Transfer (TCP)`.
- Aturan tersebut hanya menerima koneksi dari subnet lokal (`remoteip=LocalSubnet`) dan hanya pada profil jaringan Pribadi dan Domain (`profile=private,domain`). Di jaringan yang ditandai Windows sebagai Publik (kafe, hotel, bandara), komputer lain tidak dapat terhubung ke LAN Transfer.
- Jika penerima tidak terlihat di jaringan rumah Anda, periksa apakah jaringan tersebut diatur sebagai Pribadi di pengaturan jaringan Windows.
- Versi 1 membuat aturan bernama `LAN Transfer Discovery` dan `LAN Transfer Transfer` untuk semua profil, termasuk Publik. Jika versi 2 menemukannya, versi 2 menampilkan banner yang menawarkan untuk menggantinya. Tidak ada yang diubah sampai Anda mengklik **Perbarui aturan** dan mengonfirmasi permintaan Windows.

## Catatan keamanan

- Penemuan melalui UDP tidak diautentikasi; ini hanya petunjuk. Kepercayaan berasal dari pemasangan: kode ditambah sidik jari sertifikat yang disematkan.
- Selama pemasangan pertama, penyerang aktif di jaringan yang sama (man in the middle) dapat menyamar sebagai penerima. Bandingkan sidik jari sertifikat di kedua layar sebelum memasangkan.
- Pengirim yang sudah dipasangkan dapat menjelajahi folder penerima dan menyimpan file ke folder mana pun yang sudah ada di sana (opsi tujuan kustom). Pasangkan hanya komputer yang Anda percayai, dan gunakan **Kode baru** untuk mencabut semua pemasangan.
- Aturan firewall tidak membuka port di jaringan yang ditandai sebagai Publik.

## Variabel lingkungan

Variabel ini ditujukan untuk pengujian dan otomatisasi; penggunaan normal tidak memerlukan satu pun.

| Variabel | Arti |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | port lain |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | folder penerimaan, folder data program |
| `LANTRANSFER_NAME` | nama komputer yang ditampilkan kepada yang lain |
| `LANTRANSFER_LANG` | bahasa antarmuka dan CLI (`en`, `de`, `pl`, …); menggantikan pilihan yang disimpan |
| `LANTRANSFER_LOOPBACK=1` | hanya 127.0.0.1: tanpa broadcast, tanpa pemindaian LAN, tanpa aturan firewall |
| `LANTRANSFER_PEER_UDP_PORTS` | port UDP instans lain dalam mode loopback |
| `LANTRANSFER_NO_FIREWALL=1` | jangan membuat atau memeriksa aturan firewall |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | kode tetap penerima, kode yang digunakan pengirim |

## Build dan pengujian

Go 1.23 atau lebih baru di Windows 10/11.

```powershell
build.bat
go test ./...
python tools/check_i18n.py
```

`build.bat` menjalankan pengujian dan membangun `LANTransfer.exe` (jendela) dan `LANTransferCLI.exe` (konsol) dengan versi dari `VERSION` (bawaan `2.1.0`). Pengujian mencakup transfer end-to-end melalui 127.0.0.1: pemasangan dan pemblokiran, penolakan permintaan yang tidak diautentikasi, file dari 0 B hingga 96 MB, folder Unicode dan path lebih dari 260 karakter, bentrokan nama, nama berbahaya, checksum yang salah, pembatalan dan melanjutkan, disk penuh, serta penyematan sertifikat.

Terjemahan disimpan di `locales/ui.<code>.json` (bahasa Polandia sebagai sumber, bahasa Inggris sebagai cadangan). `tools/check_i18n.py` memeriksa bahwa ke-30 file memiliki kunci, placeholder, dan bentuk jamak yang sama serta tidak berisi sisa teks bahasa Polandia; sebuah pengujian Go memeriksa bahwa setiap kunci yang digunakan program ada di setiap bahasa.

Catatan rilis: [docs/releases/v2.1.0.md](../docs/releases/v2.1.0.md) · [v2.0.0](../docs/releases/v2.0.0.md).

## Lisensi

MIT, lihat [LICENSE](../LICENSE).

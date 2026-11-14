# LAN Transfer

[Polski](#polski) · [English](#english)

---

## Polski

Lekki program do bezpośredniego przesyłania plików i folderów w sieci lokalnej –
bez chmury i bez serwera pośredniego. Rdzeń w Go, okno WebView2 (HTML/CSS/JS bez
Reacta i Electrona) oraz CLI dla skryptów i agentów AI.

### Funkcje

- automatyczne wykrywanie komputerów w LAN (rozgłoszenie UDP, zapasowo skan /24),
- **parowanie 6-cyfrowym kodem** widocznym na komputerze odbiorcy – bez kodu nie da
  się nic wysłać ani przeglądać folderów,
- **szyfrowanie TLS 1.3** (certyfikat generowany przy starcie, odcisk zapamiętywany
  przy parowaniu i porównywany przy każdym połączeniu),
- pliki i całe foldery (struktura zachowana, do 4 plików równolegle),
- **wznawianie** przerwanych transferów i **suma SHA-256** każdego pliku,
- niedokończone dane trafiają do ukrytego `.lantransfer-partial` – nigdy nie
  powstaje ucięty plik pod docelową nazwą,
- kolizje nazw → `plik (1).txt`, `Folder (1)`; nic nie jest nadpisywane,
- odrzucanie niebezpiecznych nazw (`..`, ścieżki bezwzględne, `plik:strumień`,
  `CON`, `NUL`, `COM1`…), sprawdzanie wolnego miejsca,
- interfejs: przeciągnij i upuść, kolejka z postępem każdego pliku, prędkość, czas
  do końca, anulowanie, wznawianie, historia, czytelne komunikaty błędów.

### Budowanie

Wymagany Go 1.23+ i Windows 10/11 (WebView2 jest w systemie).

```powershell
build.bat
go test ./...
```

Powstają `LANTransfer.exe` (okno) i `LANTransferCLI.exe` (konsola).

### Użycie

1. Uruchom LAN Transfer na obu komputerach.
2. Na nadawcy kliknij odbiorcę i wpisz kod parowania z ekranu odbiorcy
   (sprawdź też, czy odcisk certyfikatu jest taki sam).
3. Przeciągnij pliki/foldery i kliknij **Wyślij**. Pliki trafiają do
   `Pobrane\LAN Transfer` albo do folderu wskazanego na odbiorcy.

„Nowy kod” unieważnia stary kod i wszystkie dotychczasowe parowania.
Po restarcie odbiorcy trzeba sparować się ponownie.

### CLI

```powershell
LANTransferCLI.exe serve                          # odbiór; wypisuje kod parowania
LANTransferCLI.exe devices                        # lista urządzeń (JSON)
LANTransferCLI.exe send plik.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Odbior] [--parallel 4] [--json]
LANTransferCLI.exe send "C:\sciezka\plik.zip" NAZWA-KOMPUTERA --code 123456
```

Kod można też podać w `LANTRANSFER_CODE`. Dla odbiornika bez okna stały kod
ustawia `LANTRANSFER_PIN`. Przerwane `send` wznawia się, uruchamiając tę samą
komendę ponownie.

### Porty i zapora

- UDP 45454 – wykrywanie, TCP 45455 – transfer (HTTPS).
- Przy pierwszym starcie program prosi (UAC) o dodanie dwóch reguł zapory
  ograniczonych do lokalnej podsieci (`remoteip=LocalSubnet`).

### Zmienne środowiskowe (testy, automatyzacja)

| Zmienna | Znaczenie |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | inne porty |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | folder odbioru / danych programu |
| `LANTRANSFER_LOOPBACK=1` | tylko 127.0.0.1: brak rozgłoszeń i skanowania LAN (testy na jednym komputerze) |
| `LANTRANSFER_PEER_UDP_PORTS` | porty UDP innych instancji w trybie loopback |
| `LANTRANSFER_NO_FIREWALL=1` | nie proponuj reguł zapory |
| `LANTRANSFER_PIN` / `LANTRANSFER_CODE` | stały kod odbiorcy / kod podawany przez nadawcę |

### Bezpieczeństwo – ograniczenia

- Wykrywanie (UDP) nie jest uwierzytelnione – to tylko podpowiedź. Zaufanie daje
  parowanie: kod + przypięty odcisk certyfikatu.
- Przy *pierwszym* parowaniu aktywny atakujący w tej samej sieci (MITM) mógłby
  podszyć się pod odbiorcę. Dlatego warto porównać odcisk certyfikatu na obu
  ekranach.
- Sparowany nadawca może zapisać pliki w dowolnym istniejącym folderze odbiorcy
  (funkcja „Wybrany folder na odbiorcy”) i przeglądać jego foldery. Paruj tylko
  zaufane komputery.

### Testy

`go test ./...` uruchamia m.in. testy end-to-end na 127.0.0.1: parowanie i
blokada po błędnych kodach, odrzucanie nieuwierzytelnionych żądań, pliki 0 B –
96 MB, foldery z Unicode i ścieżkami > 260 znaków, kolizje, wrogie nazwy,
błędna suma kontrolna, anulowanie i wznowienie, brak miejsca, limit rozmiaru,
przypinanie certyfikatu, kolejka w GUI. Benchmark: zobacz
`internal/service/bench_test.go`.

---

## English

A small tool for sending files and folders directly between computers on a
local network – no cloud, no relay server. Go core, a WebView2 window
(plain HTML/CSS/JS) and a CLI for scripts and AI agents. The UI is in Polish.

### Features

- automatic discovery (UDP broadcast, /24 scan as a fallback),
- **pairing with a 6-digit code** shown on the receiver – nothing can be sent
  or browsed without it,
- **TLS 1.3** with a per-process self-signed certificate whose fingerprint is
  pinned at pairing,
- files and whole folders (structure preserved, 4 files in parallel),
- **resume** of interrupted transfers and a **SHA-256** check for every file,
- partial data lives in a hidden `.lantransfer-partial` folder; a truncated
  file never appears under its final name,
- name collisions become `file (1).txt` / `Folder (1)`; nothing is overwritten,
- hostile names (`..`, absolute paths, `file:stream`, `CON`, `NUL`, `COM1`…)
  are rejected; free space is checked before receiving,
- UI: drag and drop, queue with per-file progress, speed, ETA, cancel, resume,
  history and clear error messages.

### Build

Go 1.23+ on Windows 10/11 (WebView2 ships with the OS).

```powershell
build.bat
go test ./...
```

### Usage

Start LAN Transfer on both computers, click the receiver, type the pairing
code displayed on the receiver (and compare the certificate fingerprint), then
drop files and press **Wyślij** (Send). See the CLI section above for
automation; `LANTRANSFER_CODE` / `LANTRANSFER_PIN` provide codes
non-interactively. Re-running an interrupted `send` resumes it.

### Ports and firewall

UDP 45454 (discovery) and TCP 45455 (HTTPS transfer). On first start the app
asks (UAC) to add two inbound rules limited to the local subnet.

### Security notes

Discovery is unauthenticated and only a hint; trust comes from pairing (code +
pinned certificate). An active man-in-the-middle during the *first* pairing is
not prevented by the protocol – compare the fingerprint shown on both screens.
A paired sender can browse the receiver's folders and write into any existing
folder there, so pair only computers you trust.

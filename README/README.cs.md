**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#stažení-a-spuštění)

Malý přenosný nástroj pro posílání souborů a složek přímo mezi dvěma počítači ve stejné místní síti. Žádný cloud, žádný účet a žádný přenosový (relay) server: data putují přímo z jednoho počítače na druhý přes šifrované spojení. Jádro v Go, okno WebView2 a verze pro příkazový řádek pro skripty.

![Hlavní okno LAN Transfer: spárovaný příjemce a tři položky připravené k odeslání](../docs/screenshots/main.png)

## Obsah

- [Funkce](#funkce)
- [Snímky obrazovky](#snímky-obrazovky)
- [Stažení a spuštění](#stažení-a-spuštění)
- [Jak spárovat](#jak-spárovat)
- [Odesílání souborů](#odesílání-souborů)
- [Příkazový řádek](#příkazový-řádek)
- [Porty a firewall](#porty-a-firewall)
- [Poznámky k zabezpečení](#poznámky-k-zabezpečení)
- [Proměnné prostředí](#proměnné-prostředí)
- [Sestavení a testy](#sestavení-a-testy)
- [Licence](#licence)

## Funkce

- Automatické vyhledávání dalších počítačů se spuštěným LAN Transfer (UDP broadcast, se zpomaleným skenováním /24 jako záložním řešením).
- Párování pomocí 6místného kódu zobrazeného na příjemci. Bez spárování nelze nic odeslat ani procházet žádné složky.
- Šifrování TLS 1.3. Každé spuštění vytvoří nový certifikát podepsaný sám sebou; jeho otisk se při párování připne (pinning) a ověřuje se při každém připojení.
- Soubory i celé složky, se zachováním struktury složek a až 4 soubory odesílanými paralelně.
- Navázání přerušených přenosů a kontrola SHA-256 u každého souboru.
- Nedokončená data se ukládají ve skryté složce `.lantransfer-partial`, takže se zkrácený soubor nikdy neobjeví pod svým konečným názvem.
- Při kolizi názvů vznikne `file (1).txt` nebo `Folder (1)`; existující soubory se nikdy nepřepisují.
- Nebezpečné názvy se odmítají (`..`, absolutní cesty, alternativní datové proudy `file:stream`, rezervované názvy jako `CON`, `NUL`, `COM1`). Před příjmem se kontroluje volné místo na disku.
- Přetažení z Průzkumníka, fronta s průběhem pro každý soubor, rychlostí, zbývajícím časem, zrušením, navázáním a historií odeslaných a přijatých položek.
- Přenosná aplikace: jediný soubor EXE, nastavení a historie ve složce `data` vedle něj.

Uživatelské rozhraní je zatím v polštině.

## Snímky obrazovky

| Párování | Odesílání s průběhem |
|---|---|
| ![Dialog párování se 6místným kódem a otiskem certifikátu](../docs/screenshots/pairing.png) | ![Fronta přenosů s průběhem, rychlostí a zbývajícím časem](../docs/screenshots/sending.png) |
| **Navázání po zrušení** | **Příjemce: historie** |
| ![Zrušený přenos navázaný od přijaté pozice, ověřeno SHA-256](../docs/screenshots/resume.png) | ![Okno příjemce s přijatými položkami v historii](../docs/screenshots/history.png) |

## Stažení a spuštění

1. Stáhněte `LANTransfer-2.0.0-windows-amd64.zip` z [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) a ověřte jej podle zveřejněného souboru `.sha256`.
2. Rozbalte jej do libovolné složky na obou počítačích.
3. Spusťte `LANTransfer.exe`. Při prvním spuštění Windows požádá (UAC) o povolení přidat dvě pravidla brány firewall; viz [Porty a firewall](#porty-a-firewall).

Požadavky: Windows 10 nebo 11 (64bitový) s modulem runtime WebView2, který je součástí aktuálních verzí Windows. Na obou počítačích musí běžet verze 2. Verze 2 nedokáže komunikovat s verzí 1.

## Jak spárovat

1. Spusťte LAN Transfer na obou počítačích. Příjemce zobrazuje svůj párovací kód a otisk certifikátu v pravém horním rohu.
2. Na odesílajícím počítači klikněte na příjemce v seznamu **Odbiorca** (Příjemce). Pokud se nezobrazí, použijte **Połącz ręcznie** (Připojit ručně) a zadejte jeho IP adresu.
3. Zadejte 6místný kód z obrazovky příjemce a porovnejte otisk certifikátu zobrazený v dialogu s otiskem na příjemci. Pokud se liší, nepokračujte.
4. Klikněte na **Sparuj** (Spárovat). Zařízení nyní zobrazuje **sparowano** (spárováno).

Po 5 chybných kódech z jednoho počítače příjemce zablokuje další pokusy z tohoto počítače. **Nowy kod** (Nový kód) vytvoří nový kód a zruší všechna existující spárování. Příjemce při každém spuštění vytváří nový kód a certifikát, takže po restartu příjemce je nutné znovu spárovat.

## Odesílání souborů

1. Spárujte se s příjemcem.
2. Přetáhněte soubory nebo složky do okna, použijte **Wybierz pliki** / **Wybierz folder** (Vybrat soubory / Vybrat složku) nebo vložte cestu.
3. Zvolte cíl: výchozí složku příjemce (`Downloads\LAN Transfer`) nebo jinou existující složku na příjemci.
4. Klikněte na **Wyślij** (Odeslat). Každá položka zobrazuje průběh, rychlost a zbývající čas a lze ji zrušit. Zrušený nebo přerušený přenos lze navázat; odešle se pouze chybějící část.

## Příkazový řádek

`LANTransferCLI.exe` je stejný program bez okna, určený pro skripty a automatizaci.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Kód lze předat také v `LANTRANSFER_CODE`. Příjemce bez okna může používat pevný kód z `LANTRANSFER_PIN`. Opětovné spuštění stejného příkazu `send` naváže přerušený přenos.

## Porty a firewall

- UDP 45454: vyhledávání. TCP 45455: přenos (HTTPS).
- Při prvním spuštění aplikace jednou požádá (UAC) o přidání dvou příchozích pravidel: `LAN Transfer 2 Discovery (UDP)` a `LAN Transfer 2 Transfer (TCP)`.
- Pravidla přijímají spojení pouze z místní podsítě (`remoteip=LocalSubnet`) a pouze v síťových profilech Soukromá a Doména (`profile=private,domain`). V síti, kterou Windows označí jako Veřejnou (kavárna, hotel, letiště), se ostatní počítače k LAN Transfer připojit nemohou.
- Pokud příjemce ve vaší domácí síti není vidět, zkontrolujte v nastavení sítě Windows, že je síť nastavena jako Soukromá.
- Verze 1 vytvářela pravidla s názvy `LAN Transfer Discovery` a `LAN Transfer Transfer` pro všechny profily, včetně Veřejné. Pokud je verze 2 najde, zobrazí banner s nabídkou jejich nahrazení. Nic se nezmění, dokud nekliknete na **Zaktualizuj reguły** (Aktualizovat pravidla) a nepotvrdíte výzvu Windows.

## Poznámky k zabezpečení

- Vyhledávání přes UDP není ověřované; slouží pouze jako nápověda. Důvěra vychází z párování: z kódu a připnutého otisku certifikátu.
- Během prvního párování by se aktivní útočník ve stejné síti (man-in-the-middle) mohl vydávat za příjemce. Před spárováním porovnejte otisk certifikátu na obou obrazovkách.
- Spárovaný odesílatel může procházet složky příjemce a ukládat soubory do kterékoli tamní existující složky (volba vlastního cíle). Párujte pouze počítače, kterým důvěřujete, a k odvolání všech spárování použijte **Nowy kod**.
- Pravidla brány firewall neotevírají porty v sítích označených jako Veřejné.

## Proměnné prostředí

Jsou určeny pro testy a automatizaci; při běžném používání není potřeba žádná z nich.

| Proměnná | Význam |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | jiné porty |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | složka pro příjem, složka s daty programu |
| `LANTRANSFER_NAME` | název počítače zobrazovaný ostatním |
| `LANTRANSFER_LOOPBACK=1` | pouze 127.0.0.1: žádné broadcasty, žádné skenování LAN, žádná pravidla firewallu |
| `LANTRANSFER_PEER_UDP_PORTS` | UDP porty ostatních instancí v režimu loopback |
| `LANTRANSFER_NO_FIREWALL=1` | nevytvářet ani nekontrolovat pravidla firewallu |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | pevný kód příjemce, kód používaný odesílatelem |

## Sestavení a testy

Go 1.23 nebo novější ve Windows 10/11.

```powershell
build.bat
go test ./...
```

`build.bat` spustí testy a sestaví `LANTransfer.exe` (okno) a `LANTransferCLI.exe` (konzole) s verzí ze souboru `VERSION` (výchozí `2.0.0`). Testy zahrnují end-to-end přenosy přes 127.0.0.1: párování a blokování, odmítnuté neověřené požadavky, soubory od 0 B do 96 MB, složky s Unicode a cesty delší než 260 znaků, kolize názvů, škodlivé názvy, chybný kontrolní součet, zrušení a navázání, plný disk a připínání certifikátu.

Poznámky k vydání: [docs/releases/v2.0.0.md](../docs/releases/v2.0.0.md).

## Licence

MIT, viz [LICENSE](../LICENSE).

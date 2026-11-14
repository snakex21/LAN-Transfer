**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#ladda-ner-och-kör)

Ett litet portabelt verktyg för att skicka filer och mappar direkt mellan två datorer i samma lokala nätverk. Inget moln, inget konto och ingen reläserver: data går direkt från den ena datorn till den andra över en krypterad anslutning. Kärna i Go, ett WebView2-fönster och en kommandoradsversion för skript.

![Huvudfönstret i LAN Transfer: en parkopplad mottagare och tre objekt redo att skickas](../docs/screenshots/main.png)

## Innehåll

- [Funktioner](#funktioner)
- [Skärmbilder](#skärmbilder)
- [Ladda ner och kör](#ladda-ner-och-kör)
- [Så parkopplar du](#så-parkopplar-du)
- [Skicka filer](#skicka-filer)
- [Kommandorad](#kommandorad)
- [Portar och brandvägg](#portar-och-brandvägg)
- [Säkerhetsaspekter](#säkerhetsaspekter)
- [Miljövariabler](#miljövariabler)
- [Bygga och testa](#bygga-och-testa)
- [Licens](#licens)

## Funktioner

- Automatisk identifiering av andra datorer som kör LAN Transfer (UDP-broadcast, med en begränsad /24-skanning som reserv).
- Parkoppling med en 6-siffrig kod som visas på mottagaren. Utan parkoppling går det inte att skicka något eller bläddra i några mappar.
- TLS 1.3-kryptering. Varje körning skapar ett nytt självsignerat certifikat; dess fingeravtryck fästs (pinning) vid parkopplingen och kontrolleras vid varje anslutning.
- Filer och hela mappar, med bevarad mappstruktur och upp till 4 filer som skickas parallellt.
- Återupptagning av avbrutna överföringar och SHA-256-kontroll av varje fil.
- Ofullständiga data sparas i en dold mapp `.lantransfer-partial`, så att en avkortad fil aldrig visas under sitt slutliga namn.
- Vid namnkonflikter blir namnet `file (1).txt` eller `Folder (1)`; befintliga filer skrivs aldrig över.
- Farliga namn avvisas (`..`, absoluta sökvägar, alternativa dataströmmar `file:stream`, reserverade namn som `CON`, `NUL`, `COM1`). Ledigt diskutrymme kontrolleras innan mottagning.
- Dra och släpp från Utforskaren, en kö med förlopp per fil, hastighet, återstående tid, avbryt, återuppta och en historik över skickade och mottagna objekt.
- Portabelt: en enda EXE-fil, med inställningar och historik i en mapp `data` bredvid den.

Användargränssnittet är för närvarande på polska.

## Skärmbilder

| Parkoppling | Sändning med förlopp |
|---|---|
| ![Parkopplingsdialog med den 6-siffriga koden och certifikatets fingeravtryck](../docs/screenshots/pairing.png) | ![Överföringskö med förlopp, hastighet och återstående tid](../docs/screenshots/sending.png) |
| **Återupptagning efter avbrott** | **Mottagare: historik** |
| ![En avbruten överföring återupptagen från den mottagna positionen, SHA-256 verifierad](../docs/screenshots/resume.png) | ![Mottagarens fönster med mottagna objekt i historiken](../docs/screenshots/history.png) |

## Ladda ner och kör

1. Ladda ner `LANTransfer-2.0.0-windows-amd64.zip` från [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) och kontrollera den mot den publicerade `.sha256`-filen.
2. Packa upp den till valfri mapp på båda datorerna.
3. Starta `LANTransfer.exe`. Vid första starten ber Windows (UAC) om tillåtelse att lägga till två brandväggsregler; se [Portar och brandvägg](#portar-och-brandvägg).

Krav: Windows 10 eller 11 (64-bitars) med WebView2-runtime, som ingår i aktuella Windows-versioner. Båda datorerna måste köra version 2. Version 2 kan inte kommunicera med version 1.

## Så parkopplar du

1. Starta LAN Transfer på båda datorerna. Mottagaren visar sin parkopplingskod och sitt certifikatfingeravtryck i det övre högra hörnet.
2. Klicka på mottagaren i listan **Odbiorca** (Mottagare) på den sändande datorn. Om den inte visas använder du **Połącz ręcznie** (Anslut manuellt) och anger dess IP-adress.
3. Skriv in den 6-siffriga koden från mottagarens skärm och jämför certifikatets fingeravtryck i dialogrutan med det som visas på mottagaren. Om de skiljer sig åt ska du inte fortsätta.
4. Klicka på **Sparuj** (Parkoppla). Enheten visar nu **sparowano** (parkopplad).

Efter 5 felaktiga koder från en dator blockerar mottagaren ytterligare försök från den datorn. **Nowy kod** (Ny kod) skapar en ny kod och upphäver alla befintliga parkopplingar. Mottagaren skapar en ny kod och ett nytt certifikat varje gång den startar, så parkoppla igen efter att mottagaren har startats om.

## Skicka filer

1. Parkoppla med mottagaren.
2. Dra filer eller mappar till fönstret, använd **Wybierz pliki** / **Wybierz folder** (Välj filer / Välj mapp) eller klistra in en sökväg.
3. Välj mål: mottagarens standardmapp (`Downloads\LAN Transfer`) eller en annan befintlig mapp på mottagaren.
4. Klicka på **Wyślij** (Skicka). Varje objekt visar förlopp, hastighet och återstående tid och kan avbrytas. En överföring som har avbrutits manuellt eller av andra orsaker kan återupptas; endast den saknade delen skickas.

## Kommandorad

`LANTransferCLI.exe` är samma program utan fönster, för skript och automatisering.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Koden kan också anges i `LANTRANSFER_CODE`. En mottagare utan fönster kan använda en fast kod från `LANTRANSFER_PIN`. Om samma `send`-kommando körs igen återupptas en avbruten överföring.

## Portar och brandvägg

- UDP 45454: identifiering. TCP 45455: överföring (HTTPS).
- Vid första starten ber appen en gång (UAC) om att få lägga till två inkommande regler: `LAN Transfer 2 Discovery (UDP)` och `LAN Transfer 2 Transfer (TCP)`.
- Reglerna accepterar endast anslutningar från det lokala undernätet (`remoteip=LocalSubnet`) och endast i nätverksprofilerna Privat och Domän (`profile=private,domain`). I ett nätverk som Windows markerar som Offentligt (kafé, hotell, flygplats) kan andra datorer inte ansluta till LAN Transfer.
- Om mottagaren inte syns i ditt hemnätverk kontrollerar du att nätverket är inställt som Privat i nätverksinställningarna i Windows.
- Version 1 skapade regler med namnen `LAN Transfer Discovery` och `LAN Transfer Transfer` för alla profiler, även Offentligt. Om version 2 hittar dem visas en banderoll som erbjuder att ersätta dem. Ingenting ändras förrän du klickar på **Zaktualizuj reguły** (Uppdatera regler) och bekräftar Windows-frågan.

## Säkerhetsaspekter

- Identifiering via UDP är inte autentiserad; den är bara en ledtråd. Förtroendet kommer från parkopplingen: koden plus det fästa certifikatfingeravtrycket.
- Vid den första parkopplingen kan en aktiv angripare i samma nätverk (man-in-the-middle) utge sig för att vara mottagaren. Jämför certifikatets fingeravtryck på båda skärmarna innan du parkopplar.
- En parkopplad avsändare kan bläddra i mottagarens mappar och spara filer i valfri befintlig mapp där (alternativet för anpassat mål). Parkoppla bara datorer du litar på och använd **Nowy kod** för att återkalla alla parkopplingar.
- Brandväggsreglerna öppnar inte portarna i nätverk som är markerade som Offentliga.

## Miljövariabler

Dessa är avsedda för tester och automatisering; vid normal användning behövs ingen av dem.

| Variabel | Betydelse |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | andra portar |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | mottagningsmapp, mapp för programdata |
| `LANTRANSFER_NAME` | datornamn som visas för andra |
| `LANTRANSFER_LOOPBACK=1` | endast 127.0.0.1: inga broadcasts, ingen LAN-skanning, inga brandväggsregler |
| `LANTRANSFER_PEER_UDP_PORTS` | UDP-portar för andra instanser i loopback-läge |
| `LANTRANSFER_NO_FIREWALL=1` | skapa eller kontrollera inte brandväggsregler |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | mottagarens fasta kod, kod som avsändaren använder |

## Bygga och testa

Go 1.23 eller senare på Windows 10/11.

```powershell
build.bat
go test ./...
```

`build.bat` kör testerna och bygger `LANTransfer.exe` (fönster) och `LANTransferCLI.exe` (konsol) med versionen från `VERSION` (standard `2.0.0`). Testerna omfattar överföringar från början till slut över 127.0.0.1: parkoppling och spärr, avvisade oautentiserade förfrågningar, filer från 0 B till 96 MB, Unicode-mappar och sökvägar över 260 tecken, namnkonflikter, skadliga namn, en felaktig kontrollsumma, avbrytning och återupptagning, en full disk och certifikat-pinning.

Versionsinformation: [docs/releases/v2.0.0.md](../docs/releases/v2.0.0.md).

## Licens

MIT, se [LICENSE](../LICENSE).

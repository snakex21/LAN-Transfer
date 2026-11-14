**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#last-ned-og-kjør)

Et lite, portabelt verktøy for å sende filer og mapper direkte mellom to datamaskiner på samme lokale nettverk. Ingen sky, ingen konto og ingen reléserver: dataene går rett fra den ene PC-en til den andre over en kryptert tilkobling. Kjerne i Go, et WebView2-vindu og en kommandolinjeversjon for skript.

![Hovedvinduet i LAN Transfer: en paret mottaker og tre elementer klare til sending](../docs/screenshots/main.png)

## Innhold

- [Funksjoner](#funksjoner)
- [Skjermbilder](#skjermbilder)
- [Last ned og kjør](#last-ned-og-kjør)
- [Slik parer du](#slik-parer-du)
- [Sende filer](#sende-filer)
- [Kommandolinje](#kommandolinje)
- [Porter og brannmur](#porter-og-brannmur)
- [Sikkerhetsmerknader](#sikkerhetsmerknader)
- [Miljøvariabler](#miljøvariabler)
- [Bygging og tester](#bygging-og-tester)
- [Lisens](#lisens)

## Funksjoner

- Automatisk oppdagelse av andre datamaskiner som kjører LAN Transfer (UDP-kringkasting, med en begrenset /24-skanning som reserve).
- Paring med en 6-sifret kode som vises på mottakeren. Uten paring kan ingenting sendes, og ingen mapper kan blas gjennom.
- TLS 1.3-kryptering. Hver kjøring oppretter et nytt selvsignert sertifikat; fingeravtrykket festes (pinning) under paringen og kontrolleres ved hver tilkobling.
- Filer og hele mapper, med mappestrukturen bevart og opptil 4 filer sendt parallelt.
- Gjenopptakelse av avbrutte overføringer og SHA-256-kontroll av hver fil.
- Ufullstendige data lagres i en skjult mappe `.lantransfer-partial`, slik at en avkortet fil aldri dukker opp under sitt endelige navn.
- Navnekonflikter blir til `file (1).txt` eller `Folder (1)`; eksisterende filer blir aldri overskrevet.
- Farlige navn avvises (`..`, absolutte stier, alternative datastrømmer `file:stream`, reserverte navn som `CON`, `NUL`, `COM1`). Ledig diskplass kontrolleres før mottak.
- Dra og slipp fra Filutforsker, en kø med fremdrift per fil, hastighet, gjenstående tid, avbryt, gjenoppta og en historikk over sendte og mottatte elementer.
- Portabelt: én enkelt EXE-fil, med innstillinger og historikk i en mappe `data` ved siden av.

Brukergrensesnittet er foreløpig på polsk.

## Skjermbilder

| Paring | Sending med fremdrift |
|---|---|
| ![Paringsdialog med den 6-sifrede koden og sertifikatets fingeravtrykk](../docs/screenshots/pairing.png) | ![Overføringskø med fremdrift, hastighet og gjenstående tid](../docs/screenshots/sending.png) |
| **Gjenopptakelse etter avbrudd** | **Mottaker: historikk** |
| ![En avbrutt overføring gjenopptatt fra mottatt posisjon, SHA-256 verifisert](../docs/screenshots/resume.png) | ![Mottakervindu med mottatte elementer i historikken](../docs/screenshots/history.png) |

## Last ned og kjør

1. Last ned `LANTransfer-2.0.0-windows-amd64.zip` fra [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) og kontroller den mot den publiserte `.sha256`-filen.
2. Pakk den ut til en valgfri mappe på begge datamaskinene.
3. Start `LANTransfer.exe`. Ved første oppstart ber Windows (UAC) om tillatelse til å legge til to brannmurregler; se [Porter og brannmur](#porter-og-brannmur).

Krav: Windows 10 eller 11 (64-biters) med WebView2-kjøretiden, som er en del av gjeldende Windows-versjoner. Begge datamaskinene må kjøre versjon 2. Versjon 2 kan ikke kommunisere med versjon 1.

## Slik parer du

1. Start LAN Transfer på begge datamaskinene. Mottakeren viser paringskoden og sertifikatets fingeravtrykk øverst til høyre.
2. På den sendende datamaskinen klikker du på mottakeren i listen **Odbiorca** (Mottaker). Hvis den ikke vises, bruker du **Połącz ręcznie** (Koble til manuelt) og skriver inn IP-adressen.
3. Skriv inn den 6-sifrede koden fra mottakerens skjerm, og sammenlign sertifikatets fingeravtrykk i dialogboksen med det som vises på mottakeren. Hvis de er forskjellige, må du ikke fortsette.
4. Klikk på **Sparuj** (Par). Enheten viser nå **sparowano** (paret).

Etter 5 feil koder fra én datamaskin blokkerer mottakeren videre forsøk fra den. **Nowy kod** (Ny kode) oppretter en ny kode og opphever alle eksisterende paringer. Mottakeren oppretter en ny kode og et nytt sertifikat hver gang den startes, så du må pare på nytt etter at mottakeren er startet på nytt.

## Sende filer

1. Par med mottakeren.
2. Dra filer eller mapper til vinduet, bruk **Wybierz pliki** / **Wybierz folder** (Velg filer / Velg mappe), eller lim inn en sti.
3. Velg mål: mottakerens standardmappe (`Downloads\LAN Transfer`) eller en annen eksisterende mappe på mottakeren.
4. Klikk på **Wyślij** (Send). Hvert element viser fremdrift, hastighet og gjenstående tid og kan avbrytes. En overføring som er avbrutt manuelt eller av andre årsaker, kan gjenopptas; bare den manglende delen sendes.

## Kommandolinje

`LANTransferCLI.exe` er det samme programmet uten vindu, for skript og automatisering.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Koden kan også oppgis i `LANTRANSFER_CODE`. En mottaker uten vindu kan bruke en fast kode fra `LANTRANSFER_PIN`. Hvis du kjører den samme `send`-kommandoen på nytt, gjenopptas en avbrutt overføring.

## Porter og brannmur

- UDP 45454: oppdagelse. TCP 45455: overføring (HTTPS).
- Ved første oppstart ber appen én gang (UAC) om å legge til to innkommende regler: `LAN Transfer 2 Discovery (UDP)` og `LAN Transfer 2 Transfer (TCP)`.
- Reglene godtar bare tilkoblinger fra det lokale subnettet (`remoteip=LocalSubnet`) og bare på nettverksprofilene Privat og Domene (`profile=private,domain`). På et nettverk som Windows merker som Offentlig (kafé, hotell, flyplass), kan andre datamaskiner ikke koble til LAN Transfer.
- Hvis mottakeren ikke er synlig på hjemmenettverket ditt, må du kontrollere at nettverket er satt til Privat i nettverksinnstillingene i Windows.
- Versjon 1 opprettet regler kalt `LAN Transfer Discovery` og `LAN Transfer Transfer` for alle profiler, inkludert Offentlig. Hvis versjon 2 finner dem, vises et banner som tilbyr å erstatte dem. Ingenting endres før du klikker på **Zaktualizuj reguły** (Oppdater regler) og bekrefter Windows-forespørselen.

## Sikkerhetsmerknader

- Oppdagelse over UDP er ikke autentisert; den er bare et hint. Tillit kommer fra paringen: koden pluss det festede sertifikatfingeravtrykket.
- Under den første paringen kan en aktiv angriper på samme nettverk (man-in-the-middle) utgi seg for å være mottakeren. Sammenlign sertifikatets fingeravtrykk på begge skjermene før du parer.
- En paret avsender kan bla gjennom mottakerens mapper og lagre filer i hvilken som helst eksisterende mappe der (alternativet for egendefinert mål). Par bare datamaskiner du stoler på, og bruk **Nowy kod** for å tilbakekalle alle paringer.
- Brannmurreglene åpner ikke portene på nettverk som er merket som Offentlig.

## Miljøvariabler

Disse er ment for testing og automatisering; normal bruk krever ingen av dem.

| Variabel | Betydning |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | andre porter |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | mottaksmappe, mappe for programdata |
| `LANTRANSFER_NAME` | datamaskinnavn som vises for andre |
| `LANTRANSFER_LOOPBACK=1` | bare 127.0.0.1: ingen kringkasting, ingen LAN-skanning, ingen brannmurregler |
| `LANTRANSFER_PEER_UDP_PORTS` | UDP-porter for andre instanser i loopback-modus |
| `LANTRANSFER_NO_FIREWALL=1` | ikke opprett eller kontroller brannmurregler |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | fast kode for mottakeren, kode som brukes av avsenderen |

## Bygging og tester

Go 1.23 eller nyere på Windows 10/11.

```powershell
build.bat
go test ./...
```

`build.bat` kjører testene og bygger `LANTransfer.exe` (vindu) og `LANTransferCLI.exe` (konsoll) med versjonen fra `VERSION` (standard `2.0.0`). Testene omfatter ende-til-ende-overføringer over 127.0.0.1: paring og sperring, avviste uautentiserte forespørsler, filer fra 0 B til 96 MB, Unicode-mapper og stier over 260 tegn, navnekonflikter, ondsinnede navn, en feil kontrollsum, avbryting og gjenopptakelse, en full disk og sertifikatfesting.

Versjonsmerknader: [docs/releases/v2.0.0.md](../docs/releases/v2.0.0.md).

## Lisens

MIT, se [LICENSE](../LICENSE).

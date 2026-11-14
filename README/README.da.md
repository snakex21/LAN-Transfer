**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#download-og-kør)

Et lille, portabelt værktøj til at sende filer og mapper direkte mellem to computere på det samme lokale netværk. Ingen sky, ingen konto og ingen relæserver: dataene går direkte fra den ene pc til den anden over en krypteret forbindelse. Kerne i Go, et WebView2-vindue og en kommandolinjeversion til scripts.

![Hovedvinduet i LAN Transfer: en parret modtager og tre elementer klar til afsendelse](../docs/screenshots/main.png)

## Indhold

- [Funktioner](#funktioner)
- [Skærmbilleder](#skærmbilleder)
- [Download og kør](#download-og-kør)
- [Sådan parrer du](#sådan-parrer-du)
- [Send filer](#send-filer)
- [Kommandolinje](#kommandolinje)
- [Porte og firewall](#porte-og-firewall)
- [Sikkerhedsbemærkninger](#sikkerhedsbemærkninger)
- [Miljøvariabler](#miljøvariabler)
- [Bygning og test](#bygning-og-test)
- [Licens](#licens)

## Funktioner

- Automatisk registrering af andre computere, der kører LAN Transfer (UDP-broadcast med en begrænset /24-scanning som reserve).
- Parring med en 6-cifret kode, der vises på modtageren. Uden parring kan intet sendes, og ingen mapper kan gennemses.
- TLS 1.3-kryptering. Hver kørsel opretter et nyt selvsigneret certifikat; dets fingeraftryk fastlåses (pinning) under parringen og kontrolleres ved hver forbindelse.
- Filer og hele mapper, med mappestrukturen bevaret og op til 4 filer sendt parallelt.
- Genoptagelse af afbrudte overførsler og SHA-256-kontrol af hver fil.
- Ufærdige data gemmes i en skjult mappe `.lantransfer-partial`, så en afkortet fil aldrig optræder under sit endelige navn.
- Navnekonflikter bliver til `file (1).txt` eller `Folder (1)`; eksisterende filer bliver aldrig overskrevet.
- Farlige navne afvises (`..`, absolutte stier, alternative datastrømme `file:stream`, reserverede navne som `CON`, `NUL`, `COM1`). Ledig diskplads kontrolleres før modtagelse.
- Træk og slip fra Stifinder, en kø med status for hver fil, hastighed, resterende tid, annuller, genoptag og en historik over sendte og modtagne elementer.
- Portabelt: én enkelt EXE-fil, med indstillinger og historik i en mappe `data` ved siden af.

Brugergrænsefladen er i øjeblikket på polsk.

## Skærmbilleder

| Parring | Afsendelse med status |
|---|---|
| ![Parringsdialog med den 6-cifrede kode og certifikatets fingeraftryk](../docs/screenshots/pairing.png) | ![Overførselskø med status, hastighed og resterende tid](../docs/screenshots/sending.png) |
| **Genoptagelse efter annullering** | **Modtager: historik** |
| ![En annulleret overførsel genoptaget fra den modtagne position, SHA-256 verificeret](../docs/screenshots/resume.png) | ![Modtagervindue med modtagne elementer i historikken](../docs/screenshots/history.png) |

## Download og kør

1. Download `LANTransfer-2.0.0-windows-amd64.zip` fra [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest), og kontrollér den mod den offentliggjorte `.sha256`-fil.
2. Udpak den til en vilkårlig mappe på begge computere.
3. Start `LANTransfer.exe`. Ved første start beder Windows (UAC) om tilladelse til at tilføje to firewallregler; se [Porte og firewall](#porte-og-firewall).

Krav: Windows 10 eller 11 (64-bit) med WebView2-runtime, som er en del af aktuelle Windows-versioner. Begge computere skal køre version 2. Version 2 kan ikke kommunikere med version 1.

## Sådan parrer du

1. Start LAN Transfer på begge computere. Modtageren viser sin parringskode og sit certifikatfingeraftryk i øverste højre hjørne.
2. Klik på modtageren i listen **Odbiorca** (Modtager) på den afsendende computer. Hvis den ikke vises, skal du bruge **Połącz ręcznie** (Forbind manuelt) og indtaste dens IP-adresse.
3. Indtast den 6-cifrede kode fra modtagerens skærm, og sammenlign certifikatets fingeraftryk i dialogboksen med det, der vises på modtageren. Hvis de er forskellige, må du ikke fortsætte.
4. Klik på **Sparuj** (Par). Enheden viser nu **sparowano** (parret).

Efter 5 forkerte koder fra én computer blokerer modtageren yderligere forsøg fra den. **Nowy kod** (Ny kode) opretter en ny kode og annullerer alle eksisterende parringer. Modtageren opretter en ny kode og et nyt certifikat, hver gang den starter, så du skal parre igen, når modtageren er blevet genstartet.

## Send filer

1. Par med modtageren.
2. Træk filer eller mapper ind i vinduet, brug **Wybierz pliki** / **Wybierz folder** (Vælg filer / Vælg mappe), eller indsæt en sti.
3. Vælg destination: modtagerens standardmappe (`Downloads\LAN Transfer`) eller en anden eksisterende mappe på modtageren.
4. Klik på **Wyślij** (Send). Hvert element viser status, hastighed og resterende tid og kan annulleres. En annulleret eller afbrudt overførsel kan genoptages; kun den manglende del sendes.

## Kommandolinje

`LANTransferCLI.exe` er det samme program uden vindue, til scripts og automatisering.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Koden kan også angives i `LANTRANSFER_CODE`. En modtager uden vindue kan bruge en fast kode fra `LANTRANSFER_PIN`. Hvis den samme `send`-kommando køres igen, genoptages en afbrudt overførsel.

## Porte og firewall

- UDP 45454: registrering. TCP 45455: overførsel (HTTPS).
- Ved første start beder appen én gang (UAC) om at tilføje to indgående regler: `LAN Transfer 2 Discovery (UDP)` og `LAN Transfer 2 Transfer (TCP)`.
- Reglerne accepterer kun forbindelser fra det lokale undernet (`remoteip=LocalSubnet`) og kun på netværksprofilerne Privat og Domæne (`profile=private,domain`). På et netværk, som Windows markerer som Offentligt (café, hotel, lufthavn), kan andre computere ikke oprette forbindelse til LAN Transfer.
- Hvis modtageren ikke er synlig på dit hjemmenetværk, skal du kontrollere, at netværket er indstillet til Privat i netværksindstillingerne i Windows.
- Version 1 oprettede regler med navnene `LAN Transfer Discovery` og `LAN Transfer Transfer` for alle profiler, inklusive Offentligt. Hvis version 2 finder dem, vises et banner, der tilbyder at erstatte dem. Intet ændres, før du klikker på **Zaktualizuj reguły** (Opdater regler) og bekræfter Windows-prompten.

## Sikkerhedsbemærkninger

- Registrering via UDP er ikke godkendt; den er kun et fingerpeg. Tillid kommer fra parringen: koden plus det fastlåste certifikatfingeraftryk.
- Under den første parring kan en aktiv angriber på det samme netværk (man-in-the-middle) udgive sig for at være modtageren. Sammenlign certifikatets fingeraftryk på begge skærme, før du parrer.
- En parret afsender kan gennemse modtagerens mapper og gemme filer i en hvilken som helst eksisterende mappe dér (indstillingen for brugerdefineret destination). Par kun computere, du har tillid til, og brug **Nowy kod** til at tilbagekalde alle parringer.
- Firewallreglerne åbner ikke portene på netværk, der er markeret som Offentlige.

## Miljøvariabler

Disse er beregnet til test og automatisering; normal brug kræver ingen af dem.

| Variabel | Betydning |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | andre porte |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | modtagelsesmappe, mappe til programdata |
| `LANTRANSFER_NAME` | computernavn, der vises for andre |
| `LANTRANSFER_LOOPBACK=1` | kun 127.0.0.1: ingen broadcasts, ingen LAN-scanning, ingen firewallregler |
| `LANTRANSFER_PEER_UDP_PORTS` | UDP-porte for andre instanser i loopback-tilstand |
| `LANTRANSFER_NO_FIREWALL=1` | opret eller kontrollér ikke firewallregler |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | modtagerens faste kode, kode som afsenderen bruger |

## Bygning og test

Go 1.23 eller nyere på Windows 10/11.

```powershell
build.bat
go test ./...
```

`build.bat` kører testene og bygger `LANTransfer.exe` (vindue) og `LANTransferCLI.exe` (konsol) med versionen fra `VERSION` (standard `2.0.0`). Testene omfatter end-to-end-overførsler over 127.0.0.1: parring og spærring, afviste ikke-godkendte anmodninger, filer fra 0 B til 96 MB, Unicode-mapper og stier over 260 tegn, navnekonflikter, ondsindede navne, en forkert kontrolsum, annullering og genoptagelse, en fuld disk og certifikatfastlåsning.

Udgivelsesnoter: [docs/releases/v2.0.0.md](../docs/releases/v2.0.0.md).

## Licens

MIT, se [LICENSE](../LICENSE).

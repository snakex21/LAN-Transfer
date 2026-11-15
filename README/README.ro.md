**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#descărcare-și-pornire)

Un mic instrument portabil pentru trimiterea fișierelor și folderelor direct între două calculatoare din aceeași rețea locală. Fără cloud, fără cont și fără server intermediar: datele merg direct de pe un PC pe celălalt printr-o conexiune criptată. Nucleu scris în Go, o fereastră WebView2 și o versiune în linie de comandă pentru scripturi.

![Fereastra principală LAN Transfer: un destinatar asociat și trei elemente pregătite pentru trimitere](../docs/screenshots/main.png)

## Cuprins

- [Funcționalități](#funcționalități)
- [Limba interfeței](#limba-interfeței)
- [Capturi de ecran](#capturi-de-ecran)
- [Descărcare și pornire](#descărcare-și-pornire)
- [Cum se face asocierea](#cum-se-face-asocierea)
- [Trimiterea fișierelor](#trimiterea-fișierelor)
- [Linia de comandă](#linia-de-comandă)
- [Porturi și firewall](#porturi-și-firewall)
- [Note de securitate](#note-de-securitate)
- [Variabile de mediu](#variabile-de-mediu)
- [Compilare și teste](#compilare-și-teste)
- [Licență](#licență)

## Funcționalități

- Descoperirea automată a altor calculatoare care rulează LAN Transfer (UDP broadcast, cu o scanare /24 limitată ca rezervă).
- Asociere cu un cod de 6 cifre afișat pe destinatar. Fără asociere nu se poate trimite nimic și nu se pot răsfoi foldere.
- Mod opțional fără cod: pe destinatar, dezactivați **Solicită codul de asociere** din Setări (salvat în `data\settings.json`) sau porniți `serve --no-code`. Atunci orice computer cu LAN Transfer din rețea îi poate trimite fișiere fără cod, tot criptat; între timp antetul afișează **Fără cod**.
- Criptare TLS 1.3. La fiecare pornire se creează un nou certificat autosemnat; amprenta lui este fixată la asociere și verificată la fiecare conexiune.
- Fișiere și foldere întregi, cu păstrarea structurii folderelor și până la 4 fișiere trimise în paralel.
- Reluarea transferurilor întrerupte și verificarea SHA-256 a fiecărui fișier.
- Datele neterminate sunt păstrate într-un folder ascuns `.lantransfer-partial`, astfel încât un fișier incomplet nu apare niciodată sub numele său final.
- Coliziunile de nume devin `file (1).txt` sau `Folder (1)`; fișierele existente nu sunt niciodată suprascrise.
- Numele periculoase sunt respinse (`..`, căi absolute, fluxuri de date alternative `file:stream`, nume rezervate precum `CON`, `NUL`, `COM1`). Înainte de primire se verifică spațiul liber pe disc.
- Glisare și plasare din Explorer, o coadă cu progres pentru fiecare fișier, viteză, timp rămas, anulare, reluare și un istoric al elementelor trimise și primite.
- Portabil: un singur EXE, setările și istoricul într-un folder `data` aflat lângă el.
- Fereastra, titlul ei și mesajele din linia de comandă în 30 de limbi (aceleași ca în acest README), alese automat după limba de afișare din Windows; aspect de la dreapta la stânga pentru arabă și ebraică.

## Limba interfeței

LAN Transfer vorbește aceleași 30 de limbi ca acest README. La prima pornire folosește limba de afișare din Windows (prima limbă acceptată din lista de limbi Windows); dacă Windows folosește o limbă care nu se află pe listă, interfața este în engleză.

Pentru a o schimba, faceți clic pe butonul cu roată dințată (**Setări**) din colțul din dreapta sus și alegeți o limbă. Fereastra, titlul ei și toate mesajele se schimbă imediat, iar alegerea se salvează în `data\settings.json`. **Automat (Windows)** revine la limba de afișare din Windows. Numerele, dimensiunile, vitezele, datele și formele de plural respectă regulile limbii alese; araba și ebraica folosesc un aspect în oglindă, de la dreapta la stânga, în timp ce căile, adresele IP, codurile și amprentele rămân de la stânga la dreapta. Mesajele de la celălalt calculator (de exemplu „spațiu insuficient pe discul destinatarului”) sunt afișate în limba dvs., nu în cea a destinatarului.

`LANTransferCLI.exe` folosește aceeași limbă. `LANTRANSFER_LANG` (de exemplu `LANTRANSFER_LANG=de`) are prioritate față de alegerea salvată, atât în fereastră, cât și în linia de comandă.

| English | Deutsch | العربية (RTL) |
|---|---|---|
| ![LAN Transfer în engleză](../docs/screenshots/lang-en.png) | ![LAN Transfer în germană](../docs/screenshots/lang-de.png) | ![LAN Transfer în arabă, aspect de la dreapta la stânga](../docs/screenshots/lang-ar.png) |

## Capturi de ecran

| Asociere | Trimitere cu progres |
|---|---|
| ![Dialogul de asociere cu codul de 6 cifre și amprenta certificatului](../docs/screenshots/pairing.png) | ![Coada de transferuri cu progres, viteză și timp rămas](../docs/screenshots/sending.png) |
| **Reluare după anulare** | **Destinatar: istoric** |
| ![Un transfer anulat reluat de la poziția deja primită, verificat cu SHA-256](../docs/screenshots/resume.png) | ![Fereastra destinatarului cu elementele primite în istoric](../docs/screenshots/history.png) |

## Descărcare și pornire

1. Descărcați `LANTransfer-2.1.0-windows-amd64.zip` din [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) și verificați-l cu fișierul `.sha256` publicat.
2. Extrageți-l în orice folder pe ambele calculatoare.
3. Porniți `LANTransfer.exe`. La prima pornire Windows cere (UAC) permisiunea de a adăuga două reguli de firewall; vedeți [Porturi și firewall](#porturi-și-firewall).

Cerințe: Windows 10 sau 11 (64 de biți) cu runtime-ul WebView2, care face parte din versiunile actuale de Windows. Ambele calculatoare trebuie să ruleze versiunea 2. Versiunea 2 nu poate comunica cu versiunea 1.

## Cum se face asocierea

1. Porniți LAN Transfer pe ambele calculatoare. Destinatarul își afișează codul de asociere și amprenta certificatului în colțul din dreapta sus.
2. Pe calculatorul expeditor, faceți clic pe destinatar în lista **Destinatar**. Dacă nu apare, folosiți **Conectare manuală (adresă IP)** și introduceți adresa lui IP.
3. Tastați codul de 6 cifre de pe ecranul destinatarului și comparați amprenta certificatului afișată în dialog cu cea de pe destinatar. Dacă diferă, nu continuați.
4. Faceți clic pe **Asociere**. Dispozitivul afișează acum **✓ asociat**.

După 5 coduri greșite de la un calculator, destinatarul blochează încercările ulterioare de la acesta. **Cod nou** creează un cod nou și anulează toate asocierile existente. Destinatarul creează un cod și un certificat noi la fiecare pornire, așa că după repornirea destinatarului trebuie să faceți din nou asocierea.

## Trimiterea fișierelor

1. Asociați-vă cu destinatarul.
2. Trageți fișiere sau foldere în fereastră, folosiți **Alegeți fișiere** / **Alegeți un folder** sau lipiți o cale.
3. Alegeți destinația: folderul implicit al destinatarului (`Downloads\LAN Transfer`) sau un alt folder existent pe destinatar.
4. Faceți clic pe **Trimitere**. Fiecare element afișează progresul, viteza și timpul rămas și poate fi anulat. Un transfer anulat sau întrerupt poate fi reluat; se trimite doar partea lipsă.

## Linia de comandă

`LANTransferCLI.exe` este același program fără fereastră, pentru scripturi și automatizare.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Codul poate fi transmis și prin `LANTRANSFER_CODE`. Un destinatar fără fereastră poate folosi un cod fix din `LANTRANSFER_PIN`. Rularea din nou a aceleiași comenzi `send` reia un transfer întrerupt.

## Porturi și firewall

- UDP 45454: descoperire. TCP 45455: transfer (HTTPS).
- La prima pornire aplicația cere o singură dată (UAC) adăugarea a două reguli de intrare: `LAN Transfer 2 Discovery (UDP)` și `LAN Transfer 2 Transfer (TCP)`.
- Regulile acceptă conexiuni doar din subrețeaua locală (`remoteip=LocalSubnet`) și doar pe profilurile de rețea Privată și Domeniu (`profile=private,domain`). Într-o rețea pe care Windows o marchează ca Publică (cafenea, hotel, aeroport), alte calculatoare nu se pot conecta la LAN Transfer.
- Dacă destinatarul nu este vizibil în rețeaua de acasă, verificați ca rețeaua să fie setată ca Privată în setările de rețea Windows.
- Versiunea 1 a creat reguli numite `LAN Transfer Discovery` și `LAN Transfer Transfer` pentru toate profilurile, inclusiv Publică. Dacă versiunea 2 le găsește, afișează un banner care oferă înlocuirea lor. Nu se modifică nimic până când nu faceți clic pe **Actualizare reguli** și nu confirmați solicitarea Windows.

## Note de securitate

- Descoperirea prin UDP nu este autentificată; este doar un indiciu. Încrederea provine din asociere: codul plus amprenta fixată a certificatului.
- În timpul primei asocieri, un atacator activ din aceeași rețea (man in the middle) s-ar putea da drept destinatar. Comparați amprenta certificatului pe ambele ecrane înainte de asociere.
- Un expeditor asociat poate răsfoi folderele destinatarului și poate salva fișiere în orice folder existent de acolo (opțiunea de destinație personalizată). Asociați doar calculatoare în care aveți încredere și folosiți **Cod nou** pentru a revoca toate asocierile.
- Regulile de firewall nu deschid porturile în rețelele marcate ca Publice.
- Cu **Solicită codul de asociere** dezactivat, oricine din rețea care rulează LAN Transfer poate trimite fișiere pe acest computer și îi poate răsfoi folderele fără cod, ca în versiunea 1 (dar criptat). Folosiți-l doar într-o rețea de încredere. Reactivarea codului deconectează expeditorii conectați fără el.

## Variabile de mediu

Acestea sunt destinate testelor și automatizării; utilizarea normală nu necesită niciuna dintre ele.

| Variabilă | Semnificație |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | alte porturi |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | folderul de primire, folderul de date al programului |
| `LANTRANSFER_NAME` | numele calculatorului afișat celorlalți |
| `LANTRANSFER_LANG` | limba interfeței și a liniei de comandă (`en`, `de`, `pl`, …); are prioritate față de alegerea salvată |
| `LANTRANSFER_LOOPBACK=1` | doar 127.0.0.1: fără broadcast, fără scanarea LAN, fără reguli de firewall |
| `LANTRANSFER_PEER_UDP_PORTS` | porturile UDP ale altor instanțe în modul loopback |
| `LANTRANSFER_NO_FIREWALL=1` | nu creează și nu verifică reguli de firewall |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | codul fix al destinatarului, codul folosit de expeditor |
| `LANTRANSFER_REQUIRE_CODE=0` | `0` = acest destinatar nu cere cod de asociere, `1` = cod obligatoriu; are prioritate față de setare |

## Compilare și teste

Go 1.23 sau mai nou pe Windows 10/11.

```powershell
build.bat
go test ./...
python tools/check_i18n.py
```

`build.bat` rulează testele și compilează `LANTransfer.exe` (cu fereastră) și `LANTransferCLI.exe` (consolă) cu versiunea din `VERSION` (implicit `2.1.0`). Testele includ transferuri end-to-end prin 127.0.0.1: asociere și blocare, cereri neautentificate respinse, fișiere de la 0 B la 96 MB, foldere Unicode și căi de peste 260 de caractere, coliziuni de nume, nume ostile, o sumă de control greșită, anulare și reluare, disc plin și fixarea certificatului.

Traducerile se află în `locales/ui.<code>.json` (poloneza este limba sursă, engleza limba de rezervă). `tools/check_i18n.py` verifică dacă toate cele 30 de fișiere au aceleași chei, substituenți și forme de plural și dacă nu conțin resturi de text în poloneză; un test Go verifică dacă fiecare cheie folosită de program există în fiecare limbă.

Note de versiune: [docs/releases/v2.1.0.md](../docs/releases/v2.1.0.md) · [v2.0.0](../docs/releases/v2.0.0.md).

## Licență

MIT, vedeți [LICENSE](../LICENSE).

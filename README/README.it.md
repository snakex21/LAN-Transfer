**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#download-e-avvio)

Un piccolo strumento portatile per inviare file e cartelle direttamente tra due computer della stessa rete locale. Nessun cloud, nessun account e nessun server di inoltro: i dati passano direttamente da un PC all'altro tramite una connessione cifrata. Nucleo in Go, una finestra WebView2 e una versione a riga di comando per gli script.

![Finestra principale di LAN Transfer: un destinatario associato e tre elementi pronti per l'invio](../docs/screenshots/main.png)

## Indice

- [Funzionalità](#funzionalità)
- [Screenshot](#screenshot)
- [Download e avvio](#download-e-avvio)
- [Abbinamento](#abbinamento)
- [Invio di file](#invio-di-file)
- [Riga di comando](#riga-di-comando)
- [Porte e firewall](#porte-e-firewall)
- [Note sulla sicurezza](#note-sulla-sicurezza)
- [Variabili d'ambiente](#variabili-dambiente)
- [Compilazione e test](#compilazione-e-test)
- [Licenza](#licenza)

## Funzionalità

- Rilevamento automatico degli altri computer su cui è in esecuzione LAN Transfer (broadcast UDP, con una scansione a velocità limitata della rete /24 come ripiego).
- Abbinamento tramite un codice di 6 cifre mostrato sul destinatario. Senza abbinamento non è possibile inviare nulla né sfogliare le cartelle.
- Cifratura TLS 1.3. Ogni avvio crea un nuovo certificato autofirmato; la sua impronta viene fissata (pinning) durante l'abbinamento e verificata a ogni connessione.
- File e intere cartelle, con la struttura delle cartelle preservata e fino a 4 file inviati in parallelo.
- Ripresa dei trasferimenti interrotti e verifica SHA-256 di ogni file.
- I dati incompleti vengono conservati in una cartella nascosta `.lantransfer-partial`, così un file troncato non compare mai con il suo nome definitivo.
- In caso di nomi in conflitto si ottiene `file (1).txt` o `Folder (1)`; i file esistenti non vengono mai sovrascritti.
- I nomi pericolosi vengono rifiutati (`..`, percorsi assoluti, flussi di dati alternativi `file:stream`, nomi riservati come `CON`, `NUL`, `COM1`). Prima della ricezione viene controllato lo spazio libero su disco.
- Trascinamento da Esplora file, una coda con avanzamento per singolo file, velocità, tempo rimanente, annullamento, ripresa e una cronologia degli elementi inviati e ricevuti.
- Portatile: un unico EXE, con impostazioni e cronologia in una cartella `data` accanto a esso.

L'interfaccia utente è attualmente in polacco.

## Screenshot

| Abbinamento | Invio con avanzamento |
|---|---|
| ![Finestra di abbinamento con il codice di 6 cifre e l'impronta del certificato](../docs/screenshots/pairing.png) | ![Coda dei trasferimenti con avanzamento, velocità e tempo rimanente](../docs/screenshots/sending.png) |
| **Ripresa dopo l'annullamento** | **Destinatario: cronologia** |
| ![Un trasferimento annullato ripreso dal punto già ricevuto, verificato con SHA-256](../docs/screenshots/resume.png) | ![Finestra del destinatario con gli elementi ricevuti nella cronologia](../docs/screenshots/history.png) |

## Download e avvio

1. Scarica `LANTransfer-2.0.0-windows-amd64.zip` da [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) e verificalo con il file `.sha256` pubblicato.
2. Estrailo in una cartella qualsiasi su entrambi i computer.
3. Avvia `LANTransfer.exe`. Al primo avvio Windows chiede (UAC) l'autorizzazione ad aggiungere due regole del firewall; vedi [Porte e firewall](#porte-e-firewall).

Requisiti: Windows 10 o 11 (64 bit) con il runtime WebView2, incluso nelle versioni attuali di Windows. Su entrambi i computer deve essere in esecuzione la versione 2. La versione 2 non può comunicare con la versione 1.

## Abbinamento

1. Avvia LAN Transfer su entrambi i computer. Il destinatario mostra nell'angolo in alto a destra il proprio codice di abbinamento e l'impronta del certificato.
2. Sul computer mittente, fai clic sul destinatario nell'elenco **Odbiorca** (Destinatario). Se non compare, usa **Połącz ręcznie** (Connetti manualmente) e inserisci il suo indirizzo IP.
3. Digita il codice di 6 cifre mostrato sullo schermo del destinatario e confronta l'impronta del certificato visualizzata nella finestra di dialogo con quella sul destinatario. Se sono diverse, non proseguire.
4. Fai clic su **Sparuj** (Associa). Il dispositivo ora risulta **sparowano** (associato).

Dopo 5 codici errati da uno stesso computer, il destinatario blocca gli ulteriori tentativi provenienti da esso. **Nowy kod** (Nuovo codice) crea un nuovo codice e annulla tutti gli abbinamenti esistenti. Il destinatario crea un nuovo codice e un nuovo certificato a ogni avvio, quindi dopo il riavvio del destinatario occorre ripetere l'abbinamento.

## Invio di file

1. Esegui l'abbinamento con il destinatario.
2. Trascina file o cartelle sulla finestra, usa **Wybierz pliki** / **Wybierz folder** (Scegli file / Scegli cartella) oppure incolla un percorso.
3. Scegli la destinazione: la cartella predefinita del destinatario (`Downloads\LAN Transfer`) o un'altra cartella esistente sul destinatario.
4. Fai clic su **Wyślij** (Invia). Ogni elemento mostra avanzamento, velocità e tempo rimanente e può essere annullato. Un trasferimento annullato o interrotto può essere ripreso; viene inviata solo la parte mancante.

## Riga di comando

`LANTransferCLI.exe` è lo stesso programma senza finestra, per script e automazione.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Il codice può anche essere passato in `LANTRANSFER_CODE`. Un destinatario senza finestra può usare un codice fisso impostato in `LANTRANSFER_PIN`. Eseguire di nuovo lo stesso comando `send` riprende un trasferimento interrotto.

## Porte e firewall

- UDP 45454: rilevamento. TCP 45455: trasferimento (HTTPS).
- Al primo avvio l'app chiede una sola volta (UAC) di aggiungere due regole in ingresso: `LAN Transfer 2 Discovery (UDP)` e `LAN Transfer 2 Transfer (TCP)`.
- Le regole accettano connessioni solo dalla subnet locale (`remoteip=LocalSubnet`) e solo sui profili di rete Privato e Dominio (`profile=private,domain`). Su una rete che Windows classifica come Pubblica (bar, hotel, aeroporto) gli altri computer non possono connettersi a LAN Transfer.
- Se il destinatario non è visibile nella tua rete domestica, verifica nelle impostazioni di rete di Windows che la rete sia impostata come Privata.
- La versione 1 creava regole denominate `LAN Transfer Discovery` e `LAN Transfer Transfer` per tutti i profili, compreso quello Pubblico. Se la versione 2 le trova, mostra un banner che propone di sostituirle. Non viene modificato nulla finché non fai clic su **Zaktualizuj reguły** (Aggiorna regole) e confermi la richiesta di Windows.

## Note sulla sicurezza

- Il rilevamento tramite UDP non è autenticato; è solo un'indicazione. La fiducia deriva dall'abbinamento: il codice più l'impronta del certificato fissata.
- Durante il primo abbinamento, un attaccante attivo sulla stessa rete (man in the middle) potrebbe spacciarsi per il destinatario. Prima di associare, confronta l'impronta del certificato su entrambi gli schermi.
- Un mittente associato può sfogliare le cartelle del destinatario e salvare file in qualsiasi cartella esistente su di esso (l'opzione di destinazione personalizzata). Associa solo computer di cui ti fidi e usa **Nowy kod** per revocare tutti gli abbinamenti.
- Le regole del firewall non aprono le porte sulle reti classificate come Pubbliche.

## Variabili d'ambiente

Sono pensate per test e automazione; l'uso normale non ne richiede nessuna.

| Variabile | Significato |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | porte diverse |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | cartella di ricezione, cartella dei dati del programma |
| `LANTRANSFER_NAME` | nome del computer mostrato agli altri |
| `LANTRANSFER_LOOPBACK=1` | solo 127.0.0.1: niente broadcast, niente scansione della LAN, niente regole del firewall |
| `LANTRANSFER_PEER_UDP_PORTS` | porte UDP di altre istanze in modalità loopback |
| `LANTRANSFER_NO_FIREWALL=1` | non creare né controllare le regole del firewall |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | codice fisso del destinatario, codice usato dal mittente |

## Compilazione e test

Go 1.23 o successivo su Windows 10/11.

```powershell
build.bat
go test ./...
```

`build.bat` esegue i test e compila `LANTransfer.exe` (finestra) e `LANTransferCLI.exe` (console) con la versione indicata in `VERSION` (predefinita `2.0.0`). I test comprendono trasferimenti end-to-end su 127.0.0.1: abbinamento e blocco, rifiuto delle richieste non autenticate, file da 0 B a 96 MB, cartelle Unicode e percorsi oltre 260 caratteri, conflitti di nomi, nomi malevoli, un checksum errato, annullamento e ripresa, un disco pieno e il pinning del certificato.

Note di rilascio: [docs/releases/v2.0.0.md](../docs/releases/v2.0.0.md).

## Licenza

MIT, vedi [LICENSE](../LICENSE).

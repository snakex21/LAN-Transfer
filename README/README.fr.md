**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#téléchargement-et-lancement)

Un petit outil portable pour envoyer des fichiers et des dossiers directement entre deux ordinateurs du même réseau local. Pas de cloud, pas de compte ni de serveur relais : les données passent directement d'un PC à l'autre via une connexion chiffrée. Cœur en Go, une fenêtre WebView2 et une version en ligne de commande pour les scripts.

![Fenêtre principale de LAN Transfer : un destinataire appairé et trois éléments prêts à être envoyés](../docs/screenshots/main.png)

## Sommaire

- [Fonctionnalités](#fonctionnalités)
- [Captures d'écran](#captures-décran)
- [Téléchargement et lancement](#téléchargement-et-lancement)
- [Comment appairer](#comment-appairer)
- [Envoi de fichiers](#envoi-de-fichiers)
- [Ligne de commande](#ligne-de-commande)
- [Ports et pare-feu](#ports-et-pare-feu)
- [Remarques de sécurité](#remarques-de-sécurité)
- [Variables d'environnement](#variables-denvironnement)
- [Compilation et tests](#compilation-et-tests)
- [Licence](#licence)

## Fonctionnalités

- Découverte automatique des autres ordinateurs exécutant LAN Transfer (diffusion UDP, avec en secours un balayage limité en débit du réseau /24).
- Appairage avec un code à 6 chiffres affiché sur le destinataire. Sans appairage, rien ne peut être envoyé et aucun dossier ne peut être parcouru.
- Chiffrement TLS 1.3. Chaque lancement crée un nouveau certificat auto-signé ; son empreinte est épinglée lors de l'appairage et vérifiée à chaque connexion.
- Fichiers et dossiers entiers, avec conservation de l'arborescence et envoi de jusqu'à 4 fichiers en parallèle.
- Reprise des transferts interrompus et vérification SHA-256 de chaque fichier.
- Les données incomplètes sont conservées dans un dossier caché `.lantransfer-partial`, de sorte qu'un fichier tronqué n'apparaît jamais sous son nom définitif.
- En cas de conflit de noms, on obtient `file (1).txt` ou `Folder (1)` ; les fichiers existants ne sont jamais écrasés.
- Les noms dangereux sont refusés (`..`, chemins absolus, flux de données alternatifs `file:stream`, noms réservés tels que `CON`, `NUL`, `COM1`). L'espace disque libre est vérifié avant la réception.
- Glisser-déposer depuis l'Explorateur de fichiers, file d'attente avec progression par fichier, vitesse, temps restant, annulation, reprise et historique des éléments envoyés et reçus.
- Portable : un seul fichier EXE, paramètres et historique dans un dossier `data` à côté.

L'interface utilisateur est actuellement en polonais.

## Captures d'écran

| Appairage | Envoi avec progression |
|---|---|
| ![Boîte de dialogue d'appairage avec le code à 6 chiffres et l'empreinte du certificat](../docs/screenshots/pairing.png) | ![File de transferts avec progression, vitesse et temps restant](../docs/screenshots/sending.png) |
| **Reprise après annulation** | **Destinataire : historique** |
| ![Un transfert annulé repris à partir de la position déjà reçue, vérifié par SHA-256](../docs/screenshots/resume.png) | ![Fenêtre du destinataire avec les éléments reçus dans l'historique](../docs/screenshots/history.png) |

## Téléchargement et lancement

1. Téléchargez `LANTransfer-2.0.0-windows-amd64.zip` depuis [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) et vérifiez-le à l'aide du fichier `.sha256` publié.
2. Extrayez-le dans un dossier quelconque sur les deux ordinateurs.
3. Lancez `LANTransfer.exe`. Au premier lancement, Windows demande (UAC) l'autorisation d'ajouter deux règles de pare-feu ; voir [Ports et pare-feu](#ports-et-pare-feu).

Configuration requise : Windows 10 ou 11 (64 bits) avec le runtime WebView2, inclus dans les versions actuelles de Windows. Les deux ordinateurs doivent exécuter la version 2. La version 2 ne peut pas communiquer avec la version 1.

## Comment appairer

1. Lancez LAN Transfer sur les deux ordinateurs. Le destinataire affiche son code d'appairage et l'empreinte de son certificat dans le coin supérieur droit.
2. Sur l'ordinateur émetteur, cliquez sur le destinataire dans la liste **Odbiorca** (Destinataire). S'il n'apparaît pas, utilisez **Połącz ręcznie** (Connexion manuelle) et saisissez son adresse IP.
3. Saisissez le code à 6 chiffres affiché sur l'écran du destinataire et comparez l'empreinte du certificat affichée dans la boîte de dialogue avec celle du destinataire. Si elles diffèrent, n'allez pas plus loin.
4. Cliquez sur **Sparuj** (Appairer). L'appareil affiche désormais **sparowano** (appairé).

Après 5 codes erronés provenant d'un même ordinateur, le destinataire bloque toute nouvelle tentative de sa part. **Nowy kod** (Nouveau code) génère un nouveau code et annule tous les appairages existants. Le destinataire crée un nouveau code et un nouveau certificat à chaque démarrage : après un redémarrage du destinataire, il faut donc appairer à nouveau.

## Envoi de fichiers

1. Appairez-vous avec le destinataire.
2. Faites glisser des fichiers ou des dossiers sur la fenêtre, utilisez **Wybierz pliki** / **Wybierz folder** (Choisir des fichiers / Choisir un dossier), ou collez un chemin.
3. Choisissez la destination : le dossier par défaut du destinataire (`Downloads\LAN Transfer`) ou un autre dossier existant sur le destinataire.
4. Cliquez sur **Wyślij** (Envoyer). Chaque élément affiche la progression, la vitesse et le temps restant, et peut être annulé. Un transfert annulé ou interrompu peut être repris ; seule la partie manquante est envoyée.

## Ligne de commande

`LANTransferCLI.exe` est le même programme sans fenêtre, destiné aux scripts et à l'automatisation.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Le code peut aussi être transmis via `LANTRANSFER_CODE`. Un destinataire sans fenêtre peut utiliser un code fixe défini dans `LANTRANSFER_PIN`. Relancer la même commande `send` reprend un transfert interrompu.

## Ports et pare-feu

- UDP 45454 : découverte. TCP 45455 : transfert (HTTPS).
- Au premier lancement, l'application demande une seule fois (UAC) d'ajouter deux règles de trafic entrant : `LAN Transfer 2 Discovery (UDP)` et `LAN Transfer 2 Transfer (TCP)`.
- Les règles n'acceptent les connexions que depuis le sous-réseau local (`remoteip=LocalSubnet`) et uniquement sur les profils réseau Privé et Domaine (`profile=private,domain`). Sur un réseau que Windows classe comme Public (café, hôtel, aéroport), les autres ordinateurs ne peuvent pas se connecter à LAN Transfer.
- Si le destinataire n'est pas visible sur votre réseau domestique, vérifiez dans les paramètres réseau de Windows que le réseau est défini comme Privé.
- La version 1 créait des règles nommées `LAN Transfer Discovery` et `LAN Transfer Transfer` pour tous les profils, y compris Public. Si la version 2 les détecte, elle affiche un bandeau proposant de les remplacer. Rien n'est modifié tant que vous n'avez pas cliqué sur **Zaktualizuj reguły** (Mettre à jour les règles) et confirmé l'invite de Windows.

## Remarques de sécurité

- La découverte via UDP n'est pas authentifiée ; ce n'est qu'une indication. La confiance repose sur l'appairage : le code et l'empreinte de certificat épinglée.
- Lors du premier appairage, un attaquant actif présent sur le même réseau (attaque de l'homme du milieu) pourrait se faire passer pour le destinataire. Comparez l'empreinte du certificat sur les deux écrans avant d'appairer.
- Un émetteur appairé peut parcourir les dossiers du destinataire et enregistrer des fichiers dans n'importe quel dossier existant sur celui-ci (option de destination personnalisée). N'appairez que des ordinateurs de confiance et utilisez **Nowy kod** pour révoquer tous les appairages.
- Les règles de pare-feu n'ouvrent pas les ports sur les réseaux classés comme Publics.

## Variables d'environnement

Elles sont destinées aux tests et à l'automatisation ; un usage normal n'en nécessite aucune.

| Variable | Signification |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | autres ports |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | dossier de réception, dossier de données du programme |
| `LANTRANSFER_NAME` | nom de l'ordinateur affiché aux autres |
| `LANTRANSFER_LOOPBACK=1` | 127.0.0.1 uniquement : pas de diffusion, pas de balayage du LAN, pas de règles de pare-feu |
| `LANTRANSFER_PEER_UDP_PORTS` | ports UDP des autres instances en mode loopback |
| `LANTRANSFER_NO_FIREWALL=1` | ne pas créer ni vérifier les règles de pare-feu |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | code fixe du destinataire, code utilisé par l'émetteur |

## Compilation et tests

Go 1.23 ou plus récent sous Windows 10/11.

```powershell
build.bat
go test ./...
```

`build.bat` exécute les tests et compile `LANTransfer.exe` (fenêtre) et `LANTransferCLI.exe` (console) avec la version indiquée dans `VERSION` (par défaut `2.0.0`). Les tests comprennent des transferts de bout en bout via 127.0.0.1 : appairage et blocage, rejet des requêtes non authentifiées, fichiers de 0 o à 96 Mo, dossiers Unicode et chemins de plus de 260 caractères, conflits de noms, noms malveillants, somme de contrôle incorrecte, annulation et reprise, disque plein et épinglage de certificat.

Notes de version : [docs/releases/v2.0.0.md](../docs/releases/v2.0.0.md).

## Licence

MIT, voir [LICENSE](../LICENSE).

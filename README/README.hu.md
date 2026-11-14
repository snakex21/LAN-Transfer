**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#letöltés-és-futtatás)

Kis méretű, hordozható eszköz fájlok és mappák közvetlen küldésére két, ugyanazon a helyi hálózaton lévő számítógép között. Nincs felhő, nincs fiók és nincs közvetítő szerver: az adatok titkosított kapcsolaton keresztül egyenesen az egyik gépről a másikra jutnak. Go nyelvű mag, WebView2 ablak és szkriptekhez használható parancssori változat.

![A LAN Transfer főablaka: egy párosított fogadó és három küldésre kész elem](../docs/screenshots/main.png)

## Tartalom

- [Funkciók](#funkciók)
- [Képernyőképek](#képernyőképek)
- [Letöltés és futtatás](#letöltés-és-futtatás)
- [A párosítás menete](#a-párosítás-menete)
- [Fájlok küldése](#fájlok-küldése)
- [Parancssor](#parancssor)
- [Portok és tűzfal](#portok-és-tűzfal)
- [Biztonsági megjegyzések](#biztonsági-megjegyzések)
- [Környezeti változók](#környezeti-változók)
- [Fordítás és tesztek](#fordítás-és-tesztek)
- [Licenc](#licenc)

## Funkciók

- A LAN Transfert futtató többi számítógép automatikus felderítése (UDP broadcast, tartalékként lassított /24-es pásztázással).
- Párosítás a fogadón megjelenő 6 jegyű kóddal. Párosítás nélkül semmit sem lehet küldeni, és mappákat sem lehet böngészni.
- TLS 1.3 titkosítás. Minden indításkor új, önaláírt tanúsítvány jön létre; ennek ujjlenyomatát a párosítás rögzíti, és minden kapcsolatnál ellenőrzi.
- Fájlok és teljes mappák a mappaszerkezet megőrzésével, egyszerre legfeljebb 4 fájl párhuzamos küldésével.
- A megszakadt átvitelek folytatása és minden fájl SHA-256 ellenőrzése.
- A befejezetlen adatok egy rejtett `.lantransfer-partial` mappába kerülnek, így csonka fájl soha nem jelenik meg a végleges nevén.
- Névütközés esetén `file (1).txt` vagy `Folder (1)` jön létre; a meglévő fájlok soha nem íródnak felül.
- A veszélyes nevek elutasításra kerülnek (`..`, abszolút elérési utak, `file:stream` alternatív adatfolyamok, fenntartott nevek, például `CON`, `NUL`, `COM1`). Fogadás előtt a program ellenőrzi a szabad lemezterületet.
- Húzás és ejtés a Fájlkezelőből, sor fájlonkénti folyamatjelzővel, sebességgel, hátralévő idővel, megszakítással, folytatással, valamint az elküldött és fogadott elemek előzményei.
- Hordozható: egyetlen EXE, a beállítások és az előzmények a mellette lévő `data` mappában.

A felhasználói felület jelenleg lengyel nyelvű.

## Képernyőképek

| Párosítás | Küldés folyamatjelzővel |
|---|---|
| ![Párosítási párbeszédablak a 6 jegyű kóddal és a tanúsítvány ujjlenyomatával](../docs/screenshots/pairing.png) | ![Átviteli sor folyamatjelzővel, sebességgel és hátralévő idővel](../docs/screenshots/sending.png) |
| **Folytatás megszakítás után** | **Fogadó: előzmények** |
| ![Megszakított átvitel folytatása a már fogadott pozíciótól, SHA-256-tal ellenőrizve](../docs/screenshots/resume.png) | ![A fogadó ablaka a fogadott elemekkel az előzményekben](../docs/screenshots/history.png) |

## Letöltés és futtatás

1. Töltse le a `LANTransfer-2.0.0-windows-amd64.zip` fájlt a [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) oldalról, és ellenőrizze a közzétett `.sha256` fájllal.
2. Csomagolja ki tetszőleges mappába mindkét számítógépen.
3. Indítsa el a `LANTransfer.exe` fájlt. Az első indításkor a Windows (UAC) engedélyt kér két tűzfalszabály hozzáadásához; lásd: [Portok és tűzfal](#portok-és-tűzfal).

Követelmények: Windows 10 vagy 11 (64 bites) a WebView2 futtatókörnyezettel, amely a jelenlegi Windows része. Mindkét számítógépen a 2-es verziónak kell futnia. A 2-es verzió nem tud kommunikálni az 1-es verzióval.

## A párosítás menete

1. Indítsa el a LAN Transfert mindkét számítógépen. A fogadó a jobb felső sarokban mutatja a párosítási kódját és a tanúsítványa ujjlenyomatát.
2. A küldő számítógépen kattintson a fogadóra az **Odbiorca** (Fogadó) listában. Ha nem jelenik meg, használja a **Połącz ręcznie** (Kézi csatlakozás) lehetőséget, és adja meg az IP-címét.
3. Írja be a fogadó képernyőjén látható 6 jegyű kódot, és vesse össze a párbeszédablakban megjelenő tanúsítvány-ujjlenyomatot a fogadón látható ujjlenyomattal. Ha eltérnek, ne folytassa.
4. Kattintson a **Sparuj** (Párosítás) gombra. Az eszköz ekkor a **sparowano** (párosítva) állapotot mutatja.

Ha egy számítógépről 5 hibás kód érkezik, a fogadó letiltja az onnan érkező további próbálkozásokat. A **Nowy kod** (Új kód) új kódot hoz létre, és megszünteti az összes meglévő párosítást. A fogadó minden indításkor új kódot és tanúsítványt hoz létre, ezért a fogadó újraindítása után ismét párosítani kell.

## Fájlok küldése

1. Párosítsa a gépet a fogadóval.
2. Húzza a fájlokat vagy mappákat az ablakba, használja a **Wybierz pliki** / **Wybierz folder** (Fájlok kiválasztása / Mappa kiválasztása) gombot, vagy illesszen be egy elérési utat.
3. Válassza ki a célt: a fogadó alapértelmezett mappáját (`Downloads\LAN Transfer`) vagy egy másik, a fogadón már létező mappát.
4. Kattintson a **Wyślij** (Küldés) gombra. Minden elemnél látható a folyamat, a sebesség és a hátralévő idő, és az elem megszakítható. A megszakított vagy félbeszakadt átvitel folytatható; ilyenkor csak a hiányzó rész kerül elküldésre.

## Parancssor

A `LANTransferCLI.exe` ugyanaz a program ablak nélkül, szkriptekhez és automatizáláshoz.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

A kód a `LANTRANSFER_CODE` változóban is átadható. Az ablak nélküli fogadó a `LANTRANSFER_PIN` változóból vett rögzített kódot is használhat. Ugyanazon `send` parancs ismételt futtatása folytatja a megszakadt átvitelt.

## Portok és tűzfal

- UDP 45454: felderítés. TCP 45455: átvitel (HTTPS).
- Az első indításkor az alkalmazás egyszer engedélyt kér (UAC) két bejövő szabály hozzáadásához: `LAN Transfer 2 Discovery (UDP)` és `LAN Transfer 2 Transfer (TCP)`.
- A szabályok csak a helyi alhálózatról (`remoteip=LocalSubnet`) és csak a Magánhálózat és Tartomány hálózati profilokon (`profile=private,domain`) fogadnak kapcsolatokat. Olyan hálózaton, amelyet a Windows Nyilvánosként jelöl meg (kávézó, szálloda, repülőtér), más számítógépek nem tudnak csatlakozni a LAN Transferhez.
- Ha a fogadó nem látható az otthoni hálózaton, ellenőrizze, hogy a Windows hálózati beállításaiban a hálózat Magánhálózatként van-e beállítva.
- Az 1-es verzió `LAN Transfer Discovery` és `LAN Transfer Transfer` nevű szabályokat hozott létre minden profilra, a Nyilvánost is beleértve. Ha a 2-es verzió megtalálja őket, egy sávban felajánlja a lecserélésüket. Semmi sem változik, amíg nem kattint a **Zaktualizuj reguły** (Szabályok frissítése) gombra, és nem erősíti meg a Windows kérdését.

## Biztonsági megjegyzések

- Az UDP-n keresztüli felderítés nincs hitelesítve; csupán tájékoztató jellegű. A bizalom alapja a párosítás: a kód és a rögzített tanúsítvány-ujjlenyomat.
- Az első párosítás során egy ugyanazon a hálózaton lévő aktív támadó (man in the middle) kiadhatja magát a fogadónak. Párosítás előtt vesse össze a tanúsítvány ujjlenyomatát mindkét képernyőn.
- Egy párosított küldő böngészheti a fogadó mappáit, és bármelyik ott létező mappába menthet fájlokat (egyéni cél beállítás). Csak megbízható számítógépeket párosítson, és az összes párosítás visszavonásához használja a **Nowy kod** gombot.
- A tűzfalszabályok nem nyitják meg a portokat a Nyilvánosként megjelölt hálózatokon.

## Környezeti változók

Ezek tesztekhez és automatizáláshoz készültek; normál használathoz egyikre sincs szükség.

| Változó | Jelentés |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | eltérő portok |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | fogadási mappa, a program adatmappája |
| `LANTRANSFER_NAME` | a többiek számára megjelenített számítógépnév |
| `LANTRANSFER_LOOPBACK=1` | csak 127.0.0.1: nincs broadcast, nincs LAN-pásztázás, nincsenek tűzfalszabályok |
| `LANTRANSFER_PEER_UDP_PORTS` | más példányok UDP-portjai loopback módban |
| `LANTRANSFER_NO_FIREWALL=1` | ne hozzon létre és ne ellenőrizzen tűzfalszabályokat |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | a fogadó rögzített kódja, a küldő által használt kód |

## Fordítás és tesztek

Go 1.23 vagy újabb, Windows 10/11 rendszeren.

```powershell
build.bat
go test ./...
```

A `build.bat` lefuttatja a teszteket, és elkészíti a `LANTransfer.exe` (ablakos) és a `LANTransferCLI.exe` (konzolos) programot a `VERSION` fájlban megadott verzióval (alapértelmezés: `2.0.0`). A tesztek között 127.0.0.1-en keresztüli végpontok közötti átvitelek szerepelnek: párosítás és letiltás, elutasított hitelesítetlen kérések, 0 B és 96 MB közötti fájlok, Unicode mappák és 260 karakternél hosszabb elérési utak, névütközések, rosszindulatú nevek, hibás ellenőrzőösszeg, megszakítás és folytatás, megtelt lemez és tanúsítvány-rögzítés.

Kiadási megjegyzések: [docs/releases/v2.0.0.md](../docs/releases/v2.0.0.md).

## Licenc

MIT, lásd: [LICENSE](../LICENSE).

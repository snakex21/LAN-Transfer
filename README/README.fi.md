**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#lataus-ja-käynnistys)

Pieni siirrettävä työkalu tiedostojen ja kansioiden lähettämiseen suoraan kahden saman lähiverkon tietokoneen välillä. Ei pilveä, ei tiliä eikä välityspalvelinta: tiedot kulkevat suoraan koneelta toiselle salatun yhteyden kautta. Go-ydin, WebView2-ikkuna ja komentoriviversio skriptejä varten.

![LAN Transferin pääikkuna: pariliitetty vastaanottaja ja kolme lähetysvalmista kohdetta](../docs/screenshots/main.png)

## Sisältö

- [Ominaisuudet](#ominaisuudet)
- [Käyttöliittymän kieli](#käyttöliittymän-kieli)
- [Kuvakaappaukset](#kuvakaappaukset)
- [Lataus ja käynnistys](#lataus-ja-käynnistys)
- [Pariliitoksen muodostaminen](#pariliitoksen-muodostaminen)
- [Tiedostojen lähettäminen](#tiedostojen-lähettäminen)
- [Komentorivi](#komentorivi)
- [Portit ja palomuuri](#portit-ja-palomuuri)
- [Tietoturvahuomautukset](#tietoturvahuomautukset)
- [Ympäristömuuttujat](#ympäristömuuttujat)
- [Kääntäminen ja testit](#kääntäminen-ja-testit)
- [Lisenssi](#lisenssi)

## Ominaisuudet

- Muiden LAN Transferia käyttävien tietokoneiden automaattinen löytäminen (UDP-yleislähetys, varalla rajoitettu /24-skannaus).
- Pariliitos vastaanottajan näyttämällä 6-numeroisella koodilla. Ilman pariliitosta mitään ei voi lähettää eikä kansioita voi selata.
- Valinnainen tila ilman koodia: poista vastaanottajan asetuksista **Vaadi pariliitoskoodi** (tallennetaan tiedostoon `data\settings.json`) tai käynnistä `serve --no-code`. Silloin mikä tahansa verkon LAN Transfer -tietokone voi lähettää sille ilman koodia, edelleen salattuna; otsikossa näkyy silloin **Ei koodia**.
- TLS 1.3 -salaus. Jokainen käynnistyskerta luo uuden itse allekirjoitetun varmenteen; sen sormenjälki kiinnitetään (pinning) pariliitoksen aikana ja tarkistetaan jokaisen yhteyden yhteydessä.
- Tiedostot ja kokonaiset kansiot kansiorakenne säilyttäen, enintään 4 tiedostoa lähetetään rinnakkain.
- Keskeytyneiden siirtojen jatkaminen ja jokaisen tiedoston SHA-256-tarkistus.
- Keskeneräiset tiedot säilytetään piilotetussa `.lantransfer-partial`-kansiossa, joten katkennut tiedosto ei koskaan näy lopullisella nimellään.
- Nimiristiriidoista syntyy `file (1).txt` tai `Folder (1)`; olemassa olevia tiedostoja ei koskaan korvata.
- Vaaralliset nimet hylätään (`..`, absoluuttiset polut, vaihtoehtoiset tietovirrat `file:stream`, varatut nimet kuten `CON`, `NUL`, `COM1`). Vapaa levytila tarkistetaan ennen vastaanottamista.
- Vedä ja pudota Resurssienhallinnasta, jono, jossa näkyy tiedostokohtainen edistyminen, nopeus ja jäljellä oleva aika, peruutus, jatkaminen sekä lähetettyjen ja vastaanotettujen kohteiden historia.
- Siirrettävä: yksi EXE-tiedosto, asetukset ja historia sen vieressä olevassa `data`-kansiossa.
- Ikkuna, sen otsikko ja komentorivin tulosteet 30 kielellä (samat kuin tällä README-tiedostolla); kieli valitaan automaattisesti Windowsin näyttökielen mukaan, ja arabiassa ja hepreassa käytetään oikealta vasemmalle -asettelua.

## Käyttöliittymän kieli

LAN Transfer on saatavilla samoilla 30 kielellä kuin tämä README-tiedosto. Ensimmäisellä käynnistyskerralla se käyttää Windowsin näyttökieltä (ensimmäistä tuettua kieltä Windowsin kieliluettelossa); jos Windowsin kieli ei ole luettelossa, käyttöliittymä on englanninkielinen.

Voit vaihtaa kielen napsauttamalla oikean yläkulman rataspainiketta (**Asetukset**) ja valitsemalla kielen. Ikkuna, sen otsikko ja kaikki viestit vaihtuvat heti, ja valinta tallennetaan tiedostoon `data\settings.json`. **Automaattinen (Windows)** palauttaa kielen seuraamaan Windowsin näyttökieltä. Luvut, koot, nopeudet, päivämäärät ja monikkomuodot noudattavat valitun kielen sääntöjä; arabiassa ja hepreassa käytetään peilattua oikealta vasemmalle -asettelua, mutta polut, IP-osoitteet, koodit ja sormenjäljet näytetään edelleen vasemmalta oikealle. Toisen tietokoneen viestit (esimerkiksi ”vastaanottajan levyllä ei ole tarpeeksi tilaa”) näytetään sinun kielelläsi eikä vastaanottajan kielellä.

`LANTransferCLI.exe` käyttää samaa kieltä. `LANTRANSFER_LANG` (esimerkiksi `LANTRANSFER_LANG=de`) ohittaa tallennetun valinnan sekä ikkunassa että komentorivillä.

| English | Deutsch | العربية (RTL) |
|---|---|---|
| ![LAN Transfer englanniksi](../docs/screenshots/lang-en.png) | ![LAN Transfer saksaksi](../docs/screenshots/lang-de.png) | ![LAN Transfer arabiaksi, oikealta vasemmalle -asettelu](../docs/screenshots/lang-ar.png) |

## Kuvakaappaukset

| Pariliitos | Lähetys ja edistyminen |
|---|---|
| ![Pariliitosikkuna, jossa näkyvät 6-numeroinen koodi ja varmenteen sormenjälki](../docs/screenshots/pairing.png) | ![Siirtojono, jossa näkyvät edistyminen, nopeus ja jäljellä oleva aika](../docs/screenshots/sending.png) |
| **Jatkaminen peruutuksen jälkeen** | **Vastaanottaja: historia** |
| ![Peruutettu siirto jatkettu vastaanotetusta kohdasta, SHA-256 vahvistettu](../docs/screenshots/resume.png) | ![Vastaanottajan ikkuna, jonka historiassa on vastaanotettuja kohteita](../docs/screenshots/history.png) |

## Lataus ja käynnistys

1. Lataa `LANTransfer-2.1.0-windows-amd64.zip` sivulta [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) ja tarkista se julkaistua `.sha256`-tiedostoa vasten.
2. Pura se mihin tahansa kansioon molemmilla tietokoneilla.
3. Käynnistä `LANTransfer.exe`. Ensimmäisellä käynnistyskerralla Windows pyytää (UAC) lupaa lisätä kaksi palomuurisääntöä; katso [Portit ja palomuuri](#portit-ja-palomuuri).

Vaatimukset: Windows 10 tai 11 (64-bittinen) ja WebView2-suorituspalvelu, joka sisältyy nykyisiin Windows-versioihin. Molemmissa tietokoneissa on oltava versio 2. Versio 2 ei pysty viestimään version 1 kanssa.

## Pariliitoksen muodostaminen

1. Käynnistä LAN Transfer molemmilla tietokoneilla. Vastaanottaja näyttää pariliitoskoodinsa ja varmenteensa sormenjäljen oikeassa yläkulmassa.
2. Napsauta lähettävällä tietokoneella vastaanottajaa **Vastaanottaja**-luettelossa. Jos sitä ei näy, käytä toimintoa **Muodosta yhteys manuaalisesti (IP-osoite)** ja anna sen IP-osoite.
3. Kirjoita vastaanottajan näytöllä näkyvä 6-numeroinen koodi ja vertaa valintaikkunassa näkyvää varmenteen sormenjälkeä vastaanottajan näyttämään. Jos ne eroavat, älä jatka.
4. Napsauta **Muodosta pariliitos**. Laitteen kohdalla näkyy nyt **✓ pariliitetty**.

Kun yhdeltä tietokoneelta on syötetty 5 väärää koodia, vastaanottaja estää siltä uudet yritykset. **Uusi koodi** luo uuden koodin ja peruuttaa kaikki olemassa olevat pariliitokset. Vastaanottaja luo uuden koodin ja varmenteen joka käynnistyksellä, joten muodosta pariliitos uudelleen vastaanottajan uudelleenkäynnistyksen jälkeen.

## Tiedostojen lähettäminen

1. Muodosta pariliitos vastaanottajan kanssa.
2. Vedä tiedostoja tai kansioita ikkunaan, käytä painikkeita **Valitse tiedostot** / **Valitse kansio** tai liitä polku.
3. Valitse kohde: vastaanottajan oletuskansio (`Downloads\LAN Transfer`) tai jokin muu vastaanottajan olemassa oleva kansio.
4. Napsauta **Lähetä**. Jokaisen kohteen edistyminen, nopeus ja jäljellä oleva aika näytetään, ja sen voi peruuttaa. Peruutettua tai keskeytynyttä siirtoa voi jatkaa; vain puuttuva osa lähetetään.

## Komentorivi

`LANTransferCLI.exe` on sama ohjelma ilman ikkunaa skriptejä ja automaatiota varten.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

Koodin voi antaa myös muuttujassa `LANTRANSFER_CODE`. Ilman ikkunaa toimiva vastaanottaja voi käyttää kiinteää koodia muuttujasta `LANTRANSFER_PIN`. Saman `send`-komennon suorittaminen uudelleen jatkaa keskeytynyttä siirtoa.

## Portit ja palomuuri

- UDP 45454: löytäminen. TCP 45455: siirto (HTTPS).
- Ensimmäisellä käynnistyskerralla sovellus pyytää kerran (UAC) lupaa lisätä kaksi saapuvan liikenteen sääntöä: `LAN Transfer 2 Discovery (UDP)` ja `LAN Transfer 2 Transfer (TCP)`.
- Säännöt hyväksyvät yhteyksiä vain paikallisesta aliverkosta (`remoteip=LocalSubnet`) ja vain verkkoprofiileissa Yksityinen ja Toimialue (`profile=private,domain`). Verkossa, jonka Windows merkitsee julkiseksi (kahvila, hotelli, lentoasema), muut tietokoneet eivät voi muodostaa yhteyttä LAN Transferiin.
- Jos vastaanottaja ei näy kotiverkossasi, tarkista Windowsin verkkoasetuksista, että verkon tyypiksi on asetettu Yksityinen.
- Versio 1 loi säännöt nimeltä `LAN Transfer Discovery` ja `LAN Transfer Transfer` kaikille profiileille, myös Julkinen-profiilille. Jos versio 2 löytää ne, se näyttää bannerin, jossa tarjotaan niiden korvaamista. Mitään ei muuteta ennen kuin napsautat **Päivitä säännöt** ja vahvistat Windowsin kehotteen.

## Tietoturvahuomautukset

- UDP:n kautta tapahtuvaa löytämistä ei todenneta; se on vain vihje. Luottamus perustuu pariliitokseen: koodiin ja kiinnitettyyn varmenteen sormenjälkeen.
- Ensimmäisen pariliitoksen aikana samassa verkossa oleva aktiivinen hyökkääjä (man-in-the-middle) voisi esiintyä vastaanottajana. Vertaa varmenteen sormenjälkeä molemmilla näytöillä ennen pariliitoksen muodostamista.
- Pariliitetty lähettäjä voi selata vastaanottajan kansioita ja tallentaa tiedostoja mihin tahansa siellä olemassa olevaan kansioon (mukautetun kohteen asetus). Muodosta pariliitos vain luotettujen tietokoneiden kanssa ja peruuta kaikki pariliitokset toiminnolla **Uusi koodi**.
- Palomuurisäännöt eivät avaa portteja verkoissa, jotka on merkitty julkisiksi.
- Kun **Vaadi pariliitoskoodi** on pois päältä, kuka tahansa verkossa LAN Transferia käyttävä voi lähettää tiedostoja tähän tietokoneeseen ja selata sen kansioita ilman koodia, kuten versiossa 1 (mutta salattuna). Käytä tätä vain luotetussa verkossa. Koodin ottaminen uudelleen käyttöön katkaisee ilman koodia yhdistetyt lähettäjät.

## Ympäristömuuttujat

Nämä on tarkoitettu testaukseen ja automaatioon; tavallisessa käytössä mitään niistä ei tarvita.

| Muuttuja | Merkitys |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | muut portit |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | vastaanottokansio, ohjelman datakansio |
| `LANTRANSFER_NAME` | muille näytettävä tietokoneen nimi |
| `LANTRANSFER_LANG` | käyttöliittymän ja komentorivin kieli (`en`, `de`, `pl`, …); ohittaa tallennetun valinnan |
| `LANTRANSFER_LOOPBACK=1` | vain 127.0.0.1: ei yleislähetyksiä, ei LAN-skannausta, ei palomuurisääntöjä |
| `LANTRANSFER_PEER_UDP_PORTS` | muiden instanssien UDP-portit loopback-tilassa |
| `LANTRANSFER_NO_FIREWALL=1` | älä luo tai tarkista palomuurisääntöjä |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | vastaanottajan kiinteä koodi, lähettäjän käyttämä koodi |
| `LANTRANSFER_REQUIRE_CODE=0` | `0` = tämä vastaanottaja ei vaadi pariliitoskoodia, `1` = koodi vaaditaan; ohittaa asetuksen |

## Kääntäminen ja testit

Go 1.23 tai uudempi, Windows 10/11.

```powershell
build.bat
go test ./...
python tools/check_i18n.py
```

`build.bat` suorittaa testit ja kääntää tiedostot `LANTransfer.exe` (ikkuna) ja `LANTransferCLI.exe` (konsoli) käyttäen versiota tiedostosta `VERSION` (oletus `2.1.0`). Testit sisältävät päästä päähän -siirtoja osoitteen 127.0.0.1 kautta: pariliitos ja lukitus, hylätyt todentamattomat pyynnöt, tiedostot 0 B:stä 96 Mt:hen, Unicode-kansiot ja yli 260 merkin polut, nimiristiriidat, vihamieliset nimet, väärä tarkistussumma, peruutus ja jatkaminen, täysi levy sekä varmenteen kiinnitys.

Käännökset ovat tiedostoissa `locales/ui.<code>.json` (puola on lähdekieli ja englanti varakieli). `tools/check_i18n.py` tarkistaa, että kaikissa 30 tiedostossa on samat avaimet, paikkamerkit ja monikkomuodot eikä niissä ole jäljellä puolankielistä tekstiä; Go-testi tarkistaa, että jokainen ohjelman käyttämä avain on olemassa jokaisessa kielessä.

Julkaisutiedot: [docs/releases/v2.1.0.md](../docs/releases/v2.1.0.md) · [v2.0.0](../docs/releases/v2.0.0.md).

## Lisenssi

MIT, katso [LICENSE](../LICENSE).

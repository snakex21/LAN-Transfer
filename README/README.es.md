**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#descarga-y-ejecución)

Una pequeña herramienta portátil para enviar archivos y carpetas directamente entre dos equipos de la misma red local. Sin nube, sin cuenta y sin servidor intermediario: los datos van directamente de un PC al otro a través de una conexión cifrada. Núcleo en Go, una ventana WebView2 y una versión de línea de comandos para scripts.

![Ventana principal de LAN Transfer: un receptor emparejado y tres elementos listos para enviar](../docs/screenshots/main.png)

## Contenido

- [Características](#características)
- [Idioma de la interfaz](#idioma-de-la-interfaz)
- [Capturas de pantalla](#capturas-de-pantalla)
- [Descarga y ejecución](#descarga-y-ejecución)
- [Cómo emparejar](#cómo-emparejar)
- [Envío de archivos](#envío-de-archivos)
- [Línea de comandos](#línea-de-comandos)
- [Puertos y cortafuegos](#puertos-y-cortafuegos)
- [Notas de seguridad](#notas-de-seguridad)
- [Variables de entorno](#variables-de-entorno)
- [Compilación y pruebas](#compilación-y-pruebas)
- [Licencia](#licencia)

## Características

- Detección automática de otros equipos que ejecutan LAN Transfer (difusión UDP, con un escaneo limitado de la red /24 como alternativa).
- Emparejamiento con un código de 6 dígitos que se muestra en el receptor. Sin emparejamiento no se puede enviar nada ni explorar carpetas.
- Modo opcional sin código: en el receptor, desactiva **Exigir código de emparejamiento** en Configuración (se guarda en `data\settings.json`) o inicia `serve --no-code`. Así cualquier equipo con LAN Transfer en la red puede enviarle archivos sin código, siempre cifrados; mientras tanto el encabezado muestra **Sin código**.
- Cifrado TLS 1.3. Cada ejecución crea un nuevo certificado autofirmado; su huella digital se fija durante el emparejamiento y se comprueba en cada conexión.
- Archivos y carpetas completas, conservando la estructura de carpetas y enviando hasta 4 archivos en paralelo.
- Reanudación de transferencias interrumpidas y verificación SHA-256 de cada archivo.
- Los datos incompletos se guardan en una carpeta oculta `.lantransfer-partial`, de modo que un archivo truncado nunca aparece con su nombre definitivo.
- Los conflictos de nombres generan `file (1).txt` o `Folder (1)`; los archivos existentes nunca se sobrescriben.
- Se rechazan los nombres peligrosos (`..`, rutas absolutas, flujos de datos alternativos `file:stream`, nombres reservados como `CON`, `NUL`, `COM1`). Antes de recibir se comprueba el espacio libre en disco.
- Arrastrar y soltar desde el Explorador de archivos, una cola con progreso por archivo, velocidad, tiempo restante, cancelación, reanudación y un historial de elementos enviados y recibidos.
- Portátil: un único EXE, con la configuración y el historial en una carpeta `data` junto a él.
- La ventana, su título y los mensajes de la línea de comandos en 30 idiomas (los mismos que este README), elegidos automáticamente según el idioma para mostrar de Windows; diseño de derecha a izquierda para el árabe y el hebreo.

## Idioma de la interfaz

LAN Transfer habla los mismos 30 idiomas que este README. En el primer inicio usa el idioma para mostrar de Windows (el primer idioma compatible de la lista de idiomas de Windows); si Windows usa un idioma que no está en la lista, la interfaz aparece en inglés.

Para cambiarlo, haz clic en el botón del engranaje (**Configuración**) en la esquina superior derecha y elige un idioma. La ventana, su título y todos los mensajes cambian al instante, y la elección se guarda en `data\settings.json`. **Automático (Windows)** vuelve a seguir el idioma para mostrar de Windows. Los números, tamaños, velocidades, fechas y formas del plural siguen las reglas del idioma elegido; el árabe y el hebreo usan un diseño reflejado de derecha a izquierda, mientras que las rutas, las direcciones IP, los códigos y las huellas digitales siguen escribiéndose de izquierda a derecha. Los mensajes del otro equipo (por ejemplo, «no hay espacio suficiente en el disco del destinatario») se muestran en tu idioma, no en el del receptor.

`LANTransferCLI.exe` usa el mismo idioma. `LANTRANSFER_LANG` (por ejemplo, `LANTRANSFER_LANG=de`) tiene prioridad sobre la elección guardada, tanto en la ventana como en la línea de comandos.

| English | Deutsch | العربية (RTL) |
|---|---|---|
| ![LAN Transfer en inglés](../docs/screenshots/lang-en.png) | ![LAN Transfer en alemán](../docs/screenshots/lang-de.png) | ![LAN Transfer en árabe, diseño de derecha a izquierda](../docs/screenshots/lang-ar.png) |

## Capturas de pantalla

| Emparejamiento | Envío con progreso |
|---|---|
| ![Diálogo de emparejamiento con el código de 6 dígitos y la huella digital del certificado](../docs/screenshots/pairing.png) | ![Cola de transferencias con progreso, velocidad y tiempo restante](../docs/screenshots/sending.png) |
| **Reanudación tras cancelar** | **Receptor: historial** |
| ![Una transferencia cancelada reanudada desde el punto ya recibido, verificada con SHA-256](../docs/screenshots/resume.png) | ![Ventana del receptor con los elementos recibidos en el historial](../docs/screenshots/history.png) |

## Descarga y ejecución

1. Descarga `LANTransfer-2.1.0-windows-amd64.zip` desde [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) y compruébalo con el archivo `.sha256` publicado.
2. Extráelo en cualquier carpeta de ambos equipos.
3. Inicia `LANTransfer.exe`. En el primer inicio, Windows pide permiso (UAC) para añadir dos reglas de cortafuegos; consulta [Puertos y cortafuegos](#puertos-y-cortafuegos).

Requisitos: Windows 10 u 11 (64 bits) con el runtime de WebView2, que forma parte de las versiones actuales de Windows. Ambos equipos deben ejecutar la versión 2. La versión 2 no puede comunicarse con la versión 1.

## Cómo emparejar

1. Inicia LAN Transfer en ambos equipos. El receptor muestra su código de emparejamiento y la huella digital de su certificado en la esquina superior derecha.
2. En el equipo emisor, haz clic en el receptor en la lista **Destinatario**. Si no aparece, usa **Conectar manualmente (dirección IP)** e introduce su dirección IP.
3. Escribe el código de 6 dígitos de la pantalla del receptor y compara la huella digital del certificado que muestra el diálogo con la del receptor. Si no coinciden, no continúes.
4. Haz clic en **Emparejar**. El dispositivo muestra ahora **✓ emparejado**.

Tras 5 códigos incorrectos desde un mismo equipo, el receptor bloquea los intentos posteriores de ese equipo. **Nuevo código** crea un código nuevo y cancela todos los emparejamientos existentes. El receptor crea un código y un certificado nuevos cada vez que se inicia, así que, después de reiniciar el receptor, hay que volver a emparejar.

## Envío de archivos

1. Empareja con el receptor.
2. Arrastra archivos o carpetas a la ventana, usa **Seleccionar archivos** / **Seleccionar carpeta** o pega una ruta.
3. Elige el destino: la carpeta predeterminada del receptor (`Downloads\LAN Transfer`) u otra carpeta existente en el receptor.
4. Haz clic en **Enviar**. Cada elemento muestra el progreso, la velocidad y el tiempo restante, y se puede cancelar. Una transferencia cancelada o interrumpida se puede reanudar; solo se envía la parte que falta.

## Línea de comandos

`LANTransferCLI.exe` es el mismo programa sin ventana, para scripts y automatización.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

El código también se puede pasar en `LANTRANSFER_CODE`. Un receptor sin ventana puede usar un código fijo de `LANTRANSFER_PIN`. Volver a ejecutar el mismo comando `send` reanuda una transferencia interrumpida.

## Puertos y cortafuegos

- UDP 45454: detección. TCP 45455: transferencia (HTTPS).
- En el primer inicio, la aplicación pide una sola vez (UAC) añadir dos reglas de entrada: `LAN Transfer 2 Discovery (UDP)` y `LAN Transfer 2 Transfer (TCP)`.
- Las reglas solo aceptan conexiones desde la subred local (`remoteip=LocalSubnet`) y solo en los perfiles de red Privado y Dominio (`profile=private,domain`). En una red que Windows marca como Pública (cafetería, hotel, aeropuerto), los demás equipos no pueden conectarse a LAN Transfer.
- Si el receptor no aparece en tu red doméstica, comprueba en la configuración de red de Windows que la red esté establecida como Privada.
- La versión 1 creaba reglas llamadas `LAN Transfer Discovery` y `LAN Transfer Transfer` para todos los perfiles, incluido Público. Si la versión 2 las encuentra, muestra un aviso que ofrece sustituirlas. No se cambia nada hasta que haces clic en **Actualizar reglas** y confirmas el aviso de Windows.

## Notas de seguridad

- La detección por UDP no está autenticada; es solo una indicación. La confianza procede del emparejamiento: el código más la huella digital del certificado fijada.
- Durante el primer emparejamiento, un atacante activo en la misma red (ataque de intermediario, man in the middle) podría hacerse pasar por el receptor. Compara la huella digital del certificado en ambas pantallas antes de emparejar.
- Un emisor emparejado puede explorar las carpetas del receptor y guardar archivos en cualquier carpeta existente en él (la opción de destino personalizado). Empareja solo equipos de confianza y usa **Nuevo código** para revocar todos los emparejamientos.
- Las reglas de cortafuegos no abren los puertos en redes marcadas como Públicas.
- Con **Exigir código de emparejamiento** desactivado, cualquier persona de la red que use LAN Transfer puede enviar archivos a este equipo y explorar sus carpetas sin código, como en la versión 1 (pero cifrado). Úsalo solo en una red de confianza. Al volver a activar el código se desconectan los remitentes conectados sin él.

## Variables de entorno

Están pensadas para pruebas y automatización; el uso normal no requiere ninguna.

| Variable | Significado |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | otros puertos |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | carpeta de recepción, carpeta de datos del programa |
| `LANTRANSFER_NAME` | nombre del equipo que ven los demás |
| `LANTRANSFER_LANG` | idioma de la interfaz y de la línea de comandos (`en`, `de`, `pl`, …); tiene prioridad sobre la elección guardada |
| `LANTRANSFER_LOOPBACK=1` | solo 127.0.0.1: sin difusiones, sin escaneo de la LAN, sin reglas de cortafuegos |
| `LANTRANSFER_PEER_UDP_PORTS` | puertos UDP de otras instancias en modo loopback |
| `LANTRANSFER_NO_FIREWALL=1` | no crear ni comprobar reglas de cortafuegos |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | código fijo del receptor, código que usa el emisor |
| `LANTRANSFER_REQUIRE_CODE=0` | `0` = este receptor no exige código de emparejamiento, `1` = código obligatorio; tiene prioridad sobre la configuración |

## Compilación y pruebas

Go 1.23 o posterior en Windows 10/11.

```powershell
build.bat
go test ./...
python tools/check_i18n.py
```

`build.bat` ejecuta las pruebas y compila `LANTransfer.exe` (ventana) y `LANTransferCLI.exe` (consola) con la versión de `VERSION` (por defecto `2.1.0`). Las pruebas incluyen transferencias de extremo a extremo a través de 127.0.0.1: emparejamiento y bloqueo, rechazo de solicitudes no autenticadas, archivos de 0 B a 96 MB, carpetas Unicode y rutas de más de 260 caracteres, conflictos de nombres, nombres maliciosos, una suma de comprobación incorrecta, cancelación y reanudación, un disco lleno y la fijación de certificados.

Las traducciones están en `locales/ui.<code>.json` (el polaco es el idioma de origen y el inglés, el de reserva). `tools/check_i18n.py` comprueba que los 30 archivos tienen las mismas claves, marcadores de posición y formas del plural y que no quedan restos de polaco; una prueba de Go comprueba que cada clave que usa el programa existe en todos los idiomas.

Notas de la versión: [docs/releases/v2.1.0.md](../docs/releases/v2.1.0.md) · [v2.0.0](../docs/releases/v2.0.0.md).

## Licencia

MIT, consulta [LICENSE](../LICENSE).

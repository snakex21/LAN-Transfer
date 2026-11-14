# LAN Transfer

Lekki program do bezpośredniego przesyłania plików i folderów w sieci lokalnej.

## Założenia
- Go jako rdzeń aplikacji.
- Normalne okno desktopowe.
- Interfejs HTML/CSS/JS bez Reacta i Electrona.
- Transfer bez chmury i bez serwera pośredniego.
- Automatyczne wykrywanie innych urządzeń w LAN przez UDP broadcast.
- Własny komputer jest filtrowany i nie pojawia się jako cel transferu.
- Sterowanie ręczne przez GUI oraz automatyczne przez CLI, dzięki czemu program może obsługiwać DevSpace lub inny agent z dostępem do terminala.

## Budowanie

Uruchom:

```powershell
build.bat
```

Powstaną dwa pliki:

```text
build\LANTransfer.exe
build\LANTransferCLI.exe
```

`LANTransfer.exe` jest normalną aplikacją okienkową.
`LANTransferCLI.exe` służy do automatyzacji i sterowania przez AI.

Na obu komputerach musi działać LAN Transfer lub odbiornik uruchomiony przez CLI.

## CLI

```powershell
LANTransferCLI.exe devices
LANTransferCLI.exe send "C:\sciezka\plik.zip" 192.168.0.15
LANTransferCLI.exe send "C:\sciezka\folder" NAZWA-KOMPUTERA
LANTransferCLI.exe serve
```

`devices` wypisuje JSON, aby narzędzia AI mogły łatwo odczytać dostępne urządzenia.

## Porty
- UDP 45454 — wykrywanie urządzeń
- TCP 45455 — transfer i lokalne API używane przez okno aplikacji

## Obecny zakres
Program obsługuje wykrywanie innych komputerów, przesyłanie plików, przesyłanie folderów jako archiwum ZIP bez kompresji oraz odbiór do `Downloads\LAN Transfer`.

Kolejne elementy: natywny wybór plików/folderów w GUI, pasek postępu i prędkości, bezpośredni transfer katalogów, resume, sync/delta, zaufane urządzenia i Android.

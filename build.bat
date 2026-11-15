@echo off
setlocal
cd /d "%~dp0"

rem Release version shown in the UI, "LANTransferCLI.exe version" and the file
rem properties (winres/versioninfo.json -> rsrc_windows_amd64.syso).
if "%VERSION%"=="" set VERSION=2.2.0
set LDFLAGS=-s -w -X lantransfer/internal/version.Version=%VERSION%

go test ./...
if errorlevel 1 exit /b %errorlevel%

go build -trimpath -ldflags="%LDFLAGS% -H windowsgui" -o LANTransfer.exe .
if errorlevel 1 exit /b %errorlevel%

go build -trimpath -ldflags="%LDFLAGS%" -o LANTransferCLI.exe .
if errorlevel 1 exit /b %errorlevel%

echo Built: LANTransfer.exe %VERSION%
echo Built: LANTransferCLI.exe %VERSION%

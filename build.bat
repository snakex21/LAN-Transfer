@echo off
setlocal
cd /d "%~dp0"

go build -trimpath -ldflags="-s -w -H windowsgui" -o LANTransfer.exe .
if errorlevel 1 exit /b %errorlevel%

go build -trimpath -ldflags="-s -w" -o LANTransferCLI.exe .
if errorlevel 1 exit /b %errorlevel%

echo Built: LANTransfer.exe
echo Built: LANTransferCLI.exe

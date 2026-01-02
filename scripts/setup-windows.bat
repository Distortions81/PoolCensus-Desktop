@echo off
setlocal enabledelayedexpansion
choco install -y golang nodejs-lts pkgconfiglite
call go install github.com/wailsapp/wails/v2/cmd/wails@latest
cd /d "%~dp0\..\frontend"
npm install
npm run build
cd /d "%~dp0\.."
wails build

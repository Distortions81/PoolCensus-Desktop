@echo off
setlocal enabledelayedexpansion

cd /d "%~dp0\.."

echo Packaging frontend...
cd frontend
npm install
npm run build

echo Building Wails app...
cd /d "%~dp0\.."
wails build

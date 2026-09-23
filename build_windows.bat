@echo off
setlocal
where go >nul 2>nul || (
  echo [ERROR] Go 1.23 or newer is required.
  exit /b 1
)
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -trimpath -ldflags="-s -w" -o RogueTrooper_Patcher_V64.exe ./patcher
if errorlevel 1 exit /b 1
certutil -hashfile RogueTrooper_Patcher_V64.exe SHA256

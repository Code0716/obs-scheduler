@echo off
chcp 65001 > nul
setlocal

rem .env から OBS_PASSWORD を読み込む
for /f "usebackq tokens=1,2 delims==" %%a in ("%~dp0.env") do (
    if "%%a"=="OBS_PASSWORD" set OBS_PASSWORD=%%b
    if "%%a"=="OBS_ADDR"     set OBS_ADDR=%%b
)

powershell.exe -NonInteractive -ExecutionPolicy Bypass ^
  -File "%~dp0obs-scheduler.ps1" ^
  -Start "08:44" ^
  -Stop  "10:00" ^
  -OBSPassword "%OBS_PASSWORD%" ^
  -SkipLaunch

pause

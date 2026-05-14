@echo off
title GameStore Backend (auto-restart)
echo GameStore Backend - Auto Restart Mode
echo.

:loop
echo [%date% %time%] Starting backend...
start /wait /B "" "D:\TMA_Seill\tma_new\tma-backend\server.exe"
echo [%date% %time%] Backend exited with code %errorlevel%. Restarting in 2s...
timeout /t 2 /nobreak >nul
goto loop

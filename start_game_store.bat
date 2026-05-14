@echo off
title GameStore Launcher
echo ========================================
echo   GameStore - Full Stack Launcher
echo ========================================
echo.
echo [1/3] Starting Backend (port 8080)...
start "GameStore Backend" /D "D:\TMA_Seill\tma_new\tma-backend" cmd /c server.exe
timeout /t 3 /nobreak >nul

echo [2/3] Starting TMA Frontend (port 5173)...
start "GameStore TMA" /D "D:\TMA_Seill\tma_new\tma-frontend" cmd /c npx vite --host 0.0.0.0 --port 5173
timeout /t 2 /nobreak >nul

echo [3/3] Starting Admin Panel (port 5174)...
start "GameStore Admin" /D "D:\TMA_Seill\tma_new\admin-panel" cmd /c npx vite --host 0.0.0.0 --port 5174
timeout /t 2 /nobreak >nul

echo.
echo ========================================
echo   All services started!
echo   Backend:  http://localhost:8080
echo   TMA:      http://localhost:5173
echo   Admin:    http://localhost:5174
echo.
echo   Close this window to stop all services
echo ========================================
echo.
pause

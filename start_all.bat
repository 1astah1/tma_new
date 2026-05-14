@echo off
echo ========================================
echo   GameStore - Full Stack Startup
echo ========================================
echo.
echo [1/3] Starting Backend (port 8081)...
wmic process where "name='server.exe'" delete >nul 2>&1
start "GameStore-Backend" cmd /c "set DATABASE_URL=postgres://postgres:postgres@localhost:5432/tma_shop?sslmode=disable&&set JWT_SECRET=super-secret-key-min-32-chars-long!!&&set SERVER_PORT=8081&&cd /d D:\TMA_Seill\tma_new\tma-backend&&server.exe"
timeout /t 3 /nobreak >nul
echo.
echo [2/3] Starting TMA Frontend (port 5173)...
start "GameStore-TMA" cmd /c "cd /d D:\TMA_Seill\tma_new\tma-frontend&&npx vite --host 0.0.0.0 --port 5173"
echo.
echo [3/3] Starting Admin Panel (port 5174)...
start "GameStore-Admin" cmd /c "cd /d D:\TMA_Seill\tma_new\admin-panel&&npx vite --host 0.0.0.0 --port 5174"
echo.
echo ========================================
echo   All services started!
echo   TMA Frontend:  http://localhost:5173
echo   Admin Panel:   http://localhost:5174
echo   Backend API:   http://localhost:8081
echo ========================================
echo.

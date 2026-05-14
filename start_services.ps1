$logDir = "D:\TMA_Seill\tma_new\logs"
New-Item -ItemType Directory -Path $logDir -Force | Out-Null

Start-Process -WindowStyle Hidden -FilePath "cmd.exe" -ArgumentList "/c set DATABASE_URL=postgres://postgres:postgres@localhost:5432/tma_shop?sslmode=disable&&set JWT_SECRET=super-secret-key-min-32-chars-long!!&&set SERVER_PORT=8081&&D:\TMA_Seill\tma_new\tma-backend\server.exe > $logDir\backend.log 2>&1"
Start-Sleep -Seconds 2

Start-Process -WindowStyle Hidden -FilePath "cmd.exe" -ArgumentList "/c cd /d D:\TMA_Seill\tma_new\tma-frontend&&npx vite --host 0.0.0.0 --port 5173 > $logDir\tma.log 2>&1"
Start-Sleep -Seconds 1

Start-Process -WindowStyle Hidden -FilePath "cmd.exe" -ArgumentList "/c cd /d D:\TMA_Seill\tma_new\admin-panel&&npx vite --host 0.0.0.0 --port 5174 > $logDir\admin.log 2>&1"

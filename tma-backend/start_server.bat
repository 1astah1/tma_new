@echo off
set DATABASE_URL=postgres://postgres:postgres@localhost:5432/tma_shop?sslmode=disable
set JWT_SECRET=super-secret-key-min-32-chars-long!!
set SERVER_PORT=8081
start /b server.exe
echo Server started on port 8081

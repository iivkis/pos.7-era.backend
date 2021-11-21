@echo off
set GOOS=darwin
go build -o "server" -ldflags="-s -w" ./cmd/app/main.go
pause
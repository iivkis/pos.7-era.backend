set GOOS=darwin
go build -o "server_macos" -ldflags="-s -w" ./cmd/app/main.go

set GOOS=windows
go build -o "server_windows.exe" -ldflags="-s -w" ./cmd/app/main.go
pause
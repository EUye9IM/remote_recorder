@REM SET GOOS=linux
@REM SET GOARCH=amd64
@REM go build -o webrtc-server-mod=vendor ./src

go build -o webrtc-server.exe -mod=vendor ./src
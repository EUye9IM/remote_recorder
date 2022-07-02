@REM SET GOOS=linux
@REM SET GOARCH=amd64
@REM go build -o server-mod=vendor ./src

go build -o server.exe -mod=vendor ./src
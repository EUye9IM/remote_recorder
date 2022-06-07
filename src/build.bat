@REM SET GOOS=linux
@REM SET GOARCH=amd64
go build -mod=vendor .

.\server.exe
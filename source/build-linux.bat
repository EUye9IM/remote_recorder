SET GOOS=linux
SET GOARCH=amd64
go build -o webrtc-server -mod=vendor ./src

FROM golang:latest

WORKDIR /app

# Goモジュールの初期化
RUN go mod init musicrawl-spotify

RUN go get -u github.com/rivo/tview

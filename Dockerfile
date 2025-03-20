FROM golang:1.23-alpine

WORKDIR /app

# ライブラリのインストール
COPY go.mod .
RUN go mod download

# ソースコードのコピー
COPY . .

# ビルド
RUN go build -o main ./cmd/server

# ポート
EXPOSE 8080

# 実行
CMD ["./main"] 
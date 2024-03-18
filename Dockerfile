FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o quickmatch ./cmd/quickmatch

EXPOSE 8080

CMD ["./quickmatch"]

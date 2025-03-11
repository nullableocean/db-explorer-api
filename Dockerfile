FROM golang:latest AS builder

RUN go version

WORKDIR /build
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./app ./main.go

FROM alpine:latest

WORKDIR /app/

COPY --from=builder /build/app .
COPY --from=builder /build/.env .

CMD ["./app"]
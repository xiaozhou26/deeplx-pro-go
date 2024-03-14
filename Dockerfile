FROM golang:1.19 as builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o deeplx-pro .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/deeplx-pro .

EXPOSE 9000

CMD ["/app/deeplx-pro"]

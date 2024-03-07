FROM golang:1.19 as builder

WORKDIR /app

COPY . .

RUN go mod download && go build -o main .

FROM golang:1.19

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 9000

CMD ["./main"]

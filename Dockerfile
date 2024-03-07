
FROM golang:latest





WORKDIR /app


COPY go.mod go.sum ./

RUN go mod download 

COPY . .


RUN go build -o main .


EXPOSE 9000

CMD ["./main"]

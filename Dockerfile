FROM golang:1.19 as builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .

ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -o deeplx-pro-amd64 .

ENV GOARCH=arm64
RUN go build -o deeplx-pro-arm64 .

ENV GOARCH=arm
ENV GOARM=7
RUN go build -o deeplx-pro-arm .

FROM alpine:latest

WORKDIR /app

# Copy the right binary based on the platform
COPY --from=builder /app/deeplx-pro-* .

EXPOSE 9000

# Run the right binary based on the platform
CMD ["/app/deeplx-pro-${TARGETPLATFORM}"]

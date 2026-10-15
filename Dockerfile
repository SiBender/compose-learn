FROM golang:1.27-alpine
WORKDIR /src
COPY go.mod .
COPY main.go .
RUN go mod tidy
RUN go mod download


RUN go build -trimpath -ldflags="-s -w" -o /app .

ENV ADDRESS=0.0.0.0
ENTRYPOINT ["/app"]

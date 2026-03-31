FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/agent ./cmd/server

FROM alpine:3.20

RUN adduser -D -g '' appuser
WORKDIR /app

COPY --from=builder /out/agent /app/agent
COPY data /app/data

USER appuser

ENTRYPOINT ["/app/agent"]
CMD ["analyze", "AAPL"]

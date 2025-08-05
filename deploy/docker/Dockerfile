FROM golang:1.24.2 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/main ./cmd/api

FROM gcr.io/distroless/static-debian11

COPY --from=builder /app/main /main

EXPOSE 8080

ENTRYPOINT ["/main"]
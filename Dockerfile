FROM golang:1.21-alpine AS builder
WORKDIR /build 
COPY . .
RUN go mod download 
RUN go build -o ./bin/app ./cmd/app

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /build/bin/app ./chatapp
COPY --from=builder /build/.env ./
CMD ["/app/chatapp"]

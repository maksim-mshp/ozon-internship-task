FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app cmd/shortener/main.go

FROM scratch
COPY --from=builder /build/app /app
CMD ["/app"]

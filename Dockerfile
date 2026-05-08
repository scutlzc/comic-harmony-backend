FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /server /server
COPY migrations/ /migrations/
RUN mkdir -p /uploads/comics /uploads/temp

EXPOSE 9090
CMD ["/server"]

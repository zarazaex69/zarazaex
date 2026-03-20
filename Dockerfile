FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /zarazaex server.go

FROM scratch
COPY --from=builder /zarazaex /zarazaex
EXPOSE 8801
CMD ["/zarazaex"]

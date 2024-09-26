# Build stage
FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o databridge cmd/databridge/main.go

# Run stage
FROM alpine:latest  
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/databridge .
COPY --from=builder /app/examples examples
EXPOSE 8080
ENTRYPOINT ["databridge"]
FROM golang:1.21.6-alpine as builder
RUN apk --no-cache add ca-certificates git gcc g++
WORKDIR /build

# Fetch dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . ./
RUN GO111MODULE=on CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build

# Create final image
FROM alpine
WORKDIR /app
RUN mkdir -p logs
COPY --from=builder /build/pawmap .
EXPOSE 8080
CMD ["./pawmap"]

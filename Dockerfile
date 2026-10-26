FROM golang:1.21-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN test -z "$(gofmt -l $(find . -name '*.go'))"
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=build /out/server ./server
RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["./server"]

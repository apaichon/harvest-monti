# monti-gateway image — produced by compose.e2e.yml.
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/monti-api ./cmd/monti-api

FROM alpine:3.20
RUN apk add --no-cache wget ca-certificates
COPY --from=build /out/monti-api /usr/local/bin/monti-api
COPY migrations /etc/monti/migrations
EXPOSE 8081
CMD ["monti-api"]

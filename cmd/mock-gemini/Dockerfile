# mock-gemini image — produced by compose.e2e.yml.
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/mock-gemini ./cmd/mock-gemini

FROM alpine:3.20
RUN apk add --no-cache wget ca-certificates
COPY --from=build /out/mock-gemini /usr/local/bin/mock-gemini
EXPOSE 7070
CMD ["mock-gemini"]

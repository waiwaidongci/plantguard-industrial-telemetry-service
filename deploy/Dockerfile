FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/plantguard-api ./cmd/api

FROM alpine:3.22
RUN adduser -D -u 10001 plantguard
WORKDIR /app
COPY --from=build /out/plantguard-api /app/plantguard-api
COPY configs ./configs
COPY migrations ./migrations
USER plantguard
EXPOSE 8080
ENTRYPOINT ["/app/plantguard-api"]

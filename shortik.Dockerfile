FROM golang:latest as build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /shortik ./cmd/shortik
 
FROM debian:trixie-slim as run
COPY --from=build /shortik /shortik/shortik
WORKDIR /shortik

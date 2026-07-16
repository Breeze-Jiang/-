FROM golang:1.24-alpine AS goose-build
RUN GOBIN=/out go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

FROM alpine:3.21 AS migrate
COPY --from=goose-build /out/goose /goose
ENTRYPOINT ["/goose"]

FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/alongtu-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/alongtu-api /alongtu-api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/alongtu-api"]

FROM golang:1.24.4 AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o /out/playground-api ./cmd/playground-api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/playground-api /playground-api
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/playground-api"]


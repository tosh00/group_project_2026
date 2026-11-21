# Backend: serwer API (cmd/server) i CLI solvera (cmd/solve) w jednym obrazie.
#   docker build -t planer-server .
# Uruchomienie razem z bazą i frontendem: docker compose up (docker-compose.yml).

FROM golang:1.23-alpine AS build
WORKDIR /src

# Najpierw zależności, żeby zmiana kodu nie unieważniała cache modułów.
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY migrations/ migrations/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/solve ./cmd/solve

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /out/solve /usr/local/bin/
# Migracje są osadzone w binarce (migrations/migrations.go), serwer stosuje je przy starcie.
ENV ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/server"]

# Stage 1: Toolchain — install compilers, code generators, and Node.js
# Almost never changes -> maximale Cache-Nutzung
FROM golang:trixie AS toolchain

RUN go install github.com/a-h/templ/cmd/templ@latest && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && \
    go install github.com/pressly/goose/v3/cmd/goose@latest && \
    go install github.com/air-verse/air@latest

RUN apt-get update && apt-get install -y nodejs npm

# Stage 2: Dependencies — Go modules + npm packages
# Nur invalidated wenn go.mod/go.sum/package.json/package-lock.json sich aendern
FROM toolchain AS deps

WORKDIR /opt/app

COPY go.mod go.sum package.json package-lock.json ./

# go mod tidy NICHT hier — wuerde Layer aufblaehen und Cache zerstoeren.
# go build holt fehlende Module automatisch.
RUN go mod download && npm ci

# Stage 3: Production Build — kompiliert Binary + generiert alle Assets
FROM deps AS build

RUN apt-get update && apt-get install -y just dumb-init

COPY . .

RUN mkdir -p /opt/app/files /opt/app/public

# full-build: templ -> tailwind -> sqlc -> wasm -> CGO_ENABLED=1 go build
RUN just full-build

# Stage 4: Production Runtime — minimiertes Distroless-Image
FROM gcr.io/distroless/base-debian13 AS prod

EXPOSE 8000
WORKDIR /opt/app

COPY --from=build /usr/bin/dumb-init /usr/bin/dumb-init
COPY --from=build /opt/app/bin/mainDocker /opt/app/main
COPY --from=build /opt/app/assets /opt/app/assets
COPY --from=build /opt/app/public /opt/app/public
COPY --from=build /opt/app/files /opt/app/files
# cmd/ wird NICHT kopiert — WASM-Quellcode nur zur Build-Zeit noetig,
# das kompilierte WASM liegt in public/wasm/

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["./main"]

# Stage 5: Development — Hot-Reload mit air (alles wird via bind mount
# ueberschrieben, daher kein vorheriges Build notwendig)
FROM deps AS dev

RUN apt-get update && apt-get install -y just

COPY . .

# Vorkompilieren nicht noetig — just watch im CMD macht alles bei Start
CMD ["just", "watch"]

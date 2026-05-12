FROM golang:trixie AS base

WORKDIR /opt/app

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY sqlc.yaml .

RUN go mod tidy

COPY package.json .

RUN apt-get update && apt-get install -y nodejs npm
RUN npm install

COPY Justfile .

FROM base AS prod-builder

WORKDIR /opt/app

RUN apt-get update && apt-get install -y dumb-init just

COPY --from=base /opt/app/node_modules /opt/app/node_modules
COPY .air.toml .
COPY ./assets ./assets
COPY ./docs ./docs
COPY ./internal ./internal
COPY ./sqlc ./sqlc
COPY ./main.go ./main.go
RUN rm -rf ./tmp
RUN go mod tidy

RUN mkdir -p /opt/app/files
RUN mkdir -p /opt/app/public

RUN just full-build

FROM gcr.io/distroless/base-debian13 AS prod
# TOTO: Run as nonroot user

EXPOSE 8000
WORKDIR /opt/app

COPY --from=prod-builder /usr/bin/dumb-init /usr/bin/dumb-init
COPY --from=prod-builder /opt/app/bin/mainDocker /opt/app/main
COPY --from=prod-builder /opt/app/assets /opt/app/assets
# COPY --from=prod-builder /opt/app/docs /opt/app/docs
COPY --from=prod-builder /opt/app/files /opt/app/files
COPY --from=prod-builder /opt/app/public /opt/app/public

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["./main"]

FROM base AS dev

RUN apt-get update && apt-get install -y just

EXPOSE 8000
WORKDIR /opt/app

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY ./go.mod ./go.sum ./
RUN go install github.com/air-verse/air@latest
RUN go mod download
COPY . .
RUN go mod tidy

CMD ["just", "watch"]

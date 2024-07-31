FROM golang:1.22-alpine as base

RUN apk add --no-cache git
RUN apk add --no-cache ca-certificates
 
# add a user here because addgroup and adduser are not available in scratch
RUN addgroup -S myapp && adduser -S -u 10000 -g myapp myapp
 
WORKDIR /opt/app

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY sqlc.yml .

RUN go mod tidy

COPY package.json .
COPY tailwind.config.js .

RUN apk add --no-cache npm
RUN npm install

RUN apk add --no-cache make
COPY Makefile .

FROM base as prod-builder

WORKDIR /opt/app

COPY --from=base /opt/app/node_modules /opt/app/node_modules
COPY . .
RUN rm -rf ./tmp

RUN make full-build

FROM scratch as prod

EXPOSE 8000
WORKDIR /opt/app

COPY --from=prod-builder /opt/app/bin/mainDocker /opt/app/main
COPY --from=prod-builder /opt/app/assets /opt/app/assets
COPY --from=prod-builder /opt/app/docs /opt/app/docs

CMD ["./main"]

FROM base as dev

EXPOSE 8000
WORKDIR /opt/app

COPY .air.toml .
COPY ./assets ./assets
COPY ./docs ./docs
COPY ./internal ./internal
COPY ./sqlc ./sqlc
COPY ./main.go ./main.go

RUN go install github.com/air-verse/air@latest
RUN go mod tidy

CMD ["make", "watch"]

FROM postgres:16 as pg

WORKDIR /root

RUN apt-get update && apt-get install -y wget

RUN wget https://github.com/pksunkara/pgx_ulid/releases/download/v0.1.5/pgx_ulid-v0.1.5-pg16-amd64-linux-gnu.deb
RUN dpkg -i pgx_ulid-v0.1.5-pg16-amd64-linux-gnu.deb

CMD ["postgres"]

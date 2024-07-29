FROM golang:1.22-bookworm as base

WORKDIR /opt/app

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod .
COPY go.sum .
COPY sqlc.yml .

RUN go mod tidy

COPY package.json .
COPY tailwind.config.js .

RUN apt-get update && apt-get install -y nodejs npm
RUN npm install

COPY Makefile .

FROM base as prod-builder

WORKDIR /opt/app

COPY --from=base /opt/app/node_modules /opt/app/node_modules
COPY . .
RUN rm -rf ./tmp

RUN make full-build

FROM golang:1.22-bookworm as prod

EXPOSE 8000
WORKDIR /opt/app

COPY --from=prod-builder /opt/app/bin /opt/app/bin
COPY --from=prod-builder /opt/app/assets /opt/app/assets

CMD ["./bin"]

FROM base as dev

EXPOSE 8000
WORKDIR /opt/app


RUN ls -al

COPY .air.toml .
COPY ./assets ./assets
COPY ./docs ./docs
COPY ./internal ./internal
COPY ./sqlc ./sqlc
COPY ./main.go .opt/app/main.go

CMD ["make", "watch"]

FROM postgres:16 as postgres_ulid

WORKDIR /root

RUN wget https://github.com/pksunkara/pgx_ulid/releases/download/v0.1.5/pgx_ulid-v0.1.5-pg16-amd64-linux-gnu.deb
RUN dpkg -i pgx_ulid-v0.1.5-pg16-amd64-linux-gnu.deb


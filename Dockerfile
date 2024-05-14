FROM golang:1.22 AS builder

ARG BUILD_DIR=/build
ARG FILENAME=main

WORKDIR ${BUILD_DIR}

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ${FILENAME} .

FROM debian:latest

ARG BUILD_DIR=/build
ARG APP_DIR=/app
ARG APP_USER=appuser
ARG FILENAME=main

WORKDIR ${APP_DIR}

COPY --from=builder ${BUILD_DIR}/${FILENAME} .

# ADD USER
RUN adduser --system --no-create-home --home ${APP_DIR} --disabled-login ${APP_USER}

RUN chown ${APP_USER} ${FILENAME} && chmod +x ${FILENAME}

USER ${APP_USER}

EXPOSE 8000

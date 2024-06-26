FROM golang:1.22-alpine AS backend

RUN apk add --no-cache git gcc make musl-dev

COPY ./ /app

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go mod download

RUN make build

FROM node:20 AS frontend

COPY ./web /app

WORKDIR /app

RUN npm install
RUN npm run build

FROM alpine:3.14

COPY --from=backend /app/build/cloudsdale /app/cloudsdale
COPY --from=frontend /app/dist /app/dist

WORKDIR /app

VOLUME /var/run/docker.sock

EXPOSE 8888

CMD ["./cloudsdale"]
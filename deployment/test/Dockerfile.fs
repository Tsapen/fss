FROM golang:latest

ARG FSS_ROOT_DIR
ARG FS_SERVER_PORT

ENV FSS_ROOT_DIR=/app
ENV FS_SERVER_PORT=43000
ENV FS_CONFIG=${FS_TEST_CONFIG}

WORKDIR /app

COPY . .

RUN mkdir ./stored_files

RUN go build -o fs ./cmd/file-server/main.go

CMD ["./fs"]

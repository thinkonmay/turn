FROM golang:latest
WORKDIR /src
COPY . .
RUN go build ./cmd/server/main.go

EXPOSE 49152:65535/tcp
EXPOSE 49152:65535/udp

CMD ./main
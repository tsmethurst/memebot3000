FROM golang:alpine AS builder

COPY . /root/go/src/github.com/tsmethurst/memebot3000
WORKDIR /root/go/src/github.com/tsmethurst/memebot3000
RUN go build ./cmd/memebot3000
RUN ls -lha

FROM alpine:3.12 AS runtime

COPY --from=builder /root/go/src/github.com/tsmethurst/memebot3000/memebot3000 /memebot3000

ENTRYPOINT [ "/memebot3000" ]

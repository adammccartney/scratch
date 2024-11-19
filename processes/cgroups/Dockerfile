FROM alpine:3.20.3 AS builder

WORKDIR /app
COPY coro-sieve.c /app
COPY Makefile /app

RUN apk update \
  && apk upgrade \
  && apk add -U gcc make musl-dev --no-cache \
  && make clean && make && make install

FROM alpine:3.20.3

RUN apk update \
 && apk upgrade \
 && apk add -U bash musl-dev --no-cache

COPY --from=builder /usr/local/bin/coro-sieve /usr/local/bin/coro-sieve
COPY entrypoint.sh /app/entrypoint.sh
COPY README.md /app/README.md

USER 1001
ENTRYPOINT ["/app/entrypoint.sh"]

FROM alpine:3.20.3 AS builder

WORKDIR /app
COPY coro-sieve.c /app
COPY Makefile /app

RUN apk update \
  && apk upgrade \
  && apk add -U gcc make musl-dev --no-cache \
  && make clean && make

FROM alpine:3.20.3

RUN apk update \
 && apk upgrade \
 && apk add -U bash musl-dev --no-cache

COPY --from=builder /app/coro-sieve /usr/local/bin/coro-sieve
COPY --chmod=0755 entrypoint.sh /app/entrypoint.sh
COPY --chmod=0755 README.md /app/README.md

ENTRYPOINT ["/app/entrypoint.sh"]

FROM golang:1.25-alpine@sha256:2ad042dec672e85d9e631feb0d2d72db86fd2a4e0cf8daaf2c19771a26df1062 as builder

COPY . /src/hold-this
WORKDIR /src/hold-this

RUN set -ex \
 && apk add --update git \
 && go install \
      -ldflags "-X main.version=$(git describe --tags --always || echo dev)" \
      -mod=readonly \
      -modcacherw \
      -trimpath


FROM alpine:3.22@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1

LABEL maintainer="Knut Ahlers <knut@ahlers.me>"

RUN set -ex \
 && apk --no-cache add \
      ca-certificates

COPY --from=builder /go/bin/hold-this /usr/local/bin/hold-this

EXPOSE 3000
USER 1000:1000

ENTRYPOINT ["/usr/local/bin/hold-this"]
CMD ["--"]

# vim: set ft=Dockerfile:

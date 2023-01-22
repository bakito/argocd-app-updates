FROM golang:1.19 as builder

WORKDIR /build
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y upx --no-install-recommends

ARG VERSION=main

COPY . .

ENV GOPROXY=https://goproxy.io \
    GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
RUN go build -a -installsuffix cgo -ldflags="-w -s -X github.com/bakito/argocd-app-updates/version.Version=${VERSION}" -o argocd-app-updates . \
  && upx -q argocd-app-updates


# application image

FROM scratch

LABEL maintainer="bakito <github@bakito.ch>"
EXPOSE 8080
USER 1001
ENTRYPOINT ["/go/bin/argocd-app-updates", "--server" ]
# define default k8s argocd server address
CMD ["--argo-server", "http://argocd-server:80"]
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/argocd-app-updates /go/bin/argocd-app-updates

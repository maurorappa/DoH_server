FROM golang:buster as builder
WORKDIR /go/src/doh
COPY . /go/src/doh
RUN wget -O /sbin/dep -q https://github.com/golang/dep/releases/download/v0.5.4/dep-linux-amd64 && chmod +x /sbin/dep
RUN /sbin/dep ensure -add github.com/ReneKroon/ttlcache
RUN /sbin/dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-D_FORTIFY_SOURCE=2,-static,-Wl,-z,noexecstack,relro"' -o doh_server *.go

FROM busybox:latest
RUN mkdir -p /srv/ssl
COPY --from=builder /go/src/doh/doh_server /srv/doh_server
COPY doh-server.conf /srv/doh-server.conf
USER nobody

EXPOSE     4443
ENTRYPOINT [ "/srv/doh_server" ]
CMD [ "-c", "/srv/doh-server.conf" ]

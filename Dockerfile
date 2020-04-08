FROM golang:buster as builder
WORKDIR /go/src/doh
COPY . /go/src/doh
RUN wget -O /sbin/dep -q https://github.com/golang/dep/releases/download/v0.5.4/dep-linux-amd64 && chmod +x /sbin/dep
RUN /sbin/dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-D_FORTIFY_SOURCE=2,-static,-Wl,-z,noexecstack,relro"' -o doh_server *.go

FROM scratch
COPY --from=builder /go/src/doh/doh_server /bin/doh_server
COPY doh-server.conf /srv/doh-server.conf

EXPOSE     4443
ENTRYPOINT [ "/bin/doh_server" ]
#CMD [ "-c", "/srv/doh-server.conf" ]

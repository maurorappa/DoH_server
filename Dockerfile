FROM golang:alpine as builder
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -o doh_server *.go

FROM busybox:latest
RUN mkdir -p /srv/ssl
COPY --from=builder doh_server /srv/doh_server
COPY doh-server.conf /srv/doh-server.conf

EXPOSE     443
ENTRYPOINT [ "/srv/doh_server" ]
CMD [ "-c", "/srv/doh-server.conf" ]

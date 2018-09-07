FROM quay.io/prometheus/busybox:latest
RUN mkdir /srv
COPY doh_server /srv/doh_server
COPY doh-server-docker.conf /srv/doh-server.conf
COPY cert.pem /srv/cert.pem
COPY privkey.pem /srv/privkey.pem

EXPOSE     443
ENTRYPOINT [ "/srv/doh_server" ]
CMD [ "-conf", "/srv/doh-server.conf" ]

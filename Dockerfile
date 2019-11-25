FROM busybox:latest
RUN mkdir -p /srv/ssl
COPY doh_server /srv/doh_server
COPY doh-server.conf /srv/doh-server.conf

EXPOSE     443
ENTRYPOINT [ "/srv/doh_server" ]
CMD [ "-c", "/srv/doh-server.conf" ]

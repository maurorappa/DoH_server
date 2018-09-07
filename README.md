# DoH_server
DNS over HTTP2S server 

Modified version of the server used in https://github.com/m13253/dns-over-https

STEPS to use it:

dep ensure

GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -a -installsuffix cgo -o doh_server *.go

You need to get a valid HTTPS certificate (from Letsencrypt for example)

Edit doh-server-docker.conf with the certificates keys

docker build . -t doh:0.1

docker run -tid --rm -p 443:443 --name doh doh:0.1

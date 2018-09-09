# DoH_server
DNS over HTTP2S server 

Modified version of the server used in https://github.com/m13253/dns-over-https

Steps to build a container:

 - Grab all modules needed ```dep ensure```

 -  Build a static binary ```GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -a -installsuffix cgo -o doh_server *.go```

 -  You need to get a valid HTTPS certificate (from Letsencrypt for example)

 -  Edit doh-server-docker.conf with the certificates keys

 -  Build a minimal container ```docker build . -t doh:0.1```

 -  Run it ```docker run -tid --rm -p 443:443 --name doh doh:0.1```

Enhancement to the original project:

 - ability to Skip Ipv6 dns queries to speed up resolution (details here https://github.com/m13253/dns-over-https/pull/19)

 - instead of DNS roundrobin I implemented an primitive algorithm to use the fastest DNS server out of the specified pool and continually monitor which server is the fastest

Tips for implementation:

 - I use the standard HTTPS port (443) to run this service so my Firefox can use it even behind a corporate firewall (even if having a proxy they can see my surfing activity anyway)
 

We all should be thankful forever to Let's Encrypt

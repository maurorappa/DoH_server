# DoH_server
DNS over HTTP2S server 

Inspired by https://github.com/m13253/dns-over-https

# FAQs:

 - Can I use plain HTTP?  No, by design you need HTTPS with a proper certificate 

 - Shall I run as a service or as container?  Being a service exposed over Internet, you should use a container to isolate from the OS, in case you want to run on the plain OS consider to use FireJail. 

 - Do you think my code is crappy? Help me to write a better one!

# Tips for implementation:

 - I use the standard HTTPS port (443) to run this service so my Firefox can use it even behind a corporate firewall (even if having a proxy they can see my surfing activity anyway)


 # Steps to build a container:

 - Grab all modules needed ```dep ensure```

 -  Build a static binary ```GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -a -installsuffix cgo -o doh_server *.go```

 -  You need to get a valid HTTPS certificate (from Letsencrypt for example)

 -  Edit doh-server-docker.conf with the certificates keys

 -  Build a minimal container ```docker build . -t doh:0.1```

 -  Run it ```docker run -tid --rm -p 443:443 --name doh doh:0.1```

# Enhancement to the original project:
 
 - ability to Skip Ipv6 dns queries to speed up resolution (details here https://github.com/m13253/dns-over-https/pull/19)

 - instead of DNS roundrobin I implemented a primitive algorithm to use the fastest DNS server out of the specified pool and continually monitor which server is the fastest

 

# Last notes: 

- We all should be thankful forever to Let's Encrypt

- Best OS to run it: Devuan

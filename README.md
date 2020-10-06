# DoH_server
run you own DNS over HTTP2S server 

to run quickly run it:

1- make `doh` directory

2- copy `doh-server.conf` and edit to suit your needs

3- copy certificate ad private key, for example if you used letsencrypt with certbot
   `cp /etc/letsencrypt/live/doh.yapily.com/cert.pem doh/
   cp /etc/letsencrypt/live/doh.yapily.com/privkey.pem doh/`
   
4- run docker `docker run -it -p 443:443 -v /home/mauro/doh:/svc:ro privatesurfing/doh`
 
Docker images available here: 

https://hub.docker.com/repository/docker/privatesurfing/doh

Info on browser setup: 

https://support.mozilla.org/en-US/kb/firefox-dns-over-https

https://windowsloop.com/enable-dns-over-https-chrome/

https://daniel.haxx.se/blog/2018/06/03/inside-firefoxs-doh-engine/

https://github.com/coredns/coredns/issues/1650


# FAQs:

 - Can I use plain HTTP?  No, by design you need HTTPS with a proper certificate 

 - Shall I run as a service or as container?  Being a service exposed over Internet, you should use a container to isolate from the OS, in case you want to run on the plain OS consider to use FireJail. 

 - Do you think my code is crappy? Help me to write a better one!

 - Is it a secure 'container'? The server, a static hardened Go binary, is the only process and it does not perform any outgoing connections other than dns queries
 

# Tips for implementation:

 - I use the standard HTTPS port (443) to run this service so my Firefox can use it even behind a corporate firewall (even if having a proxy they can see my surfing activity anyway)


# Enhancement to the original project
(https://github.com/m13253/dns-over-https):

 - use of the strongest TLS ciphers, random SessionTicket for every connection *

 - ability to Skip Ipv6 dns queries to speed up resolution (details here https://github.com/m13253/dns-over-https/pull/19)

 - instead of DNS roundrobin I implemented a primitive algorithm to use the fastest DNS server out of the specified pool and continually monitor which server is the fastest
 - IP whitelisting, only authorized IP/networks can use it
 
 - /stat page provides upstream DNS latency and relative usage
 
 
# ToDo

 - write some test code 


# Last notes: 

- get your certificates using Let's Encrypt, see https://letsencrypt.org/getting-started/

- I use Pi-hole (https://pi-hole.net/) as DNS server, so you block all Ads queries

# References

* https://blog.ungleich.ch/en-us/cms/blog/2018/08/04/mozillas-new-dns-resolution-is-dangerous/

* https://www.ispreview.co.uk/index.php/2019/09/firefox-says-no-dns-over-https-doh-by-default-for-uk.html

* https://blog.filippo.io/we-need-to-talk-about-session-tickets/

* https://blog.twitter.com/engineering/en_us/a/2013/forward-secrecy-at-twitter.html


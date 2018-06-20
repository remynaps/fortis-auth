FROM google/debian:wheezy
MAINTAINER Remy Span

# Add the api binary
ADD ./bin/fortis_api fortis

# Add the certs TODO: move to persistent volume
# ADD ./config/jwt/app.rsa /config/jwt/app.rsa
# ADD ./config/jwt/app.rsa.pub /config/jwt/app.rsa.pub

ENV PORT 6767
EXPOSE 6767
CMD ["/fortis"]

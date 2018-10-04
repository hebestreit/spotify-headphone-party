FROM ubuntu:latest

LABEL maintainer="Daniel Hebestreit"

WORKDIR /usr/share/spotify-headphone-party

COPY ./bin/spotify-headphone-party ./spotify-headphone-party
COPY ./template ./template

EXPOSE "8090"

CMD ["./spotify-headphone-party"]
FROM ubuntu:latest

LABEL maintainer="Daniel Hebestreit"

WORKDIR /usr/share/spotify-party

COPY ./bin/spotify-party ./spotify-party
COPY ./template ./template

EXPOSE "8090"

CMD ["./spotify-party"]
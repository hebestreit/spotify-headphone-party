# Spotify Headphone Party [![Build Status](https://travis-ci.org/hebestreit/spotify-headphone-party.svg)](https://travis-ci.org/hebestreit/spotify-headphone-party)

With this app you will be able to start your own Spotify headphone party. Host your own party and invite your friends to listen music together.

## Install

### Docker image

The easiest way is to run your own Docker container using this command. Open `http://localhost:8090` in your browser and login with your Spotify credentials.

    $ docker run \
    -p 8090:8090 \
    -e SESSION_KEY=top-secret-session-key \
    -e SPOTIFY_ID=<clientID> \
    -e SPOTIFY_SECRET=<secretKey> \
    -e SPOTIFY_REDIRECT_URL=http://localhost:8090/callback \
    -e LOG_LEVEL=info \
    hebestreit/spotify-headphone-party

I've also created an example with Docker Compose. Simply copy `docker-compose.env.dist` to `docker-compose.env` and update all environment values. 

    $ cp docker-compose.env.dist docker-compose.env # update all environment values
    $ docker-compose up -f docker-compose.yml -d

## Developing

First clone this repository. 

    $ git clone git@github.com:hebestreit/spotify-headphone-party.git

### Using Docker Compose and remote debug
    
Then simply run this Docker Compose environment which allows you debugging your code using Delve. It'll also start a redis service which is connected to this application.

    $ cp docker-compose.env.dist docker-compose.env # update all environment values
    $ docker-compose -f docker-compose.dev.yml up --build

Open `http://localhost:8090` in your browser. Inside of your IDE settings configure a new remote debugger. 

Special thanks to following authors:

* https://mikemadisonweb.github.io/2018/06/14/go-remote-debug/
* https://medium.com/@wenbinzhang0802/setup-development-environment-for-go-using-docker-and-vscode-bb41c6ab0948

## TODOs

* Use docker secrets for Spotify credentials
* Tests, tests and more tests :-)
* Modern design with JavaScript framework

## License

MIT, see [LICENSE](https://github.com/hebestreit/spotify-headphone-party/blob/master/LICENSE).

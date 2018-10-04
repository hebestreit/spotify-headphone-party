# Spotify Headphone Party [![Build Status](https://travis-ci.org/hebestreit/spotify-headphone-party.svg)](https://travis-ci.org/hebestreit/spotify-headphone-party)

With this app you will be able to start your own Spotify headphone party. Host your own party and invite your friends to listen together music.

## Install

### Docker image

The easiest way is to run your own Docker container using this command. Open `http://localhost:8090` in your browser and login with your Spotify credentials.

    $ docker run \
    -p 8090:8090 \
    -e SPOTIFY_ID=<clientID> \
    -e SPOTIFY_SECRET=<secretKey> \
    -e LOG_LEVEL=info \
    hebestreit/spotify-headphone-party

### Building from sources

First clone this repository using Git. 

    $ git clone git@github.com:hebestreit/spotify-headphone-party.git 

To build this from sources you need a Go environment and Docker installed locally.

    $ make all

Now you can start this application by running `./bin/spotify-headphone-party` or build your own container with `docker build . -t hebestreit/spotify-headphone-party:latest` and use the Docker command from above.

## License

MIT, see [LICENSE](https://github.com/hebestreit/spotify-headphone-party/blob/master/LICENSE).

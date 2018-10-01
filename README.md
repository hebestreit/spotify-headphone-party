# Spotify Headphone Party [![Build Status](https://travis-ci.org/hebestreit/spotify-headphone-party.svg)](https://travis-ci.org/hebestreit/spotify-headphone-party)

Listen Spotify together with your friends online.

## Install

### Docker image

The easiest way is to run your own Docker container using this command.

    $ docker run \
    -p 8090:8090 \
    -e SPOTIFY_ID=<clientID> \
    -e SPOTIFY_SECRET=<secretKey> \
    -e LOG_LEVEL=info \
    hebestreit/spotify-party

### Building from sources

To build this from sources you need a Go environment and Docker installed locally.

    $ make all

Now you can start this application by running `./bin/spotify-party` or using the Docker command from above.

## License

MIT, see [LICENSE](https://github.com/hebestreit/spotify-headphone-party/blob/master/LICENSE).

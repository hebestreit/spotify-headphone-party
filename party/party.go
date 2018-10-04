package party

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/hebestreit/spotify-headphone-party/redis"
	"github.com/satori/go.uuid"
	"github.com/zmb3/spotify"
	"time"
)

// threshold in ms syncing track will be ignored
const syncThreshold = 1000

// threshold in ms between a song switch by host
const songSwitchThreshold = 2500

// threshold in ms to check if song has remaining time
const remainingTimeThreshold = 500

// time in ms when current track should be fetched
const updateCurrentTrackDelay = 400

type Party struct {
	ID         string
	HostUserID string
}

// create and persist a new party to Redis
func CreateParty(host *User) *Party {
	party := &Party{ID: uuid.NewV4().String(), HostUserID: host.ID}

	data, err := json.Marshal(party)
	if err != nil {
		panic(err)
	}

	pool := redis.NewPool()
	defer pool.Close()

	c := pool.Get()

	err = c.Send("MULTI")
	if err != nil {
		panic(err)
	}

	err = c.Send("SET", fmt.Sprintf("party:%s", party.ID), data)
	if err != nil {
		panic(err)
	}

	err = c.Send("SADD", "parties", party.ID)
	if err != nil {
		panic(err)
	}

	_, err = c.Do("EXEC")
	if err != nil {
		panic(err)
	}

	return party
}

// check if party is still live
func IsPartyLive(partyId string) (bool, error) {
	party, err := FindParty(partyId)
	return nil != party, err
}

// retrieve party from redis
func FindParty(partyId string) (*Party, error) {
	pool := redis.NewPool()
	defer pool.Close()

	c := pool.Get()

	var party Party
	reply, err := c.Do("GET", fmt.Sprintf("party:%s", partyId))
	if err != nil || reply == nil {
		return nil, err
	}

	if err := json.Unmarshal(reply.([]byte), &party); err != nil {
		return nil, err
	}

	return &party, nil
}

// retrieve party from redis
func FindAll() (*[]Party, error) {
	pool := redis.NewPool()
	defer pool.Close()

	c := pool.Get()

	values, err := redigo.Values(c.Do("SORT", "parties",
		"BY", "party:*->ID",
		"GET", "party:*"))
	if err != nil {
		return nil, err
	}

	var parties []Party

	for _, value := range values {
		if value == nil {
			continue
		}

		var party Party
		if err := json.Unmarshal(value.([]byte), &party); err != nil {
			return nil, err
		}

		parties = append(parties, party)
	}

	return &parties, nil
}

// delete party from redis
func DeleteParty(party *Party) error {
	pool := redis.NewPool()
	defer pool.Close()

	c := pool.Get()

	err := c.Send("MULTI")
	if err != nil {
		panic(err)
	}

	err = c.Send("DEL", fmt.Sprintf("party:%s", party.ID))
	if err != nil {
		return err
	}

	err = c.Send("SREM", "parties", party.ID)
	if err != nil {
		panic(err)
	}

	_, err = c.Do("EXEC")
	if err != nil {
		panic(err)
	}

	return nil
}

// TODO refactor Join and Host as separate asynchronous worker

// go routine to subscribe to party channel and update currently playing track
func (party *Party) Join(user *User) error {
	if party.HostUserID == user.ID {
		return errors.New("you can't join your own party")
	}

	pool := redis.NewPool()
	defer pool.Close()

	c := pool.Get()

	pubsub := redigo.PubSubConn{c}

	err := pubsub.Subscribe(party.pubSubChannelName())
	if err != nil {
		panic(err)
	}

	var currentPlaying spotify.CurrentlyPlaying

	var inSync = false
	// this loop will receive messages from channel until an error occurs
	for c.Err() == nil {
		if isLive, _ := IsPartyLive(party.ID); !isLive {
			log.WithField("partyID", party.ID).Debugf("Party is not live anymore or has been deleted.")
			_ = user.SpotifyClient().Pause()
			return nil
		}

		switch v := pubsub.Receive().(type) {
		case redigo.Message:
			now := time.Now()
			if err := json.Unmarshal(v.Data, &currentPlaying); err != nil {
				log.Error(err)
				continue
			}

			log.WithFields(log.Fields{
				"itemURI":  currentPlaying.Item.URI,
				"progress": currentPlaying.Progress,
			}).Debugf("Received current playing track.")

			myCurrentPlaying, err := user.SpotifyClient().PlayerCurrentlyPlaying()
			playOptions := &spotify.PlayOptions{
				URIs: []spotify.URI{currentPlaying.Item.URI},
			}

			if err != nil {
				// sometimes no active device was found and we can't play any tracks - see https://github.com/spotify/web-api/issues/924
				if err.Error() == "EOF" {
					log.WithField("userID", user.ID).Debug("No active device found")
				} else {
					log.Error(err)
				}
				continue
			}

			if myCurrentPlaying.Item.URI != currentPlaying.Item.URI {
				// user has manually changed song to force leaving party
				if inSync == true && currentPlaying.Progress > songSwitchThreshold {
					// TODO check if we can do this smarter
					return nil
				}
				err := user.SpotifyClient().PlayOpt(playOptions)
				if err != nil {
					log.Error(err)
					continue
				}
			}

			// set to true is user should be in sync with host
			inSync = true

			if !currentPlaying.Playing {
				if myCurrentPlaying.Playing {
					// pause my player if host stopped track and my isn't stopped yet
					_ = user.SpotifyClient().Pause()
					continue
				}
			} else if !myCurrentPlaying.Playing {
				// user has manually paused player to force leaving party
				if inSync == true && (currentPlaying.Item.Duration-currentPlaying.Progress > remainingTimeThreshold) {
					// TODO check if we can do this smarter
					return nil
				}
				// otherwise start track in my player
				_ = user.SpotifyClient().Play()
			}

			// calculate elapsed time between api calls to get a better synchronisation between host and me
			elapsedTime := time.Now().Second() - now.Second()

			// try to estimate current position of my playing track
			estimatedHostProgressTime := elapsedTime*2 + currentPlaying.Progress

			log.WithFields(log.Fields{
				"progressHost":      currentPlaying.Progress,
				"progressMyself":    myCurrentPlaying.Progress,
				"progressEstimated": estimatedHostProgressTime,
			}).Debug("Estimated progress.")

			// seek to position of currently playing track if threshold is reached
			if (myCurrentPlaying.Progress+syncThreshold) <= estimatedHostProgressTime || (myCurrentPlaying.Progress-syncThreshold) >= estimatedHostProgressTime {
				log.Debug("Seek to track position.")

				_ = user.SpotifyClient().Seek(estimatedHostProgressTime)
			}
		case redigo.Subscription:
			log.Debugf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			log.Error(c.Err())
		}
	}

	return nil
}

// retrieve channel name
func (party *Party) pubSubChannelName() string {
	return fmt.Sprintf("party:%s", party.ID)
}

// leave from a party and unsubscribe from channel
func (party *Party) Leave(user *User) {
	// TODO implement this
}

// go routine to publish currently playing track to redis
func (party *Party) Host() {
	user, err := FindUser(party.HostUserID)
	if err != nil {
		panic(err)
	}

	if user == nil {
		panic(errors.New(fmt.Sprintf("User with id %s is undefined.", party.HostUserID)))
	}

	pool := redis.NewPool()
	defer pool.Close()

	c := pool.Get()
	log.WithField("partyID", party.ID).Debugf("User %s is hosting now...", user.ID)

	for {
		time.Sleep(updateCurrentTrackDelay * time.Millisecond)

		if isLive, _ := IsPartyLive(party.ID); !isLive {
			log.WithField("partyID", party.ID).Debugf("Party is not live anymore or has been deleted.")
			return
		}

		client := auth.NewClient(user.Token)

		// fetch current playing track and information
		currentPlaying, err := client.PlayerCurrentlyPlaying()
		if err != nil {
			if currentPlaying == nil {
				continue
			}
			log.Warnf("Party produces error \"%s\"", err)
			continue
		}

		data, err := json.Marshal(currentPlaying)
		if err != nil {
			panic(err)
		}

		_, err = c.Do("PUBLISH", party.pubSubChannelName(), data)
		if err != nil {
			panic(err)
		}

		log.WithFields(log.Fields{
			"ctemURI":  currentPlaying.Item.URI,
			"progress": currentPlaying.Progress,
		}).Debug("Updated currently playing status.")
	}
}

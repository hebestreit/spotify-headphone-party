package party

import (
	log "github.com/Sirupsen/logrus"
	"github.com/zmb3/spotify"
	"time"
)

type Party struct {
	id    string
	host  *spotify.Client
	users map[string]*User
}

func NewParty(host *spotify.Client) *Party {
	id := "randomPartyId"
	return &Party{id: id, host: host, users: nil}
}

func (party *Party) Join(user *User) {
	party.users[user.id] = user
}

func (party *Party) Leave(user *User) {
	delete(party.users, user.id)
}

func (party *Party) Host() {
	for {
		time.Sleep(3 * time.Second)
		currentPlaying, err := party.host.PlayerCurrentlyPlaying()

		if err != nil {
			if currentPlaying == nil {
				continue
			}
			log.Warnf("Party %s produces error \"%s\"", party.id, err)
			continue
		}

		// TODO implement hosting
		log.Debugf("Party %s is currently playing %s", party.id, currentPlaying.Item.Name)
	}
}

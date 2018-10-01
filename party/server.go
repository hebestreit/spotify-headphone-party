package party

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
	"github.com/zmb3/spotify"
	"html/template"
	"net/http"
	"os"
)

type User struct {
	id     string
	client *spotify.Client
}

type Server struct {
	activeSessions map[string]*User
	addUserCh      chan *User
	removeUserCh   chan *User
	doneCh         chan bool
}

func NewServer() *Server {
	activeSessions := make(map[string]*User)
	addUserCh := make(chan *User)
	removeUserCh := make(chan *User)
	doneCh := make(chan bool)

	return &Server{
		activeSessions,
		addUserCh,
		removeUserCh,
		doneCh,
	}
}

const (
	redirectURL       = "http://localhost:8090/callback"
	sessionCookieName = "sessionId"
)

var (
	auth = spotify.NewAuthenticator(redirectURL, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)

	// TODO maybe use session id instead
	state = "abc123"
	store = sessions.NewFilesystemStore("", []byte(os.Getenv("SESSION_KEY")))
)

func (server *Server) Listen() {
	log.Println("Listening...")

	http.HandleFunc("/", server.handleIndexPage)
	http.HandleFunc("/callback", server.handleSpotifyAuth)
	http.HandleFunc("/host", server.handleHostParty)

	for {
		select {
		case user := <-server.addUserCh:
			log.Debugf("User %s has been added.", user.id)
			server.activeSessions[user.id] = user
		case user := <-server.removeUserCh:
			log.Debugf("User %s has been removed", user.id)
			delete(server.activeSessions, user.id)
		case <-server.doneCh:
			return
		}
	}
}

func (server *Server) addUser(user *User) {
	server.addUserCh <- user
}

func (server *Server) removeUser(user *User) {
	server.removeUserCh <- user
}

func (server *Server) handleSpotifyAuth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		session, err := store.Get(r, sessionCookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := auth.Token(state, r)
		if err != nil {
			log.Debugf("Couldn't get token for state \"%s\"", state)
			http.Error(w, "Couldn't get token", http.StatusNotFound)
			return
		}

		client := auth.NewClient(token)
		user, err := client.CurrentUser()
		if err != nil {
			log.Fatal(err)
		}

		session.Values["userId"] = user.ID
		session.Save(r, w)

		server.addUser(&User{user.ID, &client})

		http.Redirect(w, r, "/", http.StatusFound)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (server *Server) handleIndexPage(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Url  string
		User *spotify.PrivateUser
	}{Url: auth.AuthURL(state)}

	client := server.getCurrentSpotifyClient(w, r)
	if client != nil {
		currentUser, _ := client.CurrentUser()
		data.User = currentUser
	}

	t := template.Must(template.ParseFiles("template/index.html"))
	t.Execute(w, &data)
}

func (server *Server) handleHostParty(w http.ResponseWriter, r *http.Request) {
	client := server.getCurrentSpotifyClient(w, r)

	party := NewParty(client)
	go party.Host()

	http.Redirect(w, r, fmt.Sprintf("/party/%s", party.id), http.StatusFound)
}

func (server *Server) getCurrentSpotifyClient(w http.ResponseWriter, r *http.Request) (*spotify.Client) {
	session, err := store.Get(r, sessionCookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	if user, ok := server.activeSessions[session.Values["userId"].(string)]; ok {
		return user.client
	}

	return nil
}

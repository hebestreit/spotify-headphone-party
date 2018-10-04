package party

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/zmb3/spotify"
	"html/template"
	"net/http"
	"os"
)

type Server struct {
	doneCh chan bool
	store  sessions.Store
}

func NewServer(store sessions.Store) *Server {
	doneCh := make(chan bool)

	return &Server{
		doneCh,
		store,
	}
}

const (
	sessionCookieName = "sessionId"
)

var (
	auth = spotify.NewAuthenticator(os.Getenv("SPOTIFY_REDIRECT_URL"), spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)
	// TODO maybe use session id instead
	state = "abc123"
)

func (server *Server) Listen(r *mux.Router) {
	log.Println("Listening...")

	r.HandleFunc("/", server.handleViewIndex)
	r.HandleFunc("/callback", server.handleSpotifyAuth)

	// TODO move routing and actions to separate controllers/party.go
	r.HandleFunc("/parties", server.handleCreateParty).Methods("POST")
	r.HandleFunc("/parties", server.handleListParties).Methods("GET")
	r.HandleFunc("/parties/{id}", server.handleViewParty).Methods("GET")
	r.HandleFunc("/parties/{id}", server.handleDeleteParty).Methods("DELETE")
	r.HandleFunc("/parties/{id}", server.handleJoinParty).Methods("PUT")

	for {
		select {
		case <-server.doneCh:
			return
		}
	}
}

// callback action which will be triggered after Spotify authentication
func (server *Server) handleSpotifyAuth(w http.ResponseWriter, r *http.Request) {
	session, err := server.store.Get(r, sessionCookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := auth.Token(state, r)
	if err != nil {
		log.Debug(err)
		http.Error(w, "Couldn't get token", http.StatusNotFound)
		return
	}

	client := auth.NewClient(token)

	privateUser, err := client.CurrentUser()
	if err != nil {
		panic(err)
	}

	user := CreateUser(privateUser.ID, token)

	session.Values["userId"] = user.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

// index page with login button
func (server *Server) handleViewIndex(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Url  string
		User *spotify.PrivateUser
	}{Url: auth.AuthURL(state)}

	user := server.getSessionUser(w, r)
	if user != nil {
		currentUser, _ := user.SpotifyClient().CurrentUser()
		data.User = currentUser
	}

	t := template.Must(template.ParseFiles("./template/index.html"))
	t.Execute(w, &data)
}

// action to create and host a new party
func (server *Server) handleCreateParty(w http.ResponseWriter, r *http.Request) {
	user := server.getSessionUser(w, r)

	party := CreateParty(user)
	go party.Host()

	json.NewEncoder(w).Encode(party)
}

// action to list all parties
func (server *Server) handleListParties(w http.ResponseWriter, r *http.Request) {
	// TODO return only active parties, group by HostUserID and sort by createdAt
	parties, _ := FindAll()
	json.NewEncoder(w).Encode(parties)
}

// action to view party
func (server *Server) handleViewParty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	party, err := FindParty(vars["id"])

	if err != nil {
		panic(err)
	}

	if party == nil {
		http.NotFound(w, r)
		return
	}

	log.Debugf("view %s", party.ID)

	user := server.getSessionUser(w, r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err = user.SpotifyClient().CurrentUser(); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	data := struct {
		Party *Party
		User  *User
	}{Party: party, User: user}

	t := template.Must(template.ParseFiles("./template/party.html"))
	t.Execute(w, &data)
}

// action to join a party
func (server *Server) handleJoinParty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	party, err := FindParty(vars["id"])

	if err != nil {
		panic(err)
	}

	if party == nil {
		http.NotFound(w, r)
		return
	}

	user := server.getSessionUser(w, r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err = user.SpotifyClient().CurrentUser(); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if party.HostUserID == user.ID {
		http.Error(w, "Conflict", http.StatusConflict)
		return
	}

	go party.Join(user)

	log.Debugf("join %s", party.ID)
}

// delete party
func (server *Server) handleDeleteParty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	party, err := FindParty(vars["id"])

	if err != nil {
		panic(err)
	}

	if party == nil {
		http.NotFound(w, r)
		return
	}

	user := server.getSessionUser(w, r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err = user.SpotifyClient().CurrentUser(); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if party.HostUserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	DeleteParty(party)
}

// retrieve user form current session by cookie
func (server *Server) getSessionUser(w http.ResponseWriter, r *http.Request) *User {
	session, err := server.store.Get(r, sessionCookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	if sessionUserId, ok := session.Values["userId"]; ok {
		user, err := FindUser(sessionUserId.(string))
		if err != nil {
			return nil
		}

		return user
	}

	return nil
}

package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	. "github.com/voutasaurus/event"
	"github.com/voutasaurus/event/database"
)

var (
	envDBUser = os.Getenv("EVENTS_DB_USER")
	envDBPass = os.Getenv("EVENTS_DB_PASS")
	envDBHost = os.Getenv("EVENTS_DB_HOST")
	envDBPort = os.Getenv("EVENTS_DB_PORT")
	envDBName = os.Getenv("EVENTS_DB_NAME")
	envDBRoot = os.Getenv("EVENTS_DB_ROOT")
	envDBMode = os.Getenv("EVENTS_DB_MODE")
)

func main() {
	log.SetPrefix("events: ")
	log.SetFlags(log.Llongfile)
	log.Println("starting")

	u := &database.URL{
		User: envDBUser,
		Pass: envDBPass,
		Host: envDBHost,
		Port: envDBPort,
		Name: envDBName,
		Root: envDBRoot,
		Mode: envDBMode,
	}

	db, err := database.NewDB(u.String())
	if err != nil {
		log.Fatal(err)
	}
	s := server{
		db: db,
	}

	http.HandleFunc("/log", serveLog)
	http.HandleFunc("/schedule", s.serveSchedule)
	http.HandleFunc("/", serveDefault)

	log.Println("serving on :9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}

type server struct {
	db *database.DB
}

func serveDefault(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Allow", "OPTIONS,GET")
	w.Write([]byte("Hello World"))
	// TODO: write an actual home / status page
}

func serveLog(w http.ResponseWriter, r *http.Request) {
	d, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(d))
}

func (s *server) serveSchedule(w http.ResponseWriter, r *http.Request) {
	u, err := user(r)
	if err != nil {
		httpError(w, http.StatusForbidden)
		return
	}

	w.Header().Add("Allow", "OPTIONS,POST")
	switch r.Method {
	case "POST":
		var event Event
		if err := json.NewDecoder(r.Body).Decode(event); err != nil {
			log.Println(err)
			httpError(w, http.StatusBadRequest)
			return
		}
		if err := s.db.AddEvent(u, &event); err != nil {
			log.Println(err)
			httpError(w, http.StatusInternalServerError)
			return
		}
	case "OPTIONS":
		return
	default:
		httpError(w, http.StatusMethodNotAllowed)
		return
	}
}

func user(r *http.Request) (*User, error) {
	s, err := r.Cookie("backplane.oauth2.session")
	if err != nil {
		return nil, err
	}
	b, err := base64.URLEncoding.DecodeString(s.Value)
	if err != nil {
		return nil, err
	}
	var u User
	if err := json.Unmarshal(b, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

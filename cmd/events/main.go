package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

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

var (
	errTimeout = errors.New("timeout attempting to reach event URL")
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

	go s.loop()

	http.HandleFunc("/log", serveLog)
	http.HandleFunc("/schedule", s.serveSchedule)
	http.HandleFunc("/", serveDefault)

	log.Println("serving on :9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}

type server struct {
	db *database.DB
}

func (s *server) loop() {
	for {
		if err := s.processEvents(); err != nil {
			log.Println("loop:", err)
		}
		time.Sleep(10 * time.Second)
	}
}

func (s *server) processEvents() error {
	ee, err := s.db.GetEvents()
	if err != nil {
		return err
	}
	for _, e := range ee {
		if err := s.processEvent(e); err != nil {
			log.Println("processEvent:", err)
		}
	}
	return nil
}

func (s *server) processEvent(e *Event) error {
	interval := 10 * time.Millisecond
	for n := 0; n < 20; n++ {
		resp, err := http.Get(e.What)
		if err != nil || resp.StatusCode >= 400 {
			interval = backoff(interval)
			continue
		}
		return nil
	}
	return errTimeout
}

func backoff(current time.Duration) time.Duration {
	time.Sleep(current)
	if current < 10*time.Second {
		return current * 2
	}
	return 10 * time.Second
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

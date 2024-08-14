package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/exp/maps"
)

var QRCodes = map[string]bool{
	"4nj92jh": true,
	"x5avydu": true,
	"yy6t87a": true,
	"b3v8kur": true,
	"gcphypp": true,
	"7zyy91t": true,
	"kicqj8p": true,
	"atfwhhb": true,
	"erb7e36": true,
	"2va3kpv": true,
}

func genericHttpHandler(w http.ResponseWriter, r *http.Request) {}

type TeamTime struct {
	TimeOfLog time.Time
	Team      string
}

type App struct {
	mx      sync.Mutex
	Teams   map[string]bool
	Records map[string][]TeamTime // QRcode to TeamTime
}

func (ap *App) ListTeams() []string {
	ap.mx.Lock()
	defer ap.mx.Unlock()
	return maps.Keys(ap.Teams)
}

func (ap *App) AddTeam(name string) {
	ap.mx.Lock()
	defer ap.mx.Unlock()
	ap.Teams[name] = true
}

func (ap *App) DeleteTeam(name string) {
	ap.mx.Lock()
	defer ap.mx.Unlock()
	delete(ap.Teams, name)
}

func (ap *App) ClearTeams() {
	ap.mx.Lock()
	defer ap.mx.Unlock()
	ap.Teams = map[string]bool{}
}

func (ap *App) RecordTeamTime(qrcode, team string) error {
	t0 := time.Now()
	rec := TeamTime{
		Team:      team,
		TimeOfLog: t0,
	}

	ap.mx.Lock()
	defer ap.mx.Unlock()

	slc, found := ap.Records[qrcode]
	if !found {
		return fmt.Errorf("qrcode not found: %s", qrcode)
	}

	ap.Records[qrcode] = append(slc, rec)
	return nil
}

func (ap *App) ClearRecords() {
	ap.mx.Lock()
	defer ap.mx.Unlock()

	ap.Records = map[string][]TeamTime{}
}

func (ap *App) ListRecords() map[string][]TeamTime {
	ap.mx.Lock()
	defer ap.mx.Unlock()

	res := map[string][]TeamTime{}

	for k, v := range ap.Records {
		res[k] = v[:]
	}

	return res
}

func (ap *App) HandleListRecord(w http.ResponseWriter, r *http.Request) {
	recs := ap.ListRecords()
	enc := json.NewEncoder(w)
	err := enc.Encode(recs)
	if err != nil {
		log.Print("list records marshalling: sw", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (ap *App) HandleClearRecords(w http.ResponseWriter, r *http.Request) {
	ap.ClearRecords()
	fmt.Fprint(w, "cleared")
}

func (ap *App) HandleListTeams(w http.ResponseWriter, r *http.Request) {
	teams := ap.ListTeams()
	enc := json.NewEncoder(w)
	err := enc.Encode(teams)
	if err != nil {
		log.Printf("list team marshalling: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (ap *App) HandleAddTeam(w http.ResponseWriter, r *http.Request) {
	tm := r.URL.Query().Get("team")
	if tm == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `required arg "team"`)
	}

	ap.AddTeam(tm)
	fmt.Fprintf(w, `team added %q`, tm)
}

func (ap *App) HandleClearTeams(w http.ResponseWriter, r *http.Request) {
	ap.ClearTeams()

	fmt.Fprint(w, "teams cleared")
}

func (ap *App) HandleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	tm := r.URL.Query().Get("team")
	if tm == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `required arg "team"`)
	}

	ap.DeleteTeam(tm)
	fmt.Fprintf(w, `team %q succesfully deleted`, tm)
}

func main() {
	log.Println("Server started")

	app := &App{
		Teams:   map[string]bool{},
		Records: map[string][]TeamTime{},
	}

	http.DefaultServeMux.HandleFunc("/list-records", app.HandleListRecord)
	http.DefaultServeMux.HandleFunc("/clear-records", app.HandleClearRecords)
	http.DefaultServeMux.HandleFunc("/list-teams", app.HandleListTeams)
	http.DefaultServeMux.HandleFunc("/add-team", app.HandleAddTeam)
	http.DefaultServeMux.HandleFunc("/delete-team", app.HandleDeleteTeam)
	http.DefaultServeMux.HandleFunc("/clear-teams", app.HandleClearTeams)

	// TODO: Add handle to accept QR codes - probably main handle
	// TODO: Add handle to draw recorded time of found QR

	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("handled message to: %s", r.URL.Path)
		fmt.Fprint(w, "gt5 night game")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), http.DefaultServeMux)
	if err != nil {
		log.Printf("shutting server - %s", err)
	}

	log.Print("server shuttng down ok")
}

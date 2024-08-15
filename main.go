package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
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
	mx             sync.Mutex
	Teams          map[string]bool
	Records        map[string][]TeamTime // QRcode to TeamTime
	RecordTimeTmpl *template.Template
	QRCodeTmpl     *template.Template
}

func (ap *App) loadTemplates() error {
	{
		tmpl, err := template.ParseFiles("assets/web/record-time.html.tmpl")
		if err != nil {
			return err
		}

		ap.RecordTimeTmpl = tmpl
	}

	{
		tmpl, err := template.ParseFiles("assets/web/qrcode.html.tmpl")
		if err != nil {
			return err
		}

		ap.QRCodeTmpl = tmpl
	}

	return nil
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

	if _, found := QRCodes[qrcode]; !found {
		return fmt.Errorf("qrcode not found: %s", qrcode)
	}

	slc := ap.Records[qrcode]
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

type RecordTimeTmplData struct {
	Team string
}

func (ap *App) HandleRecordTime(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("record time parsing form: %s", err)
	}

	tm := r.Form.Get("team")
	if tm == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `required arg "team"`)
	}

	qr := r.Form.Get("qrcode")
	if qr == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `required arg "qrcode"`)
	}

	err = ap.RecordTeamTime(qr, tm)
	if err != nil {
		log.Printf(`recording team time for team %q code %q: %s`, tm, qr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// renderer html template success
	ap.RecordTimeTmpl.Execute(w, &RecordTimeTmplData{
		Team: tm,
	})
}

var qrcodeReStr = fmt.Sprintf(`^/(%s)$`, strings.Join(maps.Keys(QRCodes), "|"))
var qrcodeRe = regexp.MustCompile(qrcodeReStr)

type QRCodeTemplData struct {
	QRCode string
	Teams  []string
}

func main() {
	log.Println("Server started")

	app := &App{
		Teams:   map[string]bool{},
		Records: map[string][]TeamTime{},
	}

	err := app.loadTemplates()
	if err != nil {
		log.Printf("loading templates: %s", err)
		panic("failed loading templates")
	}
	log.Print("templates loaded")

	http.DefaultServeMux.HandleFunc("/list-records", app.HandleListRecord)
	http.DefaultServeMux.HandleFunc("/clear-records", app.HandleClearRecords)
	http.DefaultServeMux.HandleFunc("/list-teams", app.HandleListTeams)
	http.DefaultServeMux.HandleFunc("/add-team", app.HandleAddTeam)
	http.DefaultServeMux.HandleFunc("/delete-team", app.HandleDeleteTeam)
	http.DefaultServeMux.HandleFunc("/clear-teams", app.HandleClearTeams)
	http.DefaultServeMux.HandleFunc("/record-time", app.HandleRecordTime)

	http.DefaultServeMux.HandleFunc("/healthcheck", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "gt5")
	})

	// TODO: Add handle to draw recorded time of found QR

	log.Printf("qrregexp %q", qrcodeReStr)

	// TODO: Add handle to accept QR codes - probably main handle
	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.EscapedPath()
		log.Printf("handled message to: %s", path)

		if !qrcodeRe.MatchString(path) {
			fmt.Fprint(w, "nothing here")
			return
		}

		qrcode, _ := strings.CutPrefix(path, "/")
		teams := app.ListTeams()

		err := app.QRCodeTmpl.Execute(w, &QRCodeTemplData{
			QRCode: qrcode,
			Teams:  teams,
		})

		if err != nil {
			log.Printf("executing template: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), http.DefaultServeMux)
	if err != nil {
		log.Printf("shutting server - %s", err)
	}

	log.Print("server shuttng down ok")
}

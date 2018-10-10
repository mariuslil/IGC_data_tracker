package main

import (
	"errors"
	"fmt"
	"github.com/marni/goigc"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

var startTime time.Time
var templates = template.Must(template.ParseFiles("edit.html", "view.html", "apiTemplate.html"))
var validPath = regexp.MustCompile("^/igcinfo/(api)/$|\0121(igc)/$|\0122([0-9][0-9][0-9][0-9])/$|\0123(pilot|glider_id|glider|calculated_total_track_length|H_date)/$")

type Page struct {
	Title string
	Body  []byte
}

type data struct {
	uptime int `json:"uptime"`
	info string `json:"info"`
	version string `json:"version"`
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/api/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "apiTemplate", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

func uptime() time.Duration{
	return time.Since(startTime)
}

func main() {

	startTime = time.Now()

	s := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
	track, err := igc.ParseLocation(s)
	if err != nil {
		fmt.Errorf("Problem reading the track", err)
	}

	fmt.Printf("Pilot: %s, gliderType: %s, date: %s",
		track.Pilot, track.GliderType, track.Date.String())

	http.HandleFunc("/igcinfo/", viewHandler)



	log.Fatal(http.ListenAndServe(":8080", nil))


}
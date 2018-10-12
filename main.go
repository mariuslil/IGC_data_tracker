package main

import (
	"encoding/json"
	"fmt"
	"github.com/marni/goigc"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)


var startTime time.Time
var tracks map[int]string
var trackinf map[int]TrackInfo
var numTracks int
//var validPath = regexp.MustCompile("^/igcinfo/(api)/$|\0121(igc)/$|\0122([0-9][0-9][0-9][0-9])/$|\0123(pilot|glider_id|glider|calculated_total_track_length|H_date)/$")
var validField = regexp.MustCompile("^(pilot|glider_id|glider|calculated_total_track_length|H_date)$")


type MetaData struct {
	Uptime string `json:"uptime"`
	Info string `json:"info"`
	Version string `json:"version"`
}


type TrackInfo struct {
	H_date string `json:"H_date"`
	Pilot string `json:"pilot"`
	Glider string `json:"glider"`
	Glider_id string `json:"glider_id"`
	Track_length json.Number `json:"track_length"`
}


 func apiHandler(w http.ResponseWriter, r *http.Request){
 	w.Header().Set("Content-Type", "application/json")

 	endTime := time.Now()

 	data := MetaData{
 		Uptime: fmt.Sprintf("P%dY%dM%dDT%dH%dM%dS",
		endTime.Year() - startTime.Year(), endTime.Month() - startTime.Month(), endTime.Day() - startTime.Day(),
		endTime.Hour() - startTime.Hour(), endTime.Minute() - startTime.Minute(), endTime.Second() - startTime.Second()),
	 	Info: "Service for IGC tracks.",
	 	Version: "V1",
	 }

	 jsonEnc := json.NewEncoder(w)
	 print(jsonEnc.Encode(data))

 }

func igcHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	URL := strings.Split(r.URL.Path, "/")

	if URL[5] != "" && r.Method == "GET"{

		i, err := strconv.Atoi(URL[4])
		if err != nil{
			http.Error(w, "Something went wrong", http.StatusBadRequest)
			return
		}

		field(w, r, i, URL[5])

	} else if URL[4] != "" && r.Method == "GET"{

		u, err := strconv.Atoi(URL[4])
		if err != nil{
			http.Error(w, "Something went wrong", http.StatusBadRequest)
			return
		}

		for i := 0; i <= numTracks; i++{
			if u == i{
				trackinf[i] = TrackInfo{
					H_date:       fmt.Sprintf(time.Now().String()),
					Pilot:        "Gary",
					Glider:       "G-" + strconv.Itoa(rand.Intn(500)),
					Glider_id:    strconv.Itoa(rand.Intn(100)),
					Track_length: json.Number(rand.Intn(10000)),
				}
				jsonEnc := json.NewEncoder(w)
				print(jsonEnc.Encode(trackinf[i]))
				return
			}
		}
		http.Error(w, "NOT FOUND", http.StatusBadRequest)


	} else if r.Method == "POST"{

		type request struct {
			Url string `json:"url"`
		}

		type response struct {
			Id string `json:"id"`
		}

		decoder := json.NewDecoder(r.Body)
		var urlObj request

		err := decoder.Decode(&urlObj)
		if err != nil{
			http.Error(w, "Something went wrong", http.StatusBadRequest)
			return
		}

		tracks[numTracks] = urlObj.Url

		resp := response{Id: string(numTracks)}

		numTracks++

		jsonEnc := json.NewEncoder(w)
		print(jsonEnc.Encode(resp))


	} else if r.Method == "GET"{

		var array []int

		for i := 0; i < numTracks; i++ {
			array = append(array, i)
		}

		jsonEnc := json.NewEncoder(w)
		print(jsonEnc.Encode(array))

	} else {
		http.Error(w, "Not a valid request", http.StatusBadRequest)
		return
	}

}

func field (w http.ResponseWriter, r *http.Request, id int, field string){
	w.Header().Set("Content-Type", "text/plain")
	switch field {
	case "h_date":
		print(trackinf[id].H_date)
	case "pilot":
		print(trackinf[id].Pilot)
	case "glider":
		print(trackinf[id].Glider)
	case "glider_id":
		print(trackinf[id].Glider_id)
	case "track_length":
		print(trackinf[id].Track_length)
	default:
		http.Error(w, "Invalid Field", http.StatusBadRequest)
		return
	}
}


func main() {
	startTime = time.Now()
	tracks = make(map[int]string)
	trackinf = make(map[int]TrackInfo)
	numTracks = 0

	s := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
	track, err := igc.ParseLocation(s)
	if err != nil {
		fmt.Errorf("Problem reading the track", err)
	}

	fmt.Printf("Pilot: %s, gliderType: %s, date: %s\n",
		track.Pilot, track.GliderType, track.Date.String())

	http.HandleFunc("/igcinfo/api/", apiHandler)
	http.HandleFunc("/igcinfo/api/igc/", igcHandler)
	http.ListenAndServe(":8080", nil)
}
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	storm "github.com/asdine/storm/v3"
	"github.com/gorilla/mux"
)

type Stream struct {
	ID          int      `json:"id" storm:"id,increment"`
	Name        string   `json:"name"`
	Token       string   `json:"token" storm:"unique"`
	TwitchToken string   `json:"twitchToken"`
	Paths       []string `json:"paths"`
	PlayUrl     string   `json:"playUrl"`
}

type NginxStream struct {
}

type allStreams []Stream

var db *storm.DB
var err error

func createStream(w http.ResponseWriter, r *http.Request) {
	var newStream Stream

	//Get Stream data
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(reqBody, &newStream)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Generate unique token
	newStream.Token, err = generateRandomString(30)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Save stream in db with token
	err = db.Save(&newStream)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send back the new stream with generated token
	json.NewEncoder(w).Encode(newStream)
}

func getAllStreams(w http.ResponseWriter, r *http.Request) {
	var streams []Stream
	err := db.All(&streams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get all streams
	json.NewEncoder(w).Encode(streams)
}

func updateStream(w http.ResponseWriter, r *http.Request) {
	var updatedStream Stream

	// Get Stream data
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	json.Unmarshal(reqBody, &updatedStream)

	// Update stream in DB
	err = db.Update(&updatedStream)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteStream(w http.ResponseWriter, r *http.Request) {
	var stream Stream

	// Get the ID of the stream to delete
	streamID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = db.One("ID", streamID, &stream)
	if err == storm.ErrNotFound {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	for _, n := range stream.Paths {
		err = os.Remove("/tmp/rec/" + n)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Delete stream from DB
	err = db.DeleteStruct(&Stream{ID: streamID})
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func canLive(w http.ResponseWriter, r *http.Request) {
	var stream Stream
	r.ParseForm()
	token := r.FormValue("name")
	// Check if token is in streaming list
	err = db.One("Token", token, &stream)
	if err == storm.ErrNotFound {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	// Get the backend IP
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req := r.URL.Query()
	if stream.TwitchToken != "" {
		req.Add("twitch_key", stream.TwitchToken)

	}

	escapedPath := replaceIllegalFilepath(stream.Name)
	req.Set("name", stream.Name)
	r.URL.RawQuery = req.Encode()

	stream.PlayUrl = strconv.Itoa(stream.ID) + "-" + escapedPath

	err = db.Update(&stream)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	http.Redirect(w, r, "rtmp://"+ip+"/hls-live/"+stream.PlayUrl, 302)
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func replaceIllegalFilepath(path string) string {
	var re = regexp.MustCompile(`[\/\\?%*:|"<>. ;,^{}$#]`)
	s := re.ReplaceAllString(path, "-")
	return s
}

func onPublishDone(w http.ResponseWriter, r *http.Request) {
	var stream Stream
	r.ParseForm()
	token := r.FormValue("name")
	// Check if token is in streaming list
	err = db.One("Token", token, &stream)
	if err == storm.ErrNotFound {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	matches, err := filepath.Glob(strconv.Itoa(stream.ID) + "-*")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stream.Paths = matches

	err = db.Update(&stream)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteVideo(w http.ResponseWriter, r *http.Request) {
	var stream Stream
	// Get the ID of the stream to delete
	streamID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	path := mux.Vars(r)["path"]

	// Delete stream from DB
	err = db.One("ID", streamID, &stream)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	for _, n := range stream.Paths {
		if n == path {
			err = os.Remove("/tmp/rec/" + path)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			matches, err := filepath.Glob(strconv.Itoa(stream.ID) + "-*")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			stream.Paths = matches
			err = db.Update(&stream)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func main() {

	db, err = storm.Open("streye.db")
	if err != nil {
		panic(err)
		return
	}
	defer db.Close()
	err = os.Chdir("/tmp/rec")
	if err != nil {
		panic(err)
		return
	}
	log.Println("Streye backend up")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink).Methods("GET")
	router.HandleFunc("/live", canLive).Methods("POST")
	router.HandleFunc("/done", onPublishDone).Methods("POST")
	router.HandleFunc("/stream", createStream).Methods("POST")
	router.HandleFunc("/stream", updateStream).Methods("PATCH")
	router.HandleFunc("/streams", getAllStreams).Methods("GET")
	router.HandleFunc("/streams/{id}", deleteStream).Methods("DELETE")
	router.HandleFunc("/streams/{id}/{path}", deleteVideo).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":3000", router))
}

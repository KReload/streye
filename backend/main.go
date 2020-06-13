package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/asdine/storm/v3"
	"github.com/gorilla/mux"
)

type Stream struct {
	ID          int    `json:"id" storm:"id,increment"`
	Name        string `json:"name"`
	Token       string `json:"token" storm:"unique"`
	TwitchToken string `json:"twitchToken"`
}

type allStreams []Stream

var db *storm.DB
var err error

func createStream(w http.ResponseWriter, r *http.Request) {
	var newStream Stream
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

	newStream.Token, err = generateRandomString(30)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.Save(&newStream)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(newStream)
}

func getAllStreams(w http.ResponseWriter, r *http.Request) {
	var streams []Stream
	err := db.All(&streams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(streams)
}

func updateStream(w http.ResponseWriter, r *http.Request) {
	var updatedStream Stream

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	json.Unmarshal(reqBody, &updatedStream)
	err = db.Update(&updatedStream)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteStream(w http.ResponseWriter, r *http.Request) {
	streamID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = db.DeleteStruct(&Stream{ID: streamID})
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func canLive(w http.ResponseWriter, r *http.Request) {
	var stream Stream
	token := mux.Vars(r)["token"]

	err = db.One("Token", token, &stream)
	if err == storm.ErrNotFound {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func homeLink(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	db, err = storm.Open("streye.db")
	if err != nil {
		panic(err)
		return
	}
	defer db.Close()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink).Methods("GET")
	router.HandleFunc("/live/{token}", canLive).Methods("GET")
	router.HandleFunc("/stream", createStream).Methods("POST")
	router.HandleFunc("/stream", updateStream).Methods("PATCH")
	router.HandleFunc("/streams", getAllStreams).Methods("GET")
	router.HandleFunc("/streams/{id}", deleteStream).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":3000", router))
}

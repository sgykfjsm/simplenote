package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	sm "github.com/sgykfjsm/simplenote"
)

var (
	client   *sm.Client
	email    = "Your Email"
	password = "Your simplenote Password"
)

func SimplenoteCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("%s is not allowed", r.Method)))
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprint(http.StatusText(http.StatusInternalServerError))))
		return
	}

	content := r.PostFormValue("content")
	note, err := client.Add(content, []string{})
	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Err.Error()))
		return
	}

	res, _ := json.Marshal(note)
	fmt.Fprint(w, string(res))
}

func SimplenoteGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("%s is not allowed", r.Method)))
		return
	}

	noteKey := r.URL.Query().Get("key")
	note, err := client.Get(noteKey)
	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Err.Error()))
		return
	}

	res, _ := json.Marshal(note)
	fmt.Fprint(w, string(res))
}

func SimplenoteIndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("%s is not allowed", r.Method)))
		return
	}

	limit := 10
	paramLength := r.URL.Query().Get("length")
	if paramLength != "" {
		limit, _ = strconv.Atoi(paramLength)
	}

	since := r.URL.Query().Get("since")
	mark := r.URL.Query().Get("mark")

	index, err := client.Index(limit, since, mark)
	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Err.Error()))
		return
	}
	log.Printf("Return %d index", index.Count)
	res, _ := json.Marshal(index)
	fmt.Fprint(w, string(res))
}

func SimplenoteUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("%s is not allowed", r.Method)))
		return
	}

	key := r.URL.Query().Get("key")
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprint(http.StatusText(http.StatusInternalServerError))))
		return
	}

	content := r.PostFormValue("content")
	note, err := client.Update(key, content, false)
	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Err.Error()))
		return
	}

	res, _ := json.Marshal(note)
	fmt.Fprint(w, string(res))
}

func SimplenoteDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("%s is not allowed", r.Method)))
		return
	}

	var note sm.Data
	var err *sm.ErrorResponse
	key := r.URL.Query().Get("key")
	if r.Method == http.MethodDelete {
		err = client.Delete(key)
	} else { // http.MethodUpdate
		note, err = client.Update(key, "", true)
	}

	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Err.Error()))
		return
	}

	result := http.StatusText(http.StatusOK)
	if note.Key != "" {
		res, _ := json.Marshal(note)
		result = string(res)
	}
	fmt.Fprint(w, result)
}

func main() {
	var err *sm.ErrorResponse
	debug := true
	client, err = sm.New(email, password, debug)
	if err != nil {
		log.Print("Login failed. Confirm your email and password")
		log.Fatal(err.Err.Error())
	}
	// TODO: Add timer to refresh token per 24h
	log.Printf("Login Suceeded. Token: %s", client.Token)

	http.HandleFunc("/simplenote/api/create", SimplenoteCreateHandler)
	http.HandleFunc("/simplenote/api/get", SimplenoteGetHandler)
	http.HandleFunc("/simplenote/api/index", SimplenoteIndexHandler)
	http.HandleFunc("/simplenote/api/update", SimplenoteUpdateHandler)
	http.HandleFunc("/simplenote/api/delete", SimplenoteDeleteHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

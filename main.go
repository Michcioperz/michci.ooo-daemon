package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var validProjects *map[string]string = nil

func panicIfErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func loadConfig() (config *map[string]string) {
	config = new(map[string]string)
	f, err := os.Open("repository-secrets.json")
	panicIfErr(err)
	defer func() { panicIfErr(f.Close()) }()
	dec := json.NewDecoder(f)
	panicIfErr(dec.Decode(config))
	return
}

func handle(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.User.Username()
	token, _ := r.URL.User.Password()
	validToken, repoExists := (*validProjects)[repo]
	if !repoExists || (token != validToken) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusCreated)
	cmd := exec.Command("./build_project.sh", repo)
	cmd.Stdout = w
	log.Printf("run of %v finished: %v", repo, cmd.Run())
}

func main() {
	go func() {
		for {
			validProjects = loadConfig()
			time.Sleep(60)
		}
	}()
	for validProjects == nil {
	}
	http.HandleFunc("/", handle)
	log.Panic(http.ListenAndServe(":31400", nil))
}

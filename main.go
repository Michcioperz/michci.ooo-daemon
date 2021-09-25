package main

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	repo, token, _ := r.BasicAuth()
	validToken, repoExists := (*validProjects)[repo]
	if !repoExists || (token != validToken) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusCreated)

	cmd := exec.Command("./build_project.sh", repo)
	out, err := cmd.StdoutPipe()
	panicIfErr(err)
	cmd.Stderr = cmd.Stdout

	flusher, isFlusher := w.(http.Flusher)
	logger := log.New(os.Stderr, fmt.Sprintf("[%v] ", repo), 0)
	logger.Printf("run started: %v", cmd.Start())
	buf := bufio.NewScanner(out)
	for buf.Scan() {
		logger.Print(buf.Text())
		fmt.Fprintln(w, buf.Text())
		if isFlusher {
			flusher.Flush()
		}
	}
	panicIfErr(buf.Err())
	logger.Printf("run finished: %v", cmd.Wait())
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

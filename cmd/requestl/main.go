package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/etowett/requestl/build"
)

func main() {
	logFile := os.Getenv("LOG_FILE")
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		wrt := io.MultiWriter(os.Stdout, f)
		log.SetOutput(wrt)
	}

	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/", handleRequest)

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "7000"
	}

	log.Printf("Server starting, listening on :%v", serverPort)
	http.ListenAndServe(fmt.Sprintf(":%v", serverPort), nil)
}

func printRequest(r *http.Request) error {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err
	}

	log.Printf("\nBody:\n %+v \n", string(dump))
	log.Printf("\n===================================================\n")
	return nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	err := printRequest(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error dumping request: %v", err), http.StatusInternalServerError)
		return
	}

	theResponse := map[string]interface{}{
		"success": true,
		"status":  "Ok",
	}

	jsResp, err := json.Marshal(theResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsResp)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	err := printRequest(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error dumping request: %v", err), http.StatusInternalServerError)
		return
	}

	theResponse := map[string]interface{}{
		"success":    true,
		"status":     "Ok",
		"sha1ver":    build.Sha1Ver,
		"build_time": build.Time,
		"git_commit": build.GitCommit,
		"git_branch": build.GitBranch,
		"version":    build.Version,
	}

	jsResp, err := json.Marshal(theResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsResp)
}

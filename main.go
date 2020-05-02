package main

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "net/http/httputil"
    "os"
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

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        dump, err := httputil.DumpRequest(r, true)
        if err != nil {
            http.Error(w, fmt.Sprintf("Error dumping request: %v", err), http.StatusInternalServerError)
            return
        }

        log.Printf("\nBody:\n %+v \n", string(dump))
        log.Printf("\n===================================================\n")

        fmt.Fprintf(w, "Ok")
    })

    serverPort := os.Getenv("PORT")
    if serverPort == "" {
        serverPort = "7000"
    }

    log.Printf("Server starting, listening on :%v", serverPort)
    http.ListenAndServe(fmt.Sprintf(":%v", serverPort), nil)
}

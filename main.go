package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

var notifyChan chan struct{}
var globalStore int = 0

func sub(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("He2")
	// flusher, ok := rw.(http.Flusher)

	// if !ok {
	// 	http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
	// 	return
	// }
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	for {
		time.Sleep(time.Second * 5)
		fmt.Println("ay 7aga")
		fmt.Fprintf(rw, "signal %v\n\n", 33)

		if f, ok := rw.(http.Flusher); ok {
			fmt.Println("flushena ?")
			f.Flush()
		}
	}

}

func notify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if notifyChan != nil && globalStore != 0 {
		notifyChan <- struct{}{}
	}
}

func fetch(w http.ResponseWriter, r *http.Request) {
	temp := globalStore
	globalStore = 0
	fmt.Fprintf(w, "Current data : %d\n", temp)
}

func addData(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	fmt.Println(data)
	if err != nil {
		fmt.Println(err)
	}
	num, _ := strconv.Atoi(string(data))
	globalStore = num

	notifyChan <- struct{}{}

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/sub", sub)
	mux.HandleFunc("/add", addData)
	mux.HandleFunc("/fetch", fetch)
	http.ListenAndServe(":5000", mux)
}

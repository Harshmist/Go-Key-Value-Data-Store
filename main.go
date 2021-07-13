package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var dta = make(map[int]string, 100)

func listData(w http.ResponseWriter, r *http.Request) {

	for i := 1; i < len(dta)+1; i++ {
		fmt.Fprintf(w, strconv.FormatInt(int64(i), 10)+": ")
		fmt.Fprintf(w, dta[i]+"\n")
	}

}

func dataMethods(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}

}
func getData(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	key, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, dta[key]+"\n")
}

func post(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Allowing POST to take multiple words seperated by hyphens "-"
	var value string
	valArr := strings.Split(parts[2], "-")
	if len(valArr) > 1 {
		value = strings.Join(valArr, " ")
	} else {
		value = parts[2]
	}
	dta[len(dta)+1] = string(value)
}

func main() {
	// Placeholder data for testing
	dta[1] = "David"
	dta[2] = "Ryan"
	dta[3] = "Craig"
	dta[4] = "Liam"

	http.HandleFunc("/list", listData)
	http.HandleFunc("/get/", getData)
	http.HandleFunc("/post/", post)
	http.ListenAndServe(":8080", nil)
}

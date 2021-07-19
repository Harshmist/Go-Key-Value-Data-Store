package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

var dta = make(map[int]string, 100)
var newDta = make(map[int]string, 100)

func listData(w http.ResponseWriter, r *http.Request) {

	for i := 1; i < len(dta)+1; i++ {
		if dta[i] != "" {
			fmt.Fprintf(w, strconv.FormatInt(int64(i), 10)+": ")
			fmt.Fprintf(w, dta[i]+"\n")
		}
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
	if dta[key] == "" {
		fmt.Fprintf(w, fmt.Sprint(key)+" empty")
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, dta[key]+"\n")
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Allowing POST to take multiple words seperated by underscores "_"
	var value string
	valArr := strings.Split(parts[2], "_")
	if len(valArr) > 1 {
		value = strings.Join(valArr, " ")
	} else {
		value = parts[2]
	}
	dta[len(dta)+1] = string(value)
}

// Delete specified ID value
func delete(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key, err := strconv.Atoi(parts[2])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, dta[key]+" deleted")
	dta[key] = ""

}

//Handling the commands via TCP
func handler(conn net.Conn) {

	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}

		switch fields[0] {
		case "GET":
			if len(fields) < 2 {
				io.WriteString(conn, "No key input! \n")
			}
			keyInt, err := strconv.Atoi(fields[1])
			if err != nil {
				log.Fatal("Fatal error")
			}
			io.WriteString(conn, dta[keyInt]+"\n")
		case "POST":
			var value string
			if len(fields) < 2 {
				io.WriteString(conn, "No value added! \n")
			}
			fieldArr := strings.Split(fields[1], "_")
			if len(fieldArr) > 1 {
				value = strings.Join(fieldArr, " ")
			} else {
				value = fields[1]
			}
			dta[len(dta)+1] = string(value)
			io.WriteString(conn, value+" added!\n")
		case "LIST":
			for i := 0; i < len(dta)+1; i++ {
				if dta[i] != "" {
					io.WriteString(conn, fmt.Sprint(i)+": "+dta[i]+"\n")
				}
			}

		case "DELETE":
			keyInt, err := strconv.Atoi(fields[1])
			if err != nil {
				log.Fatal("Fatal error")
			}
			dta[keyInt] = ""
		default:
			io.WriteString(conn, "Invalid Command "+fields[0]+"\n")
		}
	}

}

//Starting the TCP connection
func startTCP() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		handler(conn)

	}

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
	http.HandleFunc("/delete/", delete)
	go startTCP()
	http.ListenAndServe(":8080", nil)
}

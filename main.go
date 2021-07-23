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

var data = make(map[int]string)
var postChannel = make(chan string)
var intChannel = make(chan int)
var setChannel = make(chan string)
var delChannel = make(chan int)

func postData() {
	for {
		value := <-postChannel
		data[len(data)+1] = value
	}
}

func setData() {

	for {
		key := <-intChannel
		value := <-setChannel
		data[key] = value
	}
}

func delData() {
	for {
		key := <-delChannel
		data[key] = ""
	}
}

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
			postChannel <- string(value)
			io.WriteString(conn, value+" added!\n")
		case "LIST":
			for i := 0; i < len(data)+1; i++ {
				if data[i] != "" {
					io.WriteString(conn, fmt.Sprint(i)+": "+data[i]+"\n")
				}
			}
		case "SET":
			if len(fields) < 3 {
				io.WriteString(conn, "Format should be <int Key> <string Value> \n")
			}
			var value string
			keyInt, err := strconv.Atoi(fields[1])
			if err != nil {
				log.Fatal("Fatal error")
			}
			fieldArr := strings.Split(fields[2], "_")
			if len(fieldArr) > 1 {
				value = strings.Join(fieldArr, " ")
			} else {
				value = fields[2]
			}
			if data[keyInt] != "" {
				io.WriteString(conn, "Key is already in use. Try the LIST command to see keys in use \n")
			}
			intChannel <- keyInt
			setChannel <- value
		case "DELETE":
			keyInt, err := strconv.Atoi(fields[1])
			if err != nil {
				log.Fatal("Fatal error")
			}
			delChannel <- keyInt
		default:
			io.WriteString(conn, "Invalid Command "+fields[0]+"\n")
		}

	}

}

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

		go handler(conn)

	}

}

func listData(w http.ResponseWriter, r *http.Request) {

	for i := 1; i < len(data)+1; i++ {
		if data[i] != "" {
			fmt.Fprintf(w, strconv.FormatInt(int64(i), 10)+": ")
			fmt.Fprintf(w, data[i]+"\n")
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
	if data[key] == "" {
		fmt.Fprintf(w, fmt.Sprint(key)+" empty")
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, data[key]+"\n")
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
	postChannel <- value
}

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
	fmt.Fprintf(w, data[key]+" deleted")
	delChannel <- key

}

func main() {

	//Placeholder data for testing
	data[1] = "Dan"
	data[2] = "Sinead"

	go postData()
	go setData()
	go delData()
	go startTCP()

	http.HandleFunc("/list", listData)
	http.HandleFunc("/get/", getData)
	http.HandleFunc("/post/", post)
	http.HandleFunc("/delete/", delete)
	http.ListenAndServe(":8080", nil)

}

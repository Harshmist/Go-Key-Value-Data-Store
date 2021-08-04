package main

import (
	"bufio"
	"expvar"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	data        = make(map[int]string)
	postChannel = make(chan string)
	setChannel  = make(chan [2]string)
	delChannel  = make(chan int)
	logger      *log.Logger
	startTime   time.Time
	counts      = expvar.NewMap(("counters"))
)

func init() {
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)

}

// TCP Handler to take commands and send data to channels

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
			logger.Println("full list of store requested")
			counts.Add("Requests", 1)
		case "SET":
			if len(fields) < 3 {
				io.WriteString(conn, "Format should be <int Key> <string Value> \n")
			}
			var chanSlice [2]string
			var value string
			keyInt, err := strconv.Atoi(fields[1])
			if err != nil {
				log.Fatal("Fatal error")
			}
			chanSlice[0] = fields[1]
			fieldArr := strings.Split(fields[2], "_")
			if len(fieldArr) > 1 {
				value = strings.Join(fieldArr, " ")
			} else {
				value = fields[2]
			}
			if data[keyInt] != "" {
				io.WriteString(conn, "Key is already in use. Try the LIST command to see keys in use \n")
			} else {
				chanSlice[1] = value
				setChannel <- chanSlice
			}
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

//HTTP HandleFuncs

func listData(w http.ResponseWriter, r *http.Request) {

	for i := 1; i < len(data)+1; i++ {
		if data[i] != "" {
			fmt.Fprintf(w, strconv.FormatInt(int64(i), 10)+": ")
			fmt.Fprintf(w, data[i]+"\n")
		}
	}
	logger.Println("full request of data in store")
	counts.Add("Requests", 1)

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
		counts.Add("Requests", 1)
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
	counts.Add("List Size", int64(len(data)))
	counts.Add("Requests", 0)

	go startTCP()

	//Receiving input via channels and manipulate the data map
	go func() {
		for {
			select {
			case val1 := <-postChannel:
				data[len(data)+1] = val1
				logger.Printf("%v added to store at key %d\n", val1, len(data))
				counts.Add("List Size", 1)
				counts.Add("Requests", 1)
			case val2 := <-setChannel:
				key, err := strconv.Atoi(val2[0])
				if err != nil {
					log.Fatal("Fatal error")
				}
				data[key] = val2[1]
				logger.Printf("%v added to store at key %d\n", val2[1], key)
				counts.Add("List Size", 1)
				counts.Add("Requests", 1)
			case delKey := <-delChannel:
				data[delKey] = ""
				logger.Printf("key %d removed from store\n", delKey)
				counts.Add("List Size", -1)
				counts.Add("Requests", 1)
			}
		}
	}()

	//HTTP Handlers
	http.HandleFunc("/list", listData)
	http.HandleFunc("/get/", getData)
	http.HandleFunc("/post/", post)
	http.HandleFunc("/delete/", delete)
	http.ListenAndServe(":8123", nil)

}

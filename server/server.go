package main

import (
	"encoding/binary"
	"fmt"
	//"github.com/gordonklaus/portaudio"
	"log"
	"net/http"
	"os"
	"strconv"
)

const sampleRate = 44100

type user struct {
	name string
	chatWriter http.ResponseWriter
	audioWriter http.ResponseWriter
}

var users map[string]user

func main() {
	//portaudio.Initialize()
	//defer portaudio.Terminate()
	users = map[string]user{}
	var bufLen string
	if len(os.Args) < 2 {
		fmt.Println("How many seconds long should the buffer be (values lower than 0.03 result in low audio quality delay will be at least 2*length):")
		fmt.Scanln(&bufLen)
	} else {
		bufLen = os.Args[1]
	}
	var seconds, _ = strconv.ParseFloat(bufLen, 64)
	buffer := make([]float32, int64(sampleRate * seconds))
	//stream, err := portaudio.OpenDefaultStream(1, 0, sampleRate, len(buffer), func(in []float32) {
	//	for i := range buffer {
	//		buffer[i] = in[i]
	//	}
	//})
	//chk(err)
	//chk(stream.Start())
	//defer stream.Close()

	http.HandleFunc("/audio", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("expected http.ResponseWriter to be an http.Flusher")
		}
		userData, exists := users[r.RemoteAddr]
		if exists {
			userData.audioWriter = w
			users[r.RemoteAddr] = userData
		} else {
			users[r.RemoteAddr] = user{r.RemoteAddr, nil, w}
		}
		w.Header().Set("Connection", "Keep-Alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Content-Type", "audio/wave")
		for true {
			binary.Write(w, binary.BigEndian, &buffer)
			flusher.Flush() // Trigger "chunked" encoding and send a chunk...
			return
		}
	})
	http.HandleFunc("/bufsize", func(writer http.ResponseWriter, request *http.Request) {
		binary.Write(writer, binary.BigEndian, seconds)
	})
	http.HandleFunc("/chatin", func(writer http.ResponseWriter, request *http.Request) {
		userData, exists := users[request.RemoteAddr]
		if exists {
			userData.audioWriter = writer
			users[request.RemoteAddr] = userData
		} else {
			users[request.RemoteAddr] = user{request.RemoteAddr, writer, nil}
		}
		fmt.Println("recieved post")
		request.ParseForm()
		fmt.Println(request.Body)
		sendmsg(request.Form["message"][0], users[request.RemoteAddr])
		defer request.Body.Close()
	})
	http.HandleFunc("/chatout", func(writer http.ResponseWriter, request *http.Request) {
		userData, exists := users[request.RemoteAddr]
		if exists {
			userData.chatWriter = writer
			users[request.RemoteAddr] = userData
		} else {
			users[request.RemoteAddr] = user{request.RemoteAddr, writer, nil}
		}
	})
	fmt.Println("Server Created")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sendmsg(msg string, sourceUser user){
	msg = sourceUser.name + ": " + msg
	fmt.Println(msg)
	for _, destUser := range users {
		if destUser == sourceUser {continue}
		binary.Write(destUser.chatWriter, binary.BigEndian, msg)
	}
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
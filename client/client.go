package main

import (
	"encoding/binary"
	"fmt"
	"net/url"

	"net/http"
	"strconv"
	"strings"
	"time"
)

var dest string = "http://localhost:8080"

const sampleRate = 44100

var seconds float64

func main() {
	//portaudio.Initialize()
	//defer portaudio.Terminate()
	var err error
	//var devices []*DeviceInfo
	//devices, err = portaudio.Devices()
	//chk(err)
	//for i := 0; i < len(devices); i++ {
	//	fmt.Println(devices[i].Name)
	//}
setupDestination:
	start := time.Now()
	fmt.Println("Connecting to " + dest + "...")
	var resp *http.Response
	resp, err = http.Get(dest + "/bufsize")
	if err != nil {
		fmt.Println("Connection to " + dest + " failed")
		var newDest string
		fmt.Println("Press enter to try again or enter a different address to connect to:")
		fmt.Scanln(&newDest)
		if newDest == "" {
			goto setupDestination
		} else {
			dest = newDest
		}
		if !strings.HasPrefix(dest, "http://") {
			dest = "http://" + dest
		}
		if strings.Count(dest, ":") != 2 {
			dest += ":8080"
		}
		goto setupDestination
	}
	ping := time.Now().Sub(start)
	fmt.Print(ping)
	fmt.Println(" milliseconds of ping")
	err = binary.Read(resp.Body, binary.BigEndian, &seconds)
	fmt.Println("Buffer is " + strconv.FormatFloat(seconds, 'f', 3, 64) + " seconds long")
	fmt.Println("Estimated audio delay of " + strconv.FormatFloat(ping.Seconds()+seconds*2, 'f', 4, 64) + " seconds")
	//buffer := make([]float32, int64(sampleRate*seconds))
	//stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, len(buffer), func(out []float32) {
	//	resp, err := http.Get(dest + "/audio")
	//	chk(err)
	//	body, _ := ioutil.ReadAll(resp.Body)
	//	responseReader := bytes.NewReader(body)
	//	binary.Read(responseReader, binary.BigEndian, &buffer)
	//	for i := range out {
	//		out[i] = buffer[i]
	//	}
	//})
	chk(err)
	//chk(stream.Start()) //initialize audio stream
	go readCharMessages() //initialize text "stream" on new thread (not sure if thread or just async but ¯\_(ツ)_/¯)
	var consoleEntrance string
askstop:
	fmt.Println("Enter chat messages in the console, enter \"/Stop\" to close")
	fmt.Scanln(&consoleEntrance)
	if strings.ToLower(consoleEntrance) != "/stop" {
		fmt.Println("Sending Message")
		data := url.Values{"message" : {consoleEntrance}}
		//postBody, err := json.Marshal(map[string]string{"message" : consoleEntrance})
		//chk(err)
		fmt.Println(data)
		http.PostForm(dest+"/chatin", data)
		goto askstop
	}
	//chk(stream.Stop())
	//defer stream.Close()
}

func readCharMessages(){
	resp, err := http.Get(dest + "/chatout")
	chk(err)
	var message string
	binary.Read(resp.Body, binary.BigEndian, &message)
	fmt.Println(message)
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

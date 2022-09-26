package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Verify nil error values
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Imports data contained in config.txt via ioutil's ReadFile() funciton
// Input: properly formatted config.txt
// Output: A map containing ip/port data in the form {"ID": "IP_ADDR:PORT", "ID2": "IP_ADDR:PORT", etc}
func read_configuration(file string) (string, string, map[string]string, []string) {
	dat, err := ioutil.ReadFile(file)
	check(err)
	lines := strings.Split(string(dat), "\n")
	minDelay := strings.Split(lines[0], " ")[0]
	maxDelay := strings.Split(lines[0], " ")[1]
	hosts := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		values := strings.Split(lines[i], " ")
		hosts[values[0]] = values[1] + ":" + values[2]
	}
	ids := make([]string, len(hosts))
	i := 0
	for k := range hosts {
		ids[i] = k
		i++
	}
	return minDelay, maxDelay, hosts, ids
}

// Sleep for a random amount of time between the interval [minDelay, maxDelay] as specified in config.txt
// This function is called in a separate goroutine. The main process blocks sending until the sleep duration has expired.
func delay(delayed chan string) {
	delay := rand.Intn(maxDelay-minDelay) + minDelay
	time.Sleep(time.Duration(delay) * time.Millisecond)
	delayed <- "resume"
}

// Send message <message> to the destination given by "IP_ADDR:PORT".
// A TCP client is created which connects to an already listening destination
// The message is passed over the socket via fprintf
func unicast_send(destination string, message string, id string) {
	delayed := make(chan string)
	c, err := net.Dial("tcp", destination)
	check(err)
	defer c.Close()
	fmt.Printf("Sent %s to process %s, system time is %s\n", message, id, time.Now().Format(time.UnixDate))
	go delay(delayed)
	<-delayed
	fmt.Fprintf(c, message)
}

// Print the message received from the source device given by conn.LocalAddr() - func server()
// The time of receipt is printed immediately and formatted in Unix style
func unicast_receive(source string, message string) {
	fmt.Printf("Received %s from process %s, system time is %s\n", message, source, time.Now().Format(time.UnixDate))
}

// Creates the TCP server that listens on the port specified in config.txt
// Forever accepts incoming connections and transfers data to a 1024-byte buffer before passing to unicast_receive
func server(ip string, live chan string, hosts map[string]string) {
	l, err := net.Listen("tcp", ip)
	check(err)
	defer l.Close()
	fmt.Println("Listening on ", ip)
	live <- "live"
	for {
		conn, err := l.Accept()
		check(err)
		buf := make([]byte, 1024)
		len, err := conn.Read(buf)
		check(err)
		// This section of commented code implements functionality for the "from proccess <ID>" statement in unicast_receive. 
		// See README.md for further detail on the chosen alternative.
		/*
			value := strings.Split(conn.LocalAddr().String(), ":")[0]
			fmt.Println("Connection was from: ", value)
			var clientId string
			for key := range hosts {
				ip := strings.Split(hosts[key], ":")[0]
				if ip == value {
					clientId = key
				}
			} */
		go unicast_receive(conn.RemoteAddr().String(), string(buf[:len]))
	}
}

// Global vars for min and max delay as specified in config.txt to avoid passing through multiple functions
var minDelay int
var maxDelay int

// Main driver function.
// Reads input from the user and passes the correct IP_ADDR/PORT combinaition to the server() and unicast_send() functions
func main() {
	live := make(chan string)
	id := os.Args[1]
	min, max, hosts, ids := read_configuration("config.txt")
	minDelay, _ = strconv.Atoi(min)
	maxDelay, _ = strconv.Atoi(max)
	fmt.Println(minDelay, maxDelay, hosts, ids)
	myIp := hosts[id]
	go server(myIp, live, hosts)
	<-live
	fmt.Printf("Process %s live at %s\n", id, myIp)
	fmt.Println("Press enter once all hosts are live.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	fmt.Println("[USAGE]: send <process ID> <message>")
	for {
		in := bufio.NewScanner(os.Stdin)
		in.Scan()
		cmd := in.Text()
		cmdSections := strings.Split(cmd, " ")
		go unicast_send(hosts[cmdSections[1]], cmdSections[2], cmdSections[1])
	}
}

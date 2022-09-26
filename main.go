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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

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

func delay(delayed chan string) {
	delay := rand.Intn(maxDelay-minDelay) + minDelay
	time.Sleep(time.Duration(delay) * time.Millisecond)
	delayed <- "resume"
}

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

func unicast_receive(source string, message string) {
	fmt.Printf("Received %s from process %s, system time is %s\n", message, source, time.Now().Format(time.UnixDate))
}

func server(ip string, live chan string) {
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
		unicast_receive(conn.RemoteAddr().String(), string(buf[:len]))
	}
}

var minDelay int
var maxDelay int

func main() {
	live := make(chan string)
	id := os.Args[1]
	min, max, hosts, ids := read_configuration("config.txt")
	minDelay, _ = strconv.Atoi(min)
	maxDelay, _ = strconv.Atoi(max)
	fmt.Println(minDelay, maxDelay, hosts, ids)
	myIp := hosts[id]
	go server(myIp, live)
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

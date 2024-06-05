package main

import (
	"flag"
	"fmt"
	"github.com/rij12/RPJCache/cache"
	"log"
	"net"
	"time"
)

func main() {
	var (
		listenAddr = flag.String("listenAddr", ":8080", "listen address of the server")
		leaderAddr = flag.String("leaderAddr", "", "listen address of the leader")
	)
	flag.Parse()

	opts := ServerOpts{
		listenAddr: *listenAddr,
		IsLeader:   true,
		LeaderAddr: *leaderAddr,
	}

	go func() {
		time.Sleep(time.Second * 2)
		conn, err := net.Dial("tcp", ":3000")
		if err != nil {
			log.Fatal(err)
		}
		_, _ = conn.Write([]byte("SET Foo Bar 2500"))

		time.Sleep(time.Second * 2)
		_, _ = conn.Write([]byte("GET Foo"))

		buf := make([]byte, 1000)
		n, _ := conn.Read(buf)
		fmt.Println(string(buf[:n]))

	}()

	server := NewService(opts, cache.New())
	server.start()
}

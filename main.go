package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/rij12/RPJCache/cache"
	"github.com/rij12/RPJCache/client"
	"log"
	"time"
)

func main() {
	listenAddr := flag.String("listenaddr", ":3000", "listen address of the server")
	leaderAddr := flag.String("leaderaddr", "", "listen address of the leader")

	flag.Parse()

	opts := ServerOpts{
		listenAddr: *listenAddr,
		IsLeader:   len(*leaderAddr) == 0,
		LeaderAddr: *leaderAddr,
	}

	go func() {
		time.Sleep(time.Second * 10)
		if opts.IsLeader {
			sendData()
		}
	}()

	server := NewService(opts, cache.New())
	err := server.start()
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func sendData() {
	for i := 0; i < 100; i++ {
		go func(i int) {
			time.Sleep(time.Second * 2)
			c, err := client.New(":3000")
			if err != nil {
				log.Fatal(err)
			}
			key := []byte(fmt.Sprintf("key_%d", i))
			value := []byte(fmt.Sprintf("val_%d", i))
			err = c.Set(context.Background(), key, value, 0)
			if err != nil {
				log.Fatal(err)
			}
			fetchedValue, err := c.Get(context.Background(), key)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(fetchedValue))
			c.Close()
		}(i)
	}
}

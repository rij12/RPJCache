package main

import (
	"context"
	"fmt"
	"github.com/rij12/RPJCache/cache"
	"log"
	"net"
)

type ServerOpts struct {
	listenAddr string
	IsLeader   bool
	LeaderAddr string
}

type Server struct {
	ServerOpts ServerOpts
	cache      cache.Cacher

	followers map[net.Conn]struct{}
}

func NewService(opts ServerOpts, c cache.Cacher) *Server {
	return &Server{
		ServerOpts: opts,
		cache:      c,
		// TODO: only allocate this when we are the leader
		followers: make(map[net.Conn]struct{}),
	}
}
func (s *Server) start() error {
	ln, err := net.Listen("tcp", s.ServerOpts.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	log.Printf("serving on [%s]", s.ServerOpts.listenAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %s \n", err)
			// Continue loop to allow others to connect
			continue
		}
		go s.handleConn(conn)
	}

}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	buf := make([]byte, 2048)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			log.Printf("read error: %s\n", err)
			break
		}
		go s.handleCommand(conn, buf)
	}
}

func (s *Server) handleCommand(conn net.Conn, rawCmd []byte) {
	msg, err := parseMessage(rawCmd)
	if err != nil {
		print("failed to parseMessage: %s\n", err)
		log.Println("failed to parse command", msg)
		_, _ = conn.Write([]byte(err.Error()))
		return
	}

	fmt.Println("Received command: ", msg.Cmd)

	switch msg.Cmd {
	case CMDSet:
		err = s.handleSetCommand(conn, msg)
	case CMDGet:
		err = s.handleGetCommand(conn, msg)
	}

	if err != nil {
		log.Println("failed to parse command", msg, err)
		_, _ = conn.Write([]byte(err.Error()))
		return
	}
}

func (s *Server) handleGetCommand(conn net.Conn, msg *Message) error {
	val, err := s.cache.Get(msg.Key)
	if err != nil {
		return err
	}
	_, err = conn.Write(val)
	return err
}

func (s *Server) handleSetCommand(conn net.Conn, msg *Message) error {
	if err := s.cache.Set(msg.Key, msg.Value, msg.TTL); err != nil {
		return err
	}

	go s.sendToFollowers(context.TODO())

	return nil
}

func (s *Server) sendToFollowers(ctx context.Context, msg *Message) error {
	// TODO: Fix Lock protection
	for conn := range s.followers {
		_, err := conn.Write(msg.ToBytes())
		if err != nil {
			log.Printf("write error: %s\n", err)
			continue
		}
	}

	return nil
}

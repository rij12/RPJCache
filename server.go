package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/rij12/RPJCache/cache"
	"github.com/rij12/RPJCache/client"
	"github.com/rij12/RPJCache/proto"
	"io"
	"log"
	"net"
	"time"
)

type ServerOpts struct {
	listenAddr string
	IsLeader   bool
	LeaderAddr string
}

type Server struct {
	ServerOpts ServerOpts
	cache      cache.Cacher

	members map[*client.Client]struct{}
}

func NewService(opts ServerOpts, c cache.Cacher) *Server {
	return &Server{
		ServerOpts: opts,
		cache:      c,
		// TODO: only allocate this when we are the leader
		members: make(map[*client.Client]struct{}),
	}
}
func (s *Server) start() error {
	ln, err := net.Listen("tcp", s.ServerOpts.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %s", err)
	}
	log.Printf("serving on [%s]", s.ServerOpts.listenAddr)

	if !s.ServerOpts.IsLeader {
		go func() {
			err := s.dialLeader()
			if err != nil {
				log.Println(err)
			}
		}()
	}

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

func (s *Server) dialLeader() error {
	conn, err := net.Dial("tcp", s.ServerOpts.LeaderAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to leader: %s", err)
	}
	log.Println("connected with leader")

	err = binary.Write(conn, binary.LittleEndian, proto.CmdJoin)
	if err != nil {
		log.Printf("failed to send command to leader: %s", err)
		return err
	}

	s.handleConn(conn)
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	for {
		cmd, err := proto.ParseCommand(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Printf("failed to parse command: %s", err)
			break
		}
		go s.handleCommand(conn, cmd)
	}
}

func (s *Server) handleCommand(conn net.Conn, cmd any) {
	switch v := cmd.(type) {
	case *proto.CommandGet:
		s.handleGetCommand(conn, v)
	case *proto.CommandSet:
		s.handleSetCommand(conn, v)
	case *proto.CommandJoin:
		s.handleJoinCommand(conn, v)
	}
}

func (s *Server) handleJoinCommand(conn net.Conn, cmd *proto.CommandJoin) error {
	log.Printf("member joined server: [%s]", conn.RemoteAddr())
	s.members[client.NewClientFromConn(conn)] = struct{}{}
	return nil
}

func (s *Server) handleSetCommand(conn net.Conn, cmd *proto.CommandSet) error {
	log.Printf("Set %s to %s", cmd.Key, cmd.Value)
	// Not Thread Safe
	go func() {
		for m := range s.members {
			err := m.Set(context.TODO(), cmd.Key, cmd.Value, cmd.TTL)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	resp := proto.ResponseSet{}
	if err := s.cache.Set(cmd.Key, cmd.Value, time.Duration(cmd.TTL)); err != nil {
		resp.Status = proto.StatusError
		_, err = conn.Write(resp.Bytes())
		return err
	}
	resp.Status = proto.StatusOk
	_, err := conn.Write(resp.Bytes())
	return err
}

func (s *Server) handleGetCommand(conn net.Conn, cmd *proto.CommandGet) error {
	resp := proto.ResponseGet{}
	value, err := s.cache.Get(cmd.Key)
	if err != nil {
		resp.Status = proto.StatusError
		_, err = conn.Write(resp.Bytes())
		return err
	}
	resp.Status = proto.StatusOk
	resp.Value = value
	_, err = conn.Write(resp.Bytes())
	return err
}

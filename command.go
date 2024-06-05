package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type command string

const (
	CMDSet command = "SET"
	CMDGet command = "GET"
)

type Message struct {
	Cmd   command
	Key   []byte
	Value []byte
	TTL   time.Duration
}

func (m *Message) ToBytes() []byte {
	switch m.Cmd {
	case CMDSet:
		cmd := fmt.Sprintf("%s %s %s %s", m.Cmd, m.Key, m.Value, m.TTL)
		return []byte(cmd)
	case CMDGet:
		cmd := fmt.Sprintf("%s %s %s", m.Cmd, m.Key, m.Value)
		return []byte(cmd)
	default:
		panic("Unknown command")
	}
}

type MSGSet struct {
	Key   []byte
	Value []byte
	TTL   time.Duration
}

type MSGGet struct {
	Key []byte
}

func parseMessage(raw []byte) (*Message, error) {
	var (
		rawStr = string(bytes.Trim(raw, "\x00"))
		parts  = strings.Split(rawStr, " ")
	)

	if len(parts) == 0 {
		// Response with conn
		log.Println("invalid command")
		return nil, errors.New("invalid protocol format")
	}

	msg := &Message{
		Cmd: command(parts[0]),
		Key: []byte(parts[1]),
	}

	if msg.Cmd == CMDSet {
		if len(parts) < 4 {
			return nil, errors.New("invalid Set command")
		}
		msg.Value = []byte(parts[2])
		ttl, err := strconv.Atoi(parts[3])
		if err != nil {
			log.Println("invalid TTL command")
			return nil, errors.New("invalid Set command")
		}
		msg.TTL = time.Duration(ttl) * time.Second
	}
	return msg, nil
}

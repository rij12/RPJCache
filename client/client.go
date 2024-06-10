package client

import (
	"context"
	"fmt"
	proto "github.com/rij12/RPJCache/proto"
	"log"
	"net"
)

type Options struct {
}

type Client struct {
	conn net.Conn
}

func NewClientFromConn(conn net.Conn) *Client {
	return &Client{conn: conn}
}

func New(endpoint string) (*Client, error) {
	conn, err := net.Dial("tcp", endpoint)
	//log.Println("Connecting to ", endpoint)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Get(ctx context.Context, key []byte) ([]byte, error) {
	cmd := &proto.CommandGet{
		Key: key,
	}

	_, err := c.conn.Write(cmd.Bytes())

	if err != nil {
		return nil, err
	}

	resp, err := proto.ParseGetResponse(c.conn)
	if err != nil {
		return nil, err
	}
	if resp.Status == proto.StatusKeyNotFound {
		return nil, fmt.Errorf("could not find key: [%s]", key)
	}
	if resp.Status != proto.StatusOk {
		return nil, fmt.Errorf("server returned status [%s]", resp.Status)
	}

	return resp.Value, nil
}

func (c *Client) Set(ctx context.Context, key, value []byte, ttl int) error {
	cmd := &proto.CommandSet{
		Key:   []byte(key),
		Value: []byte(value),
		TTL:   ttl,
	}

	_, err := c.conn.Write(cmd.Bytes())

	if err != nil {
		log.Fatalf(err.Error())
		return err
	}

	resp, err := proto.ParseSetResponse(c.conn)
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}
	if resp.Status != proto.StatusOk {
		return fmt.Errorf("server returned status [%s]", resp.Status)
	}

	return nil
}

package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type Command byte
type Status byte

func (s Status) String() string {
	switch s {
	case StatusOk:
		return "Ok"
	case StatusError:
		return "Error"
	case StatusKeyNotFound:
		return "KeyNotFound"
	default:
		return "unknown status"
	}
}

const (
	StatusNone = iota
	StatusOk
	StatusError
	StatusKeyNotFound
)

const (
	CmdNonce Command = iota
	CmdSet
	CmdGet
	CmdDel
	CmdJoin
)

type ResponseGet struct {
	Status Status
	Value  []byte
}

func (r *ResponseGet) Bytes() []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.LittleEndian, r.Status)

	valueLen := int32(len(r.Value))
	binary.Write(buff, binary.LittleEndian, valueLen)
	binary.Write(buff, binary.LittleEndian, r.Value)
	return buff.Bytes()
}

func ParseSetResponse(r io.Reader) (*ResponseSet, error) {
	resp := new(ResponseSet)
	err := binary.Read(r, binary.LittleEndian, resp)
	return resp, err
}

type ResponseSet struct {
	Status Status
}

func ParseGetResponse(r io.Reader) (*ResponseGet, error) {
	resp := &ResponseGet{}
	binary.Read(r, binary.LittleEndian, &resp.Status)

	var valueLen int32
	binary.Read(r, binary.LittleEndian, &valueLen)

	resp.Value = make([]byte, valueLen)
	binary.Read(r, binary.LittleEndian, &resp.Value)

	return resp, nil
}

func (r *ResponseSet) Bytes() []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.LittleEndian, r.Status)
	return buff.Bytes()
}

type CommandJoin struct{}

type CommandSet struct {
	Key   []byte
	Value []byte
	TTL   int
}

func (c *CommandSet) Bytes() []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, CmdSet)

	var keyLen = int32(len(c.Key))
	_ = binary.Write(buf, binary.LittleEndian, keyLen)
	_ = binary.Write(buf, binary.LittleEndian, c.Key)

	var valueLen = int32(len(c.Value))
	_ = binary.Write(buf, binary.LittleEndian, valueLen)
	_ = binary.Write(buf, binary.LittleEndian, c.Value)

	_ = binary.Write(buf, binary.LittleEndian, int32(c.TTL))

	return buf.Bytes()
}

type CommandGet struct {
	Key []byte
}

func (c *CommandGet) Bytes() []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, CmdGet)

	var keyLen = int32(len(c.Key))
	_ = binary.Write(buf, binary.LittleEndian, keyLen)
	_ = binary.Write(buf, binary.LittleEndian, c.Key)
	return buf.Bytes()
}

func ParseCommand(r io.Reader) (any, error) {
	var cmd Command
	err := binary.Read(r, binary.LittleEndian, &cmd)

	if err != nil {
		return nil, err
	}
	switch cmd {
	case CmdSet:
		return parseSetCommand(r), nil
	case CmdGet:
		return parseGetCommand(r), nil
	case CmdJoin:
		return &CommandJoin{}, nil
	default:
		return nil, errors.New("unknown command")
	}
}

func parseSetCommand(r io.Reader) *CommandSet {

	cmd := &CommandSet{}

	var keyLen int32
	binary.Read(r, binary.LittleEndian, &keyLen)
	cmd.Key = make([]byte, keyLen)
	binary.Read(r, binary.LittleEndian, &cmd.Key)

	var valueLen int32
	binary.Read(r, binary.LittleEndian, &valueLen)
	cmd.Value = make([]byte, keyLen)
	binary.Read(r, binary.LittleEndian, &cmd.Value)

	var ttl int32
	binary.Read(r, binary.LittleEndian, &ttl)
	cmd.TTL = int(ttl)

	return cmd
}

func parseGetCommand(r io.Reader) *CommandGet {

	cmd := &CommandGet{}

	var keyLen int32
	binary.Read(r, binary.LittleEndian, &keyLen)
	cmd.Key = make([]byte, keyLen)
	binary.Read(r, binary.LittleEndian, &cmd.Key)

	return cmd
}

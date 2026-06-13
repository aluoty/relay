package protocol

import (
	"bufio"
	"encoding/json"
	"io"
)

const (
	TypeJoin  = "join"
	TypeLeave = "leave"
	TypeChat  = "msg"
	TypeSys   = "sys"
	TypeUsers = "users"
)

type Message struct {
	Type  string   `json:"t"`
	From  string   `json:"f,omitempty"`
	Text  string   `json:"x,omitempty"`
	Users []string `json:"u,omitempty"`
}

func Write(w io.Writer, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

func Read(r *bufio.Reader) (Message, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return Message{}, err
	}
	var msg Message
	err = json.Unmarshal(line, &msg)
	return msg, err
}

func MarshalLine(msg Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

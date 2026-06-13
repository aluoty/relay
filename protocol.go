package main

import (
	"bufio"
	"encoding/json"
	"io"
)

const (
	msgJoin  = "join"
	msgLeave = "leave"
	msgChat  = "msg"
	msgSys   = "sys"
)

type Message struct {
	Type string `json:"t"`
	From string `json:"f,omitempty"`
	Text string `json:"x,omitempty"`
}

func writeMessage(w io.Writer, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

func readMessage(r *bufio.Reader) (Message, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return Message{}, err
	}
	var msg Message
	err = json.Unmarshal(line, &msg)
	return msg, err
}

package protocol

import (
	"bufio"
	"encoding/json"
	"io"
)

const (
	TypeJoin   = "join"
	TypeLeave  = "leave"
	TypeChat   = "msg"
	TypeSys    = "sys"
	TypeUsers  = "users"
	TypeGroups = "groups"
	TypeSwitch = "switch"
	TypeCreate = "create"
	TypeAvatar = "avatar"
)

const DefaultGroup = "general"

type Message struct {
	Type     string            `json:"t"`
	From     string            `json:"f,omitempty"`
	Text     string            `json:"x,omitempty"`
	Group    string            `json:"g,omitempty"`
	Users    []string          `json:"u,omitempty"`
	Groups   []string          `json:"gs,omitempty"`
	Avatar   string            `json:"a,omitempty"`
	Profiles map[string]string `json:"p,omitempty"`
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

func Join(name, group, avatar string) Message {
	if group == "" {
		group = DefaultGroup
	}
	return Message{Type: TypeJoin, From: name, Group: group, Avatar: avatar}
}

func Chat(text, group string) Message {
	return Message{Type: TypeChat, Text: text, Group: group}
}

func Switch(group string) Message {
	return Message{Type: TypeSwitch, Group: group}
}

func CreateGroup(name string) Message {
	return Message{Type: TypeCreate, Text: name}
}

func SetAvatar(avatar string) Message {
	return Message{Type: TypeAvatar, Avatar: avatar}
}

func Sys(text string) Message {
	return Message{Type: TypeSys, Text: text}
}

func SysGroup(group, text string) Message {
	return Message{Type: TypeSys, Group: group, Text: text}
}

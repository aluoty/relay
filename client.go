package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func runClient(addr, name string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	if err := writeMessage(conn, Message{Type: msgJoin, From: name}); err != nil {
		return fmt.Errorf("join: %w", err)
	}

	app := tview.NewApplication()

	messages := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	messages.SetBorder(true).SetTitle(" Relay Chat ")

	status := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf(" [gray]connected to %s as [green]%s", addr, name))
	status.SetBorder(false)

	input := tview.NewInputField().
		SetLabel("Message: ").
		SetFieldWidth(0)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(messages, 0, 1, false).
		AddItem(status, 1, 0, false).
		AddItem(input, 3, 0, true)

	appendMessage := func(msg Message) {
		switch msg.Type {
		case msgChat:
			if msg.From == name {
				fmt.Fprintf(messages, "[green]%s:[white] %s\n", msg.From, msg.Text)
			} else {
				fmt.Fprintf(messages, "[yellow]%s:[white] %s\n", msg.From, msg.Text)
			}
		case msgJoin:
			fmt.Fprintf(messages, "[gray]* %s joined\n", msg.From)
		case msgLeave:
			fmt.Fprintf(messages, "[gray]* %s left\n", msg.From)
		case msgSys:
			fmt.Fprintf(messages, "[gray]* %s\n", msg.Text)
		}
		messages.ScrollToEnd()
	}

	go func() {
		reader := bufio.NewReader(conn)
		for {
			msg, err := readMessage(reader)
			if err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(messages, "[red]* disconnected\n")
					status.SetText(" [red]disconnected")
					messages.ScrollToEnd()
				})
				return
			}
			msgCopy := msg
			app.QueueUpdateDraw(func() {
				appendMessage(msgCopy)
			})
		}
	}()

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		text := strings.TrimSpace(input.GetText())
		if text == "" {
			return
		}
		if err := writeMessage(conn, Message{Type: msgChat, Text: text}); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(messages, "[red]* send failed: %v\n", err)
				messages.ScrollToEnd()
			})
			return
		}
		input.SetText("")
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		return event
	})

	return app.SetRoot(flex, true).Run()
}

func defaultName() string {
	if u, err := os.UserHomeDir(); err == nil {
		if base := strings.TrimPrefix(u, "/home/"); base != u && base != "" {
			return base
		}
	}
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "guest"
}

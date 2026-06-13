package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/aluoty/relay/internal/protocol"
	"github.com/aluoty/relay/internal/tlsconfig"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Config struct {
	Addr string
	Name string
	TLS  tlsconfig.ClientConfig
}

func Run(cfg Config) error {
	conn, err := cfg.TLS.Dial(cfg.Addr)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	if err := protocol.Write(conn, protocol.Message{Type: protocol.TypeJoin, From: cfg.Name}); err != nil {
		return fmt.Errorf("join: %w", err)
	}

	ui := newUI(cfg.Addr, cfg.Name)
	return ui.run(conn)
}

type ui struct {
	app      *tview.Application
	messages *tview.TextView
	users    *tview.TextView
	status   *tview.TextView
	input    *tview.InputField
	addr     string
	self     string
	online   map[string]struct{}
}

func newUI(addr, name string) *ui {
	return &ui{
		addr:   addr,
		self:   name,
		online: make(map[string]struct{}),
	}
}

func (u *ui) run(conn net.Conn) error {
	u.app = tview.NewApplication()

	u.messages = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	u.messages.SetBorder(true).SetTitle(" Relay Chat ")

	u.users = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	u.users.SetBorder(true).SetTitle(" Users ")

	u.status = tview.NewTextView().SetDynamicColors(true)
	u.status.SetText(fmt.Sprintf(" [gray]connected to %s as [green]%s", u.addr, u.self))

	u.input = tview.NewInputField().
		SetLabel("Message: ").
		SetFieldWidth(0)

	chatPane := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(u.messages, 0, 1, false).
		AddItem(u.status, 1, 0, false).
		AddItem(u.input, 3, 0, true)

	root := tview.NewFlex().
		AddItem(chatPane, 0, 1, true).
		AddItem(u.users, 22, 0, false)

	go u.readLoop(conn)

	u.input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		text := strings.TrimSpace(u.input.GetText())
		if text == "" {
			return
		}
		if err := protocol.Write(conn, protocol.Message{Type: protocol.TypeChat, Text: text}); err != nil {
			u.app.QueueUpdateDraw(func() {
				u.printSystem(fmt.Sprintf("[red]* send failed: %v", err))
			})
			return
		}
		u.input.SetText("")
	})

	u.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			u.app.Stop()
			return nil
		}
		return event
	})

	return u.app.SetRoot(root, true).SetFocus(u.input).Run()
}

func (u *ui) readLoop(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := protocol.Read(reader)
		if err != nil {
			u.app.QueueUpdateDraw(func() {
				u.printSystem("[red]* disconnected")
				u.status.SetText(" [red]disconnected")
			})
			return
		}
		msgCopy := msg
		u.app.QueueUpdateDraw(func() {
			u.handleMessage(msgCopy)
		})
	}
}

func (u *ui) handleMessage(msg protocol.Message) {
	switch msg.Type {
	case protocol.TypeChat:
		if msg.From == u.self {
			fmt.Fprintf(u.messages, "[green]%s:[white] %s\n", msg.From, msg.Text)
		} else {
			fmt.Fprintf(u.messages, "[yellow]%s:[white] %s\n", msg.From, msg.Text)
		}
	case protocol.TypeJoin:
		u.printSystem(fmt.Sprintf("[gray]* %s joined", msg.From))
	case protocol.TypeLeave:
		u.printSystem(fmt.Sprintf("[gray]* %s left", msg.From))
	case protocol.TypeSys:
		u.printSystem(fmt.Sprintf("[gray]* %s", msg.Text))
	case protocol.TypeUsers:
		u.setUsers(msg.Users)
	}
	u.messages.ScrollToEnd()
}

func (u *ui) printSystem(text string) {
	fmt.Fprintf(u.messages, "%s\n", text)
}

func (u *ui) setUsers(names []string) {
	u.online = make(map[string]struct{}, len(names))
	for _, name := range names {
		u.online[name] = struct{}{}
	}

	sorted := append([]string(nil), names...)
	sort.Strings(sorted)

	var b strings.Builder
	for _, name := range sorted {
		if name == u.self {
			fmt.Fprintf(&b, "[green]%s[white]\n", name)
		} else {
			fmt.Fprintf(&b, "%s\n", name)
		}
	}
	u.users.SetText(b.String())
}

func DefaultName() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("LOGNAME"); user != "" {
		return user
	}
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "guest"
}

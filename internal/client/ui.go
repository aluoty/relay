package client

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/aluoty/relay/internal/ascii"
	"github.com/aluoty/relay/internal/commands"
	"github.com/aluoty/relay/internal/protocol"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ui struct {
	app      *tview.Application
	state    *state
	messages *tview.TextView
	groups   *tview.List
	users    *tview.TextView
	status   *tview.TextView
	input    *tview.InputField
}

func newUI(addr, name, group, avatar string) *ui {
	return &ui{state: newState(addr, name, group, avatar)}
}

func (u *ui) run(conn net.Conn) error {
	u.build(conn)
	go u.readLoop(conn)
	return u.app.Run()
}

func (u *ui) build(conn net.Conn) {
	u.app = tview.NewApplication()

	u.messages = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	u.messages.SetBorder(true).SetTitle(" Relay Chat ")

	u.groups = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedFunc(func(index int, _, _ string, _ rune) {
			if index < 0 || index >= len(u.state.groups) {
				return
			}
			u.switchGroup(conn, u.state.groups[index])
		})
	u.groups.SetBorder(true).SetTitle(" Groups ")

	u.users = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	u.users.SetBorder(true).SetTitle(" Users ")

	u.status = tview.NewTextView().SetDynamicColors(true)
	u.status.SetText(u.state.statusText())

	u.input = tview.NewInputField().
		SetLabel("Message: ").
		SetFieldWidth(0)

	u.renderGroups()
	u.renderUsers()

	chatPane := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(u.messages, 0, 1, false).
		AddItem(u.status, 1, 0, false).
		AddItem(u.input, 3, 0, true)

	root := tview.NewFlex().
		AddItem(u.groups, 18, 0, false).
		AddItem(chatPane, 0, 1, true).
		AddItem(u.users, 22, 0, false)

	u.input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		text := strings.TrimSpace(u.input.GetText())
		u.input.SetText("")
		if text == "" {
			return
		}
		u.handleInput(conn, text)
	})

	u.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			u.app.Stop()
			return nil
		case tcell.KeyTab:
			u.cycleFocus()
			return nil
		}
		return event
	})

	u.app.SetRoot(root, true).SetFocus(u.input)
}

func (u *ui) cycleFocus() {
	switch u.app.GetFocus().(type) {
	case *tview.InputField:
		u.app.SetFocus(u.groups)
	case *tview.List:
		u.app.SetFocus(u.users)
	default:
		u.app.SetFocus(u.input)
	}
}

func (u *ui) readLoop(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := protocol.Read(reader)
		if err != nil {
			u.queue(func() {
				u.printSystem("[red]* disconnected")
				u.status.SetText(" [red]disconnected")
			})
			return
		}
		msgCopy := msg
		u.queue(func() {
			u.handleMessage(msgCopy)
		})
	}
}

func (u *ui) queue(fn func()) {
	u.app.QueueUpdateDraw(fn)
}

func (u *ui) handleMessage(msg protocol.Message) {
	switch msg.Type {
	case protocol.TypeChat:
		group := msg.Group
		if group == "" {
			group = protocol.DefaultGroup
		}
		if group != u.state.group {
			return
		}
		u.printChat(msg.From, msg.Avatar, msg.Text)
	case protocol.TypeJoin:
		if msg.Group == u.state.group {
			u.printSystem(fmt.Sprintf("[gray]* %s joined #%s", ascii.FormatSpeaker(msg.Avatar, msg.From), msg.Group))
		}
	case protocol.TypeLeave:
		if msg.Group == u.state.group {
			u.printSystem(fmt.Sprintf("[gray]* %s left #%s", msg.From, msg.Group))
		}
	case protocol.TypeSys:
		u.printSystem(fmt.Sprintf("[gray]* %s", msg.Text))
	case protocol.TypeUsers:
		if msg.Group != u.state.group {
			return
		}
		u.state.setUsers(msg.Users, msg.Profiles)
		u.renderUsers()
	case protocol.TypeGroups:
		u.state.setGroups(msg.Groups)
		u.renderGroups()
	case protocol.TypeSwitch:
		u.onGroupChanged(msg.Group)
	case protocol.TypeAvatar:
		if msg.Group != u.state.group {
			return
		}
		u.state.setUserAvatar(msg.From, msg.Avatar)
		u.renderUsers()
		if msg.From != u.self() {
			u.printSystem(fmt.Sprintf("[gray]* %s updated avatar to %s", msg.From, ascii.Prefix(msg.Avatar)))
		}
	}
	u.messages.ScrollToEnd()
}

func (u *ui) handleInput(conn net.Conn, text string) {
	cmd, payload := commands.Parse(text)
	switch cmd.Kind {
	case commands.KindNone:
		u.sendChat(conn, payload)
	case commands.KindHelp:
		u.printSystem("[gray]" + strings.ReplaceAll(commands.HelpText(), "\n", "\n[gray]"))
	case commands.KindGroup:
		u.switchGroup(conn, cmd.Arg)
	case commands.KindGroups:
		u.printSystem(fmt.Sprintf("[gray]* groups: %s", strings.Join(u.state.groups, ", ")))
	case commands.KindCreate:
		u.send(conn, protocol.CreateGroup(cmd.Arg))
	case commands.KindAvatar:
		u.setAvatar(conn, cmd.Arg)
	}
}

func (u *ui) sendChat(conn net.Conn, text string) {
	if err := protocol.Write(conn, protocol.Chat(text, u.state.group)); err != nil {
		u.printSystem(fmt.Sprintf("[red]* send failed: %v", err))
	}
}

func (u *ui) switchGroup(conn net.Conn, group string) {
	group = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(group)), "#")
	if group == "" || group == u.state.group {
		return
	}
	u.send(conn, protocol.Switch(group))
}

func (u *ui) setAvatar(conn net.Conn, raw string) {
	avatar, err := ascii.Resolve(raw)
	if err != nil {
		u.printSystem(fmt.Sprintf("[red]* %v", err))
		return
	}
	u.state.setSelfAvatar(avatar)
	u.renderUsers()
	u.send(conn, protocol.SetAvatar(avatar))
}

func (u *ui) send(conn net.Conn, msg protocol.Message) {
	if err := protocol.Write(conn, msg); err != nil {
		u.printSystem(fmt.Sprintf("[red]* send failed: %v", err))
	}
}

func (u *ui) printChat(from, avatar, text string) {
	if avatar == "" {
		avatar = u.state.avatarFor(from)
	}
	fmt.Fprintf(u.messages, "%s\n", formatChatLine(from, avatar, text, u.self()))
}

func (u *ui) printSystem(text string) {
	fmt.Fprintf(u.messages, "%s\n", text)
}

func (u *ui) renderGroups() {
	u.groups.Clear()
	for _, name := range u.state.groups {
		u.groups.AddItem(formatGroupLabel(name, u.state.group), "", 0, nil)
	}
}

func (u *ui) renderUsers() {
	var b strings.Builder
	for _, name := range u.state.users {
		fmt.Fprintf(&b, "%s\n", formatUserLine(name, u.state.avatarFor(name), u.self()))
	}
	u.users.SetText(b.String())
}

func (u *ui) self() string {
	return u.state.self
}

func (u *ui) onGroupChanged(group string) {
	u.state.group = group
	u.messages.Clear()
	u.status.SetText(u.state.statusText())
	u.renderGroups()
	u.renderUsers()
}

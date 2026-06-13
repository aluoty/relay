package client

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/aluoty/relay/internal/avatar"
	"github.com/aluoty/relay/internal/commands"
	"github.com/aluoty/relay/internal/protocol"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ui struct {
	app      *tview.Application
	pages    *tview.Pages
	conn     net.Conn
	state    *state
	messages *tview.TextView
	groups   *tview.List
	users    *tview.TextView
	status   *tview.TextView
	input    *tview.InputField
	help     *tview.TextView
}

func newUI(addr, name, group, avatarText string) *ui {
	return &ui{state: newState(addr, name, group, avatarText)}
}

func (u *ui) run(conn net.Conn) error {
	u.conn = conn
	u.build()
	go u.readLoop(conn)
	return u.app.Run()
}

func (u *ui) build() {
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
			u.selectGroup(index)
		})
	u.groups.SetBorder(true).SetTitle(" Groups ")

	u.users = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	u.users.SetBorder(true).SetTitle(" Users ")

	u.status = tview.NewTextView().SetDynamicColors(true)

	u.input = tview.NewInputField().
		SetLabel("Message: ").
		SetFieldWidth(0)

	u.help = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	u.help.SetBorder(true).SetTitle(" Help ")
	u.help.SetText(commands.HelpFormatted())

	u.renderGroups()
	u.renderUsers()
	u.updateStatus()

	chatPane := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(u.messages, 0, 1, false).
		AddItem(u.status, 1, 0, false).
		AddItem(u.input, 3, 0, true)

	main := tview.NewFlex().
		AddItem(u.groups, 18, 0, false).
		AddItem(chatPane, 0, 1, true).
		AddItem(u.users, 22, 0, false)

	helpBox := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 2, false).
			AddItem(tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(u.help, 0, 1, true).
				AddItem(tview.NewTextView().
					SetDynamicColors(true).
					SetText("\n[gray]Esc or Enter to close[white]"), 1, 0, false),
				0, 1, true).
			AddItem(nil, 0, 2, false),
			0, 3, true).
		AddItem(nil, 0, 1, false)

	u.pages = tview.NewPages()
	u.pages.AddPage("main", main, true, true)
	u.pages.AddPage("help", helpBox, true, false)

	u.input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		text := strings.TrimSpace(u.input.GetText())
		u.input.SetText("")
		if text == "" {
			return
		}
		u.handleInput(text)
	})

	u.app.SetInputCapture(u.captureInput)
	u.app.SetRoot(u.pages, true).SetFocus(u.input)
}

func (u *ui) captureInput(event *tcell.EventKey) *tcell.EventKey {
	if u.helpVisible() {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyEnter:
			u.hideHelp()
			return nil
		}
		return event
	}

	switch event.Key() {
	case tcell.KeyCtrlC:
		u.app.Stop()
		return nil
	case tcell.KeyEscape:
		u.focusInput()
		return nil
	case tcell.KeyTab:
		u.cycleFocus()
		return nil
	case tcell.KeyCtrlG:
		if u.focusedGroups() {
			u.focusInput()
		} else {
			u.focusGroups()
		}
		return nil
	}

	if event.Key() == tcell.KeyRune && u.focusedGroups() {
		if event.Rune() >= '1' && event.Rune() <= '9' {
			u.selectGroup(int(event.Rune() - '1'))
			return nil
		}
	}
	return event
}

func (u *ui) helpVisible() bool {
	name, _ := u.pages.GetFrontPage()
	return name == "help"
}

func (u *ui) showHelp() {
	u.pages.ShowPage("help")
	u.app.SetFocus(u.help)
}

func (u *ui) hideHelp() {
	u.pages.HidePage("help")
	u.focusInput()
}

func (u *ui) focusInput() {
	u.app.SetFocus(u.input)
	u.updateStatus()
}

func (u *ui) focusGroups() {
	u.app.SetFocus(u.groups)
	u.highlightCurrentGroup()
	u.updateStatus()
}

func (u *ui) focusedGroups() bool {
	_, ok := u.app.GetFocus().(*tview.List)
	return ok
}

func (u *ui) updateStatus() {
	if u.focusedGroups() {
		u.status.SetText(formatGroupsStatus())
		return
	}
	u.status.SetText(u.state.statusText())
}

func (u *ui) selectGroup(index int) {
	if index < 0 || index >= len(u.state.groups) {
		return
	}
	u.switchGroup(u.state.groups[index])
	u.focusInput()
}

func (u *ui) highlightCurrentGroup() {
	for i, name := range u.state.groups {
		if name == u.state.group {
			u.groups.SetCurrentItem(i)
			return
		}
	}
}

func (u *ui) cycleFocus() {
	switch u.app.GetFocus().(type) {
	case *tview.InputField:
		u.focusGroups()
	case *tview.List:
		u.app.SetFocus(u.users)
		u.updateStatus()
	default:
		u.focusInput()
	}
}

func (u *ui) readLoop(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := protocol.Read(reader)
		if err != nil {
			u.queue(func() {
				u.printSystem("disconnected")
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
			u.printSystem(fmt.Sprintf("%s joined #%s", avatar.FormatSpeaker(msg.Avatar, msg.From), msg.Group))
		}
	case protocol.TypeLeave:
		if msg.Group == u.state.group {
			u.printSystem(fmt.Sprintf("%s left #%s", msg.From, msg.Group))
		}
	case protocol.TypeSys:
		u.printSystem(msg.Text)
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
			u.printSystem(fmt.Sprintf("%s updated avatar to %s", msg.From, avatar.Prefix(msg.Avatar)))
		}
	}
	u.messages.ScrollToEnd()
}

func (u *ui) handleInput(text string) {
	cmd, payload := commands.Parse(text)
	switch cmd.Kind {
	case commands.KindNone:
		u.sendChat(payload)
	case commands.KindHelp:
		u.showHelp()
	case commands.KindGroup:
		u.switchGroup(cmd.Arg)
	case commands.KindGroups:
		u.printSystem("groups: " + strings.Join(u.state.groups, ", "))
	case commands.KindCreate:
		u.send(protocol.CreateGroup(cmd.Arg))
	case commands.KindAvatar:
		u.setAvatar(cmd.Arg)
	}
}

func (u *ui) sendChat(text string) {
	if err := protocol.Write(u.conn, protocol.Chat(text, u.state.group)); err != nil {
		u.printSystem(fmt.Sprintf("send failed: %v", err))
	}
}

func (u *ui) switchGroup(group string) {
	group = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(group)), "#")
	if group == "" || group == u.state.group {
		return
	}
	u.send(protocol.Switch(group))
}

func (u *ui) setAvatar(raw string) {
	v, err := avatar.Resolve(raw)
	if err != nil {
		u.printSystem(fmt.Sprintf("avatar: %v", err))
		return
	}
	u.state.setSelfAvatar(v)
	u.renderUsers()
	u.send(protocol.SetAvatar(v))
}

func (u *ui) send(msg protocol.Message) {
	if err := protocol.Write(u.conn, msg); err != nil {
		u.printSystem(fmt.Sprintf("send failed: %v", err))
	}
}

func (u *ui) printChat(from, av, text string) {
	if av == "" {
		av = u.state.avatarFor(from)
	}
	fmt.Fprintf(u.messages, "%s\n", formatChatLine(from, av, text, u.self()))
}

func (u *ui) printSystem(text string) {
	fmt.Fprintf(u.messages, "%s\n", formatSystemLine(text))
}

func (u *ui) renderGroups() {
	u.groups.Clear()
	current := 0
	for i, name := range u.state.groups {
		u.groups.AddItem(formatGroupLabel(name, u.state.group), "", 0, nil)
		if name == u.state.group {
			current = i
		}
	}
	u.groups.SetCurrentItem(current)
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
	u.renderGroups()
	u.renderUsers()
	u.highlightCurrentGroup()
	u.updateStatus()
}

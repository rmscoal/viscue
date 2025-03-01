package message

import (
	"viscue/tui/entity"

	"github.com/charmbracelet/bubbles/help"
)

type SwitchFocusMsg int8

var (
	SidebarFocused = SwitchFocusMsg(0)
	ShelfFocused   = SwitchFocusMsg(1)
	PromptFocused  = SwitchFocusMsg(2)
)

type ShouldReloadMsg struct{}

type OpenPromptMsg[T interface {
	entity.Category | entity.Password
}] struct {
	Payload    T
	IsDeletion bool
}

type ClosePromptMsg[T interface {
	entity.Category | entity.Password
}] struct{}

type CategorySelectedMsg int64

type SetHelpKeysMsg struct {
	Keys help.KeyMap
}

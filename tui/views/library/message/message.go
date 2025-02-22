package message

import "viscue/tui/entity"

type SwitchFocusMsg int8

var (
	SidebarFocused = SwitchFocusMsg(0)
	ShelfFocused   = SwitchFocusMsg(1)
	PromptFocused  = SwitchFocusMsg(2)
)

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

package event

type AppMessage int8

const (
	UserLoggedIn AppMessage = iota
)

type LibraryMessage int8

const ClosePrompt LibraryMessage = iota

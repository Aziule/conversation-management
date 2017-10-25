package conversation

type Status string

const (
	StatusOngoing           Status = "ongoing"
	StatusHumanIntervention Status = "human"
	StatusOver              Status = "over"
)

type Conversation struct {
	Status       Status
	MessagesFlow *MessagesFlow
}

type Reader interface {
	FindLatestConversation(*User) (*Conversation, error)
	FindUser(UserId) (*User, error)
}

type Writer interface {
	InsertUser(*User) error
	Save(*Conversation) error
}

package domain

type Notification struct {
	ID      string
	UserID  string
	Channel string // email, sms, push
	Payload string
}

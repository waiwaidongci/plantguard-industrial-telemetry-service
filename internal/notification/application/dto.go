package application

type SendNotificationInput struct {
	Channel string `json:"channel"`
	Target  string `json:"target"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

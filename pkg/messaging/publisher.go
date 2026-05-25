package messaging

type PublishRequest struct {
	Topic string
	Key   string
	Msg   Envelope
}

package notifier

type TelegramSender struct {
}

func NewTelegramSender() TelegramSender {
	return TelegramSender{}
}

func (sender TelegramSender) Send(sub Subscription) error {

	return nil
}

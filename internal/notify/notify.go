package notify

type Notifier interface {
	Send(title, body string) error
}

type NullNotifier struct{}

func (NullNotifier) Send(_, _ string) error { return nil }

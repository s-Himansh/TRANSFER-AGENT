package service

type Sender interface {
	Send(string) error
}

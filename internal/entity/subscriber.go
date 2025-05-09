package entity

type MessageHandler func(msg interface{})
type Subscriber struct {
	Ch chan interface{}
	Cb MessageHandler
}

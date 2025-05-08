package eventmanager

type baseEvent struct {
	done   bool
	action func()
}

func (event *baseEvent) Action() {
	event.action()
	event.done = true
}

func (event *baseEvent) Done() bool {
	return event.done
}

package event

type JoinEvent struct {
	Name string
}

type LeaveEvent struct {
	Name string
}

type CloseEvent struct{}

func (je JoinEvent) Send()  {}
func (le LeaveEvent) Send() {}
func (ce CloseEvent) Send() {}

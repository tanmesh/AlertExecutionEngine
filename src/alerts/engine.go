package alerts

type Engine interface {
	Run(alert *Alert)
}

type engine struct {
	client Client
}

func NewEngine(client_ Client) Engine {
	return &engine{client_}
}

type alertStateInfo struct {
	PreviousState string
	Timestamp     int64
}

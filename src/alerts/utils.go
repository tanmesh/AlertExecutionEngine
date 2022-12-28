package alerts

type Engine interface {
	Run(alert *Alert) error
}

type engine struct {
	client Client
}

func NewEngine(client_ Client) *engine {
	return &engine{
		client: client_,
	}
}

type alertStateInfo struct {
	PreviousState string
	Timestamp     int64
}

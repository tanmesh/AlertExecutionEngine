package alerts

import (
	"context"
	"fmt"
	"gopkg.in/retry.v1"
	"log"
	"time"
)

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

const (
	PASS     = "PASS"
	WARN     = "WARN"
	CRITICAL = "CRITICAL"
	RESOLVE  = "RESOLVE"
	NOTIFY   = "NOTIFY"
	QUERY    = "QUERY"
)

func (_engine engine) Run(alert *Alert) {
	alertState := alertStateInfo{PreviousState: PASS, Timestamp: time.Now().Unix()}

	ticker := time.NewTicker(time.Second * time.Duration(alert.IntervalSecs))

	for {
		select {
		case <-ticker.C:
			err := _engine.process(alert, &alertState)
			if err != nil {
				fmt.Printf("Error while processing %v: %v\n", alert.Name, err)
			}
		}
	}
}

func (_engine engine) process(alert *Alert, alertState *alertStateInfo) error {
	currentState := _engine.getCurrentState(alert)

	fmt.Printf("%v ; %v ; [%v -> %v] ; %v\n", alert.Query, currentState, alertState.PreviousState, currentState, time.Now().Unix())

	if currentState == PASS {
		if alertState.PreviousState != PASS {
			err := _engine.resolve(alertState, alert, currentState)
			if err != nil {
				fmt.Printf("Error while resolving: %v\n", err)
				return err
			}
		}
	} else if currentState != alertState.PreviousState || time.Now().Unix()-alertState.Timestamp >= alert.RepeatIntervalSecs {
		err := _engine.notify(alertState, alert, currentState)
		if err != nil {
			fmt.Printf("Error while notifying: %v\n", err)
			return err
		}
	}
	return nil
}

func (_engine engine) resolve(alertState *alertStateInfo, alert *Alert, currentState string) error {
	*alertState = alertStateInfo{PreviousState: currentState, Timestamp: time.Now().Unix()}
	fmt.Printf("Resolving %v! \t Next alert at %v(from %v)\n", alert.Query, alertState.Timestamp, time.Now().Unix())
	_, err := _engine.retryHelper(RESOLVE, alert, "")
	return err
}

func (_engine engine) notify(alertState *alertStateInfo, alert *Alert, currentState string) error {
	*alertState = alertStateInfo{currentState, time.Now().Unix()}
	fmt.Printf("Sending ALERT %v! \t Next alert at %v(from %v)\n", alert.Query, alertState.Timestamp, time.Now().Unix())

	message := alert.Warn.Message
	if currentState == CRITICAL {
		message = alert.Critical.Message
	}

	_, err := _engine.retryHelper(NOTIFY, alert, message)
	return err
}

func (_engine engine) getCurrentState(alert *Alert) string {
	queryResponse, err := _engine.retryHelper(QUERY, alert, "")

	if err != nil {
		log.Printf("Error while extracting CurrentState: %v %v\n", err, alert.Name)
	}

	currentValue := queryResponse
	if currentValue <= alert.Warn.Value {
		return PASS
	} else if currentValue <= alert.Critical.Value {
		return WARN
	}
	return CRITICAL
}

func (_engine engine) retryHelper(parameter string, alert *Alert, message string) (float32, error) {
	var err error
	var queryResponse float32
	for attempt := retry.Start(getRetryStrategy(), nil); attempt.Next(); {
		switch parameter {
		case NOTIFY:
			err = _engine.client.Notify(context.Background(), alert.Name, message)
		case RESOLVE:
			err = _engine.client.Resolve(context.Background(), alert.Name)
		case QUERY:
			queryResponse, err = _engine.client.Query(context.Background(), alert.Query)
		}
		if !shouldRetry(err) {
			break
		}
		fmt.Printf("Retrying %v", alert.Name)
	}
	return queryResponse, err
}

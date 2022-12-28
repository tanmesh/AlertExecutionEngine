package alerts

import (
	"context"
	"fmt"
	"log"
	"time"
)

func (_engine engine) Run(alert *Alert) {
	alertState := alertStateInfo{PreviousState: "PASS", Timestamp: time.Now().Unix()}

	ticker := time.NewTicker(time.Second * time.Duration(alert.IntervalSecs))

	for {
		select {
		case <-ticker.C:
			err := _engine.process(alert, &alertState)
			if err != nil {
				fmt.Errorf("Error processing %v %v\n", alert.Name, err)
			}
		}
	}
}

func (_engine engine) process(alert *Alert, alertState *alertStateInfo) error {
	currentState := _engine.getCurrentState(alert)

	fmt.Printf("%v ; %v ; [%v -> %v] ; %v\n", alert.Query, currentState, alertState.PreviousState, currentState, time.Now().Unix())

	if currentState == "PASS" {
		if alertState.PreviousState != "PASS" {
			err := _engine.resolve(alertState, alert, currentState)
			if err != nil {
				fmt.Printf("Line 99: ")
				return err
			}
		}
	} else if currentState != alertState.PreviousState || time.Now().Unix()-alertState.Timestamp >= alert.RepeatIntervalSecs {
		err := _engine.notify(alertState, alert, currentState)
		if err != nil {
			fmt.Println("Line 118: ")
			return err
		}
	}
	return nil
}

func (_engine engine) resolve(alertState *alertStateInfo, alert *Alert, currentState string) error {
	*alertState = alertStateInfo{PreviousState: currentState, Timestamp: time.Now().Unix()}
	fmt.Printf("Resolving %v! \t Next alert at %v(from %v)\n", alert.Query, alertState.Timestamp, time.Now().Unix())

	err := _engine.client.Resolve(context.Background(), alert.Name)
	for i := 0; i < 2 && err != nil; i++ {
		fmt.Printf("Retrying resolving %v", alert.Name)
		err = _engine.client.Resolve(context.Background(), alert.Name)
	}
	return err
}

func (_engine engine) notify(alertState *alertStateInfo, alert *Alert, currentState string) error {
	*alertState = alertStateInfo{currentState, time.Now().Unix()}

	fmt.Printf("Sending ALERT %v! \t Next alert at %v(from %v)\n", alert.Query, alertState.Timestamp, time.Now().Unix())

	message := alert.Warn.Message
	if currentState == "CRITICAL" {
		message = alert.Critical.Message
	}

	err := _engine.client.Notify(context.Background(), alert.Name, message)
	for i := 0; i < 2 && err != nil; i++ {
		fmt.Printf("Retrying notifying %v", alert.Name)
		err = _engine.client.Notify(context.Background(), alert.Name, message)
	}
	return err
}

func (_engine engine) getCurrentState(alert *Alert) string {
	queryResponse, err := _engine.client.Query(context.Background(), alert.Query)
	for i := 0; i < 2 && err != nil; i++ {
		fmt.Printf("Retrying getting current state %v", alert.Name)
		queryResponse, err = _engine.client.Query(context.Background(), alert.Query)
	}
	if err != nil {
		log.Printf("Line 86: %v %v\n", err, alert.Name)
	}

	currentValue := queryResponse
	if currentValue <= alert.Warn.Value {
		return "PASS"
	} else if currentValue <= alert.Critical.Value {
		return "WARN"
	}
	return "CRITICAL"
}

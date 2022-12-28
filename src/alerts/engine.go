package alerts

import (
	"context"
	"fmt"
	"log"
	"time"
)

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

//func AlertService(node *alertStateInfo, client *alerts.Client) {
//	currentState := getCurrentState((*node).Data, *client)
//
//	if currentState == "PASS" {
//		fmt.Printf("%v ; %v ; [%v -> %v] ; %v\n", node.Data.Query, currentState, node.PreviousState, currentState, time.Now().Unix())
//		if node.PreviousState != "PASS" {
//			node.PreviousState = currentState
//			node.Timestamp = time.Now().Unix()
//			fmt.Printf("Resolving %v! \t Next alert at %v(from %v)\n", node.Data.Query, node.Timestamp, time.Now().Unix())
//
//			err := (*client).Resolve(context.Background(), node.Data.Name)
//			if err != nil {
//				fmt.Println("Line 24: ", err)
//			}
//		}
//	} else if currentState == "WARN" || currentState == "CRITICAL" {
//		fmt.Printf("%v ; %v ; [%v -> %v] ; %v\n", node.Data.Query, currentState, node.PreviousState, currentState, time.Now().Unix())
//		if currentState != node.PreviousState || time.Now().Unix()-node.Timestamp > node.Data.RepeatIntervalSecs {
//			node.PreviousState = currentState
//			node.Timestamp = time.Now().Unix()
//			fmt.Printf("Sending ALERT %v! \t Next alert at %v(from %v)\n", node.Data.Query, node.Timestamp, time.Now().Unix())
//
//			message := node.Data.Warn.Message
//			if currentState == "CRITICAL" {
//				message = node.Data.Critical.Message
//			}
//
//			err := (*client).Notify(context.Background(), node.Data.Name, message)
//			if err != nil {
//				fmt.Println("Line 43: ", err)
//			}
//		}
//	}
//	time.Sleep(time.Second * time.Duration(node.Data.IntervalSecs))
//	AlertService(node, client)
//}

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
		fmt.Printf("Retrying resolving %v", alert.Name)
		err = _engine.client.Notify(context.Background(), alert.Name, message)
	}
	return err
}

func (_engine engine) getCurrentState(alert *Alert) string {
	queryResponse, err := _engine.client.Query(context.Background(), alert.Query)
	if err != nil {
		log.Printf("Line 71: %v %v\n", err, alert.Name)
	}

	currentValue := queryResponse
	if currentValue <= alert.Warn.Value {
		return "PASS"
	} else if currentValue <= alert.Critical.Value {
		return "WARN"
	}
	return "CRITICAL"
}

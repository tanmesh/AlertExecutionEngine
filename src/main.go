package main

import (
	"context"
	"fmt"
	"github.com/chronosphereio/interviews-alerts-execution-engine/golang/src/alerts"
	"sync"
)

func main() {
	client := alerts.NewClient("")

	alertList, err := client.QueryAlerts(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	engine := alerts.NewEngine(client)

	var wg sync.WaitGroup
	for _, alert := range alertList {
		wg.Add(1)

		alert := alert
		go func() {
			defer wg.Done()
			engine.Run(alert)
		}()
	}

	wg.Wait()
}

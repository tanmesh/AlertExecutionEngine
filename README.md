# Go Client

This implements a simple Go client for the alert execution engine interview. You can test the client using the code
snippet below. First start up the alerts server using
```
docker run -p 9001:9001 quay.io/chronosphereiotest/interview-alerts-engine:latest
```
And Run the main.go using
```
go run main.go
```

The following code snippet available in `engine.go` creates an alert engine interface.
```
// Engine is the alerts engine interface.
type Engine interface {
	Run(alert *Alert) (error)
}

// engine struct contains the alert client
type engine struct {
	client Client
}
```

Alert Engine is responsible for running each alert independently in an infinite loop.
Alert engine uses time ticker to schedule the processing of an alert with given interval seconds.

The following code snippet also available in `main.go` shows how you can use alert engine client.
```
// Alert engine maintains the lifecycle of the alert processing
	engine := alerts.NewEngine(client)
	wg := sync.WaitGroup{}
	wg.Add(len(alertList))

	for _, alert := range alertList {
		// invoke goroutines for each alerts
		go func(localAlert *alerts.Alert) {
			err := engine.Run(localAlert);
			if (err != nil) {
				fmt.Errorf("error running alert: %+v %v\n", localAlert, err)
			}
			wg.Done();
		} (alert)
	}
	wg.Wait()
```

Trade-offs
1. retry : Network or any node failures are very common in distributed systems. In case of transient failures, its always important to retry the request (GET/POST request in case of alert query and notification). We are using exponential backoff with soem jitter with an interval of 100 ms with max delay of 1 seconds. This would be suffiecient to avoid any transient network issues and get the results in such cases. For long standing issues, it doesn't make sense to do longer retry for alert query engine as it is anyway scheduled to repeat the process after certain interval

2. Error handling :  In terms of error handling, we are avoiding intermediate errors while querying the alert values/notifying etc. This is because some transient network errors might cause some failures. Since alert notification engire running the inifite loop, the query will again happen after fixed time interval.

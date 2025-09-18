package main

import "github.com/jaku01/caddyservicediscovery/internal/scheduler"

func main() {

	err := scheduler.StartScheduleDiscovery("http://localhost:2019")
	if err != nil {
		panic(err)
	}
}

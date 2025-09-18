package main

import (
	"fmt"
	"github.com/jaku01/caddyservicediscovery/cmd/internal/caddy"
)

func main() {

	connector := caddy.NewConnector("http://localhost:2019")

	config, err := connector.GetCaddyConfig()
	if err != nil {
		panic(err)
	}

	if config == nil {
		fmt.Println("No config found")
		err = connector.CreateCaddyConfig()
		if err != nil {
			panic(err)
		}
	}

	serverMap := map[string]caddy.Server{
		"server1": caddy.NewReverseProxyServer(8082, "localhost:3000"),
		"server2": caddy.NewReverseProxyServer(8081, "localhost:8080"),
	}

	err = connector.ReplaceServers(serverMap)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"context"
	"log"
	proxy2 "router/proxy"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())
	proxy := proxy2.NewProxy("127.0.0.1", ":9000", ctx)

	//c := make(chan os.Signal, 1)
	//signal.Notify(c, os.Interrupt)
	//go func() {
	//	for sig := range c {
	//		log.Printf("Signal received %v, stopping and exiting...", sig)
	//		cancel()
	//	}
	//}()

	err := proxy.Start("9001")
	if err != nil {
		log.Fatal(err)
	}
}

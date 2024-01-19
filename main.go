package main

import (
	"git/protocol"
	"log"
	"log/slog"
	"net/http"
)

func main() {
	slog.Info("Server initiated")

	service := protocol.NewHTTPServer(protocol.Config{
		Dir:        "./test",
		AutoCreate: true,
		Auth:       false,
	})

	if err := service.Setup(); err != nil {
		log.Fatal(err)
	}

	http.Handle("/", &service)

	slog.Info("Git server running")

	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatal(err)
	}
}

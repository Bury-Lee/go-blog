package main

import (
	"fmt"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/sirupsen/logrus"
)

type MyEventHandler struct {
	canal.DummyEventHandler
}

func (h *MyEventHandler) OnRow(e *canal.RowsEvent) error {
	fmt.Printf("Name: %s Action:%s e:%#v\n", e.Table.Name, e.Action, e)
	return nil
}

func (h *MyEventHandler) String() string {
	return "MyEventHandler"
}

func main() {
	cfg := canal.NewDefaultConfig()
	cfg.Addr = "192.168.0.2:3306"
	cfg.User = "root"
	cfg.Password = "123456"
	cfg.Dump.Databases = []string{"blogx"}
	cfg.Dump.Tables = []string{}

	c, err := canal.NewCanal(cfg)
	if err != nil {
		logrus.Fatal(err)
	}

	// Register a handler to handle RowsEvent
	c.SetEventHandler(&MyEventHandler{})

	// Start canal
	c.Run()
}

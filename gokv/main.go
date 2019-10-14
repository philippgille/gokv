//go:generate go run github.com/philippgille/gokv/gokv/gen > ./store_generated.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func get(c *cli.Context) error {
	key := c.Args().First()
	if len(key) == 0 {
		return errors.New("invalid key")
	}

	config, err := readConfig(c.GlobalString("config"))
	if err != nil {
		return err
	}

	store, err := newStore(*config)
	if err != nil {
		return err
	}

	var value interface{}
	ok, err := store.Get(key, &value)
	if !ok {
		fmt.Println("not found")
	}
	if err != nil {
		return err
	}

	fmt.Println(value)

	return nil
}

func set(c *cli.Context) error {
	key := c.Args().Get(0)
	if len(key) == 0 {
		return errors.New("invalid key")
	}

	value := c.Args().Get(1)
	if len(value) == 0 {
		return errors.New("invalid value")
	}
	var valueJson interface{}
	valueJsonErr := json.Unmarshal([]byte(value), &valueJson)

	config, err := readConfig(c.GlobalString("config"))
	if err != nil {
		return err
	}

	store, err := newStore(*config)
	if err != nil {
		return err
	}

	if valueJsonErr == nil {
		return store.Set(key, valueJson)
	} else {
		return store.Set(key, value)
	}
}

func main() {
	app := cli.NewApp()
	app.Usage = "A cli tool for getting and setting key values"
	app.UsageText = fmt.Sprint(app.Name, " get <KEY>\n   ", app.Name, " set <KEY> <VALUE>")
	app.HideVersion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "./gokv.yaml",
			Usage: "path to config",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "get",
			Usage:     "get key value",
			UsageText: "get <KEY>",
			Action:    get,
		},
		{
			Name:      "set",
			Usage:     "set key value",
			UsageText: "set <KEY> <VALUE>",
			Action:    set,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

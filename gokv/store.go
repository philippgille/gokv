package main

import (
	"errors"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/encoding"
)

func newStore(conf Config) (store gokv.Store, err error) {
	var codec encoding.Codec
	if conf.Encoding == "json" {
		codec = encoding.JSON
	} else if conf.Encoding == "gob" {
		codec = encoding.Gob
	} else {
		return nil, errors.New("error")
	}

	return createStore(conf.Implementation, codec, conf.Options)
}

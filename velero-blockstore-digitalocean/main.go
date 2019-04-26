package main

import (
	veleroplugin "github.com/heptio/velero/pkg/plugin"
	"github.com/sirupsen/logrus"
)

func main() {
	veleroplugin.NewServer(veleroplugin.NewLogger()).
		RegisterBlockStore("digitalocean-blockstore", newBlockStore).
		Serve()
}

func newBlockStore(logger logrus.FieldLogger) (interface{}, error) {
	return &BlockStore{FieldLogger: logger}, nil
}

package main

import (
	"github.com/Bortnyak/file-syncer/pkg/server"
	"github.com/Bortnyak/file-syncer/pkg/watcher"
)

func main() {
	go watcher.Watch()
	server.Main()
}

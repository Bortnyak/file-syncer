package syncer

import (
	"sync"

	"github.com/Bortnyak/file-syncer/pkg/client"
	"github.com/Bortnyak/file-syncer/pkg/server"
	"github.com/Bortnyak/file-syncer/pkg/watcher"
)

func Run() {
	var wg sync.WaitGroup

	wg.Add(3)
	go watcher.Watch()
	go server.Main()
	go client.ListenToUpdates()

	wg.Wait()
}

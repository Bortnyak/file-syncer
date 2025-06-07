package watcher

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Bortnyak/file-syncer/pkg/client"
	"github.com/Bortnyak/file-syncer/pkg/storage"
	"github.com/radovskyb/watcher"
)

func Watch(ctx context.Context) error {
	w := watcher.New()
	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	//
	// If SetMaxEvents is not set, the default is to send all events.
	// w.SetMaxEvents(1)

	// Only notify rename and move events.
	// w.FilterOps(watcher.Rename, watcher.Move)

	// Only files that match the regular expression during file listings
	// will be watched.

	// r := regexp.MustCompile("^abc$")
	// w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		for {
			select {
			case event := <-w.Event:
				// fmt.Println(event) // Print the event's info.
				fmt.Println()
				eventHandler(event)
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			case <-ctx.Done():
				return
			}

		}
	}()

	// Watch this folder for changes.
	if err := w.AddRecursive("./test-folder"); err != nil {
		log.Println("Error while addidng folder to watcher, ", err)
		return err
	}

	// Watch test_folder recursively for changes.
	// if err := w.AddRecursive("./test-folder"); err != nil {
	// 	log.Fatalln(err)
	// }

	// Print a list of all of the files and folders currently
	// being watched and their paths.
	for path, f := range w.WatchedFiles() {
		log.Println("watched = ", path, f.Name())
	}

	fmt.Println()
	// // Trigger 2 events after watcher started.
	// go func() {
	// 	w.Wait()
	// 	w.TriggerEvent(watcher.Create, nil)
	// 	w.TriggerEvent(watcher.Remove, nil)
	// }()

	done := make(chan struct{})
	go func() {
		// Start the watching process - it'll check for changes every 100ms.
		if err := w.Start(time.Millisecond * 100); err != nil {
			done <- struct{}{}
		}
	}()

	select {
	case <-done:
		break
	case <-ctx.Done():
		break
	}

	return nil
}

func eventHandler(event watcher.Event) {
	log.Println("eventHandler started")
	log.Println("Event = ", event)

	// operations declared by iotas:
	// const (
	// 	Create Op = iota (0)
	// 	Write            (1)
	// 	Remove           (2)
	// 	Rename           (3)
	// 	Chmod            (4)
	// 	Move             (5)
	// )
	switch operation := event.Op; operation {
	case 0:
		log.Println("|----------------------------------|")
		log.Println("Upload file to bucket")
		err := storage.UploadFile(event.Path)
		if err != nil {
			log.Println("Error while uploading file: ", err)
			return
		}

		eventPayload := client.UpdateEventPayload{
			Event: "Create",
			Info:  event.Path,
		}
		client.SendUpdate(&eventPayload)

	case 1:
		log.Println("|----------------------------------|")
		log.Println("Write file to bucket")
		storage.UploadFile(event.Path)
	case 2:
		log.Println("|----------------------------------|")
		log.Println("Remove file from the bucket")
		storage.DeleteFile(event.Path)
	}
}

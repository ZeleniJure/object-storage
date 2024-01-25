package main

import (
	"github.com/ZeleniJure/object-storage/objectstorage"
	"github.com/ZeleniJure/object-storage/server"
	"github.com/ZeleniJure/object-storage/storagebackend"
)

func main() {
	server.Config()
	go func() {
		storagebackend.New()
	}()
	server := server.New()

	objectstorage.New(server)

	<-server.Ctx.Done()
}

package main

import (
	"financial-app/internals/database"
	"financial-app/internals/service"
	"financial-app/internals/transport/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// Run - sets up our application
func Run() error {
	// logFile := "/var/log/financial-api.log"
	// f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	// if err != nil {
	// 	fmt.Println("Failed to create logfile", logFile)
	// 	panic(err)
	// }
	// defer f.Close()
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// Only log the debug severity or above
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("setting up financial app")

	store, err := database.NewDatabase()
	if err != nil {
		log.Error("failed to setup connection to the database")
		return err
	}
	err = store.MigrateDB()
	if err != nil {
		log.Error("failed to setup database")
		return err
	}

	transactionService := service.NewService(store)
	handler := http.NewHandler(transactionService)

	if err := handler.Serve(); err != nil {
		log.Error("failed to gracefully serve financial app")
		return err
	}

	return nil
}

func main() {
	if err := Run(); err != nil {
		log.Error(err)
		log.Fatal("Error starting up financial app")
	}
}

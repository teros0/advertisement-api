package main

import (
	"log"
	"net/http"
	"spMDOImages/server/api"
	"spMDOImages/server/conf"
	"spUtils/logger"
)

func main() {
	log.SetOutput(logger.NewLogger("logs/alltogether", conf.LogMaxSize, conf.LogMaxAge, conf.LogMaxBackups))

	if err := conf.InitConfig(); err != nil {
		log.Fatalf("Couldn't initialize configuration %s", err)
	}

	router := http.NewServeMux()
	router = api.RegisterRoutes(router)

	log.Printf("Starting service on %s", conf.ServerAddress)
	log.Fatal(http.ListenAndServe(conf.ServerAddress, router))
}

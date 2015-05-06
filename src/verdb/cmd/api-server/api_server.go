package main

import (
	"log"
	"os"
)

import (
	"verdb"
)

func getenv(key, orVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return orVal
	}
	return val
}

func init() {
	if os.Getenv("IS_DEBUG") == "true" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}
}

func main() {
	var (
		// mongo url
		mgoURL = getenv("MONGO_URL", "localhost")

		// db name to save registries
		metaDb = getenv("META_DB", "metadb")

		// host:port to serve http service
		hostPort = getenv("HOST_POST", "localhost:8888")

		// max number of running jobs
		max = 10
	)

	server := verdb.NewApiServer(mgoURL, metaDb, max, nil)
	server.Run(hostPort)
}

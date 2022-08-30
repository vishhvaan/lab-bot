package db

import (
	"path"

	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"

	"github.com/vishhvaan/lab-bot/pkg/logging"
)

type database struct {
	db     *bolt.DB
	logger *log.Entry
}

var botDB database

func OpenDB() {
	botDB.logger = logging.CreateNewLogger("database", "database")
	exePath := logging.FindExeDir()
	dbPath := path.Join(exePath, dbFile)

	var err error
	botDB.db, err = bolt.Open(dbPath, 0644, nil)
	if err != nil {
		botDB.logger.WithError(err).Panic("database could not be opened")
	}
	defer botDB.db.Close()
}

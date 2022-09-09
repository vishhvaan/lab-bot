package db

import (
	"path"

	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"

	"github.com/vishhvaan/lab-bot/logging"
)

type database struct {
	db     *bolt.DB
	logger *log.Entry
}

var botDB database

func Open() {
	botDB.logger = logging.CreateNewLogger("database", "database")
	exePath := logging.FindExeDir()
	dbPath := path.Join(exePath, dbFile)

	var err error
	botDB.db, err = bolt.Open(dbPath, 0600, nil)
	if err != nil {
		botDB.logger.WithError(err).Panic("database could not be opened")
	} else {
		botDB.logger.Info("Opened database")
	}
	defer botDB.db.Close()
}

func CheckIfBucketExists(bucket string) bool {
	var b *bolt.Bucket
	_ = botDB.db.View(func(tx *bolt.Tx) error {
		b = tx.Bucket([]byte(bucket))
		return nil
	})

	if b == nil {
		return false
	} else {
		return true
	}
}

func CreateBucket(bucket string) {
	err := botDB.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})

	if err != nil {
		botDB.logger.WithError(err).WithField("bucket", bucket).Error("Cannot create bucket")
	}
}

func AddValue(bucket string, key string, value []byte) {
	err := botDB.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(key), value)
		return err
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"bucket": bucket,
			"key":    key,
			"value":  value,
		}).Error("Cannot update database")
	}
}

func ReadValue(bucket string, key string) (value []byte) {
	err := botDB.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		value = b.Get([]byte(key))
		return nil
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"bucket": bucket,
			"key":    key,
		}).Error("Cannot read database")
	}

	return value
}

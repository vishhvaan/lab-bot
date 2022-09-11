package db

import (
	"errors"
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
		botDB.logger.Info("opened database")
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

func CreateBucket(bucket string) error {
	err := botDB.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})

	if err != nil {
		botDB.logger.WithError(err).WithField("bucket", bucket).Error("Cannot create bucket")
	}

	return err
}

func AddValue(bucket string, key string, value []byte) error {
	if !CheckIfBucketExists(bucket) {
		return errors.New("bucket does not exist")
	}

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
		}).Error("cannot update database")
	}

	return err
}

func ReadValue(bucket string, key string) (value []byte, err error) {
	err = botDB.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		value = b.Get([]byte(key))
		return nil
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"bucket": bucket,
			"key":    key,
		}).Error("cannot read database")
	}

	return value, err
}

func GetAllKeysValues(bucket string) (keys [][]byte, values [][]byte, err error) {
	err = botDB.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucket))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, k)
			values = append(values, v)
		}

		return nil
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"bucket": bucket,
		}).Error("cannot get keys or values in this bucket")
	}

	return keys, values, err
}

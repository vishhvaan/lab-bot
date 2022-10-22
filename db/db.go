package db

import (
	"errors"
	"os"
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
		botDB.logger.WithError(err).Panic("Database could not be opened")
	} else {
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			botDB.logger.Info("Created database")
		} else {
			botDB.logger.Info("Opened database")
		}
	}
}

func Close() {
	botDB.db.Close()
	botDB.logger.Info("Closed database")
}

func bucketFinder(tx *bolt.Tx, path []string) (b *bolt.Bucket, err error) {
	notExist := errors.New("bucket does not exist at path")
	var bucket *bolt.Bucket
	if len(path) == 0 {
		return nil, notExist
	} else {
		bucket = tx.Bucket([]byte(path[0]))
		path = path[1:]
		for len(path) > 0 && bucket != nil {
			bucket = bucket.Bucket([]byte(path[0]))
			path = path[1:]
		}
	}

	if bucket == nil {
		return nil, notExist
	} else {
		return bucket, nil
	}
}

func bucketCreator(tx *bolt.Tx, path []string) (b *bolt.Bucket, err error) {
	if len(path) == 0 {
		return nil, errors.New("cannot create empty bucket")
	} else {
		p := path
		bucket, err := tx.CreateBucketIfNotExists([]byte(p[0]))
		if err != nil {
			return nil, err
		}
		p = p[1:]
		for len(p) > 0 {
			bucket, err = bucket.CreateBucketIfNotExists([]byte(p[0]))
			if err != nil {
				return nil, err
			}
			p = p[1:]
		}

		botDB.logger.WithField("path", path).Info("Created bucket")

		return bucket, nil
	}
}

func CheckBucketExists(path []string) (exists bool) {
	err := botDB.db.View(func(tx *bolt.Tx) error {
		b, err := bucketFinder(tx, path)
		if b != nil {
			exists = true
			return nil
		}
		return err
	})

	l := botDB.logger.WithError(err).WithFields(log.Fields{
		"path": path,
	})
	if err.Error() == "bucket does not exist at path" {
		l.Info("Bucket does not exist at path")
	} else if err != nil {
		l.Error("Cannot check if bucket exists")
	}

	return exists
}

func CreateBucket(path []string) error {
	err := botDB.db.Update(func(tx *bolt.Tx) error {
		_, err := bucketCreator(tx, path)
		return err
	})

	if err != nil {
		botDB.logger.WithError(err).WithField("path", path).Error("Cannot create bucket at path")
	}

	return err
}

func AddValue(path []string, key string, value []byte) error {
	err := botDB.db.Update(func(tx *bolt.Tx) error {
		b, err := bucketFinder(tx, path)
		if err != nil {
			return err
		}
		err = b.Put([]byte(key), value)
		return err
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"path":  path,
			"key":   key,
			"value": value,
		}).Error("Cannot update database")
	}

	return err
}

func ReadValue(path []string, key string) (value []byte, err error) {
	// returns nil if key doesn't exist or is a nested bucket
	err = botDB.db.View(func(tx *bolt.Tx) error {
		b, err := bucketFinder(tx, path)
		if err != nil {
			return err
		}
		value = b.Get([]byte(key))
		return nil
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"path": path,
			"key":  key,
		}).Error("Cannot read database")
	}

	return value, err
}

func DeleteValue(path []string, key string) error {
	err := botDB.db.Update(func(tx *bolt.Tx) error {
		b, err := bucketFinder(tx, path)
		if err != nil {
			return err
		}
		err = b.Delete([]byte(key))
		return err
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"path": path,
			"key":  key,
		}).Error("Cannot delete key from database")
	}

	return err
}

func GetAllKeysValues(path []string) (keys [][]byte, values [][]byte, err error) {
	err = botDB.db.View(func(tx *bolt.Tx) error {
		b, err := bucketFinder(tx, path)
		if err != nil {
			return err
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, k)
			values = append(values, v)
		}

		return nil
	})

	if err != nil {
		botDB.logger.WithError(err).WithFields(log.Fields{
			"path": path,
		}).Error("Cannot get keys or values in this bucket")
	}

	return keys, values, err
}

func IncrementBucketInteger(path []string) (i int, err error) {
	err = botDB.db.Update(func(tx *bolt.Tx) error {
		b, err := bucketFinder(tx, path)
		if err != nil {
			return err
		}
		i64, err := b.NextSequence()
		i = int(i64)
		return err
	})

	return i, err
}

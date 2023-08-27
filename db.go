package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"
)

const ( // Buckets
	counterEndingBucket   = "tick"
	counterInfiniteBucket = "infiniteBucket"
	roleBucket            = "roles"
	chatIDBucket          = "chatID"
	intervalBucket        = "interval"
)
const ( // Important Users id
	daunitseID  = 841547487
	carpawellID = 209328250
)
const ( // Bullshit you need to work with bbolt
	lastChatIDKey  = "lastChatIDKey"
	userRWFileMode = 0600
	uint32Size     = 4
	uint64Size     = 8
	lastInterval   = "lastInterval"
)
const ( // Bullshit you need to work with telegram
	privateCHat = "private"
	easterEgg   = "easter Egg"
)
const ( // Roles
	bomzRole = iota // не пользователь
	userRole        // пользователь, которому разрешено общаться с ботом
	adminRole
)

type database struct {
	b *bolt.DB
}

func newDb(path string) (*database, error) {
	db, err := bolt.Open(path, userRWFileMode, bolt.DefaultOptions)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	createBucketsFunc := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(counterEndingBucket))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(counterInfiniteBucket))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(chatIDBucket))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(intervalBucket))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(roleBucket))
		return err
	}

	err = db.Update(createBucketsFunc)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("could not init database: %s", err)
	}

	return &database{
		b: db,
	}, nil
}

// IncUserMessages increments user's counter and returns the updated value.
// Returns any error that did not allow to perform the operation.
func (db *database) IncUserMessages(user string) error {
	err := db.b.Update(func(tx *bolt.Tx) error {
		b1 := tx.Bucket([]byte(counterEndingBucket))
		if b1 == nil {
			return fmt.Errorf("expected %s bucket does not exits", counterEndingBucket)
		}

		b2 := tx.Bucket([]byte(counterInfiniteBucket))
		if b2 == nil {
			return fmt.Errorf("expected %s bucket does not exits", counterInfiniteBucket)
		}
		var updatedEndingValue uint32

		userKey := []byte(user)
		val1 := b1.Get(userKey)
		if val1 != nil {
			updatedEndingValue = binary.LittleEndian.Uint32(val1)
		}

		updatedEndingValue++

		val1 = make([]byte, uint32Size)
		binary.LittleEndian.PutUint32(val1, updatedEndingValue)

		if err := b1.Put(userKey, val1); err != nil {
			return err
		}
		var updatedInfiniteValue uint32

		val2 := b2.Get(userKey)
		if val2 != nil {
			updatedInfiniteValue = binary.LittleEndian.Uint32(val2)
		}

		updatedInfiniteValue++

		val2 = make([]byte, uint32Size)
		binary.LittleEndian.PutUint32(val2, updatedInfiniteValue)

		return b2.Put(userKey, val2)
	})
	if err != nil {
		return fmt.Errorf("could not update val in db: %w", err)
	}

	return nil
}

func (db *database) GiveUserRoles(user int64, role byte) error {
	giveUserRole := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(roleBucket))
		if b == nil {
			return fmt.Errorf("expected %s bucket does not exist", roleBucket)
		}
		userKey := make([]byte, uint64Size)
		binary.LittleEndian.PutUint64(userKey, uint64(user))
		val := []byte{role}

		return b.Put(userKey, val)
	}

	err := db.b.Update(giveUserRole)
	if err != nil {
		return fmt.Errorf("could not assign role: %w", err)
	}
	return nil
}

func (db *database) UserRole(user int64) (byte, error) {
	var role byte

	userRole := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(roleBucket))
		if b == nil {
			return fmt.Errorf("expected %s bucket does not exist", roleBucket)
		}
		userKey := make([]byte, uint64Size)
		binary.LittleEndian.PutUint64(userKey, uint64(user))
		val := b.Get(userKey)
		if val != nil {
			role = val[0]
		}
		return nil
	}

	err := db.b.View(userRole)
	if err != nil {
		return 0, fmt.Errorf("could not check user`s role: %w", err)
	}

	return role, nil
}

type UserStat struct {
	Name         string
	MessageCount uint32
}

func (db *database) GetUsersEndingStat() ([]UserStat, error) {
	var stat []UserStat
	getUsersStat := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(counterEndingBucket))
		if b == nil {
			return fmt.Errorf("expected %s bucket does not exist", counterEndingBucket)
		}
		return b.ForEach(func(k, v []byte) error {
			var us UserStat
			us.Name = string(k)
			us.MessageCount = binary.LittleEndian.Uint32(v)
			stat = append(stat, us)

			return nil
		})
	}
	err := db.b.View(getUsersStat)
	if err != nil {
		return nil, fmt.Errorf("could not get all stat from db: %w", err)
	}

	return stat, nil
}

func (db *database) GetUsersInfiniteStat() ([]UserStat, error) {
	var stat []UserStat
	getUsersStat := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(counterInfiniteBucket))
		if b == nil {
			return fmt.Errorf("expected %s bucket does not exist", counterEndingBucket)
		}
		return b.ForEach(func(k, v []byte) error {
			var us UserStat
			us.Name = string(k)
			us.MessageCount = binary.LittleEndian.Uint32(v)
			stat = append(stat, us)

			return nil
		})
	}
	err := db.b.View(getUsersStat)
	if err != nil {
		return nil, fmt.Errorf("could not get all stat from db: %w", err)
	}

	return stat, nil
}

func (db *database) ResetBucket(bucket string) error {
	return db.b.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *database) DeleteBucket(bucket string) error {
	err := db.b.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(bucket))
	})
	if err != nil {
		return err
	}

	return nil
}

func (db *database) CreateBucket(bucket string) error {
	return db.b.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})
}

// Close closes the underlying database.
func (db *database) Close() error {
	return db.b.Close()
}

func (db *database) SaveChatID(chatID int64) error {
	return db.b.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(chatIDBucket))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(lastChatIDKey), []byte(fmt.Sprintf("%d", chatID)))
	})
}

func (db *database) ReadChatID() (int64, error) {
	var chatID int64
	err := db.b.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(chatIDBucket))
		if bucket == nil {
			return nil
		}
		value := bucket.Get([]byte(lastChatIDKey))
		if value == nil {
			return nil
		}
		var err error
		chatID, err = strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return chatID, nil
}

func (db *database) SaveStatsInterval(interval time.Duration) error {
	return db.b.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(intervalBucket))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(lastInterval), []byte(interval.String()))
	})
}

func (db *database) ReadSendStatsInterval() (time.Duration, error) {
	var interval time.Duration
	err := db.b.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(intervalBucket))
		if bucket == nil {
			return nil
		}
		value := bucket.Get([]byte(lastInterval))
		if value == nil {
			return nil
		}
		var err error
		interval, err = time.ParseDuration(string(value))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return interval, nil
}

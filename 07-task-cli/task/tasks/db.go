package tasks

import (
	"encoding/binary"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB
var taskBucket = []byte("tasks")

type Task struct {
	Key   int
	Value string
}

func InitDb(path string) error {
	var err error
	fmt.Println("Opening db at path:", path)

	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		fmt.Printf("error opening db: %s", err)
		return err
	}
	err = initBuckets(db)
	if err != nil {
		fmt.Printf("error occurred initializing buckets: %s", err)
		return err
	}
	return err
}

func initBuckets(db *bolt.DB) error {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(taskBucket)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func CloseDb() error {
	fmt.Println("shutting down db")
	err := db.Close()
	return err
}

func AddTask(desc string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)
		id, err := b.NextSequence()
		if err != nil {
			fmt.Printf("error getting next sequence in bucket: %s", err)
			return err
		}
		key := itob(id)
		val := []byte(desc)
		err = b.Put(key, val)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func GetTasks() ([]Task, error) {
	var tasks []Task
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			tasks = append(tasks, Task{Key: btoi(k), Value: string(v)})
		}
		return nil
	})
	return tasks, err
}

func DoTask(id int) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)
		key := itob(uint64(id))
		err := b.Delete(key)
		return err
	})
	return err
}

func itob(u uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, u)
	return b
}

func btoi(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}

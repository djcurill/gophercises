package tasks

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

func InitDb(path string) error {
	var err error
	fmt.Println("Opening db at path:", path)
	db, err = bolt.Open(path, 0600, nil)
	return err
}

func CloseDb() error {
	fmt.Println("shutting down db")
	err := db.Close()
	return err
}

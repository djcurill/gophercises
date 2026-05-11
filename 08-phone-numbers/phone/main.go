package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/djcurill/phone/db"
	_ "github.com/jackc/pgx/v5/stdlib" // blank import registers the driver
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

func connectionString() (string, error) {
	var err error = nil

	db := os.Getenv("DB_NAME")
	if db == "" {
		db = "app"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		return "", fmt.Errorf("unable to build connection string without DB_USER env var")
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		return "", fmt.Errorf("unable to build connection string without DB_PASSWORD env var")
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, db), err
}

func main() {
	connString, err := connectionString()
	if err != nil {
		panic(err)
	}
	DB, err := db.Open(connString)
	if err != nil {
		panic(err)
	}
	defer DB.Close()

	err = DB.Migrate()
	if err != nil {
		panic(err)
	}

	err = DB.Seed()
	if err != nil {
		panic(err)
	}

	phoneNumbers, err := DB.GetAllPhoneNumbers()
	if err != nil {
		panic(err)
	}

	for _, p := range phoneNumbers {
		pNorm := normalize(p)
		if p != pNorm {
			existing, err := DB.FindPhone(pNorm)
			if err != nil {
				panic(err)
			}
			if existing != nil {
				fmt.Println("Phone number already exists! Deleting duplicate")
				err := DB.DeletePhone(p)
				if err != nil {
					panic(err)
				}
			} else {
				fmt.Println("Updating unnormalized phone record")
				err := DB.UpdateNumber(pNorm)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func normalize(p db.Phone) db.Phone {
	re := regexp.MustCompile(`[\(\)\s\-]+`)
	pNorm := re.ReplaceAllString(p.Number, "")
	return db.Phone{ID: p.ID, Number: pNorm}
}

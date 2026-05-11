package db

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib" // blank import registers the driver
)

type Phone struct {
	ID     int
	Number string
}

type DB struct {
	db *sql.DB
}

func Open(connectionString string) (*DB, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) Close() error {
	err := db.db.Close()
	return err
}

func (db *DB) Seed() error {
	data := []string{
		"1234567890",
		"123 456 7891",
		"(123) 456 7892",
		"(123) 456-7893",
		"123-456-7894",
		"123-456-7890",
		"1234567892",
		"(123)456-7892",
	}
	for _, phoneNumber := range data {
		err := insertPhoneNumber(db.db, phoneNumber)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) Migrate() error {
	_, err := db.db.Exec(`DROP TABLE IF EXISTS public.phone_numbers;`)
	if err != nil {
		return err
	}

	_, err = db.db.Exec(`
		CREATE TABLE IF NOT EXISTS phone_numbers (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL
		);
		`)
	return nil
}

func (db *DB) GetAllPhoneNumbers() ([]Phone, error) {
	phoneNumbers := []Phone{}
	rows, err := db.db.Query(`select * from public.phone_numbers;`)
	if err != nil {
		return phoneNumbers, err
	}

	for rows.Next() {
		var p Phone
		if err := rows.Scan(&p.ID, &p.Number); err != nil {
			return phoneNumbers, nil
		}
		phoneNumbers = append(phoneNumbers, p)
	}
	return phoneNumbers, nil
}

func (db *DB) UpdateNumber(p Phone) error {
	_, err := db.db.Exec(`UPDATE phone_numbers SET value = $2 WHERE id = $1`, p.ID, p.Number)
	return err
}

func (db *DB) FindPhone(p Phone) (*Phone, error) {
	var existing Phone
	if err := db.db.QueryRow(`select * from phone_numbers where value = $1`, p.Number).Scan(&existing.ID, &existing.Number); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &existing, nil
}

func (db *DB) DeletePhone(p Phone) error {
	_, err := db.db.Exec("delete from phone_numbers where id = $1", p.ID)
	return err
}

func insertPhoneNumber(db *sql.DB, phoneNumber string) error {
	_, err := db.Exec(`
		insert into public.phone_numbers (value)
		values ($1)
		`, phoneNumber)
	return err
}

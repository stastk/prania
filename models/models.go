package models

import (
	"encoding/json"
	"log"
	"math/rand"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config //

var DBpr *sqlx.DB

// Test DB name
var DBname = "prania_exp"

// Test DB username
var DBusername = "gouser"

// Test DB password
var DBusepassword = "gopassword"

var err error

// TODO schema must be placed into another file
var schema = `
	CREATE TABLE IF NOT EXISTS Ingridients (
		id serial PRIMARY KEY,
		name VARCHAR(50) UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS IngridientsVariations (
		id serial PRIMARY KEY,
		name VARCHAR(50) UNIQUE NOT NULL,
		ingridient_id int NOT NULL,
		FOREIGN KEY (ingridient_id) REFERENCES Ingridients(id) ON UPDATE CASCADE
	);
	`

// Temporary struct for getting Variations as string
type IngridientWithVariations struct {
	Id         int    `db:"id" json:"id"`
	Name       string `db:"name" json:"name"`
	Variations string `db:"variations" json:"variations"`
}

// Ingridient and all Variations attached to it
type Ingridient struct {
	Id         int         `db:"id" json:"id"`
	Name       string      `db:"name" json:"name"`
	Variations []Variation `db:"variations" json:"variations"`
}

// Variation of specific Ingridient
type Variation struct {
	Id           int    `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	IngridientId int    `db:"ingridient_id" json:"ingridient_id"`
}

// Single Ingridient with Variations
func IngridientShow(id int) ([]Ingridient, error) {

	ingridients := []Ingridient{}

	rows, err := DBpr.Queryx(`
		SELECT
			i.id,
			i.name,
			COALESCE(json_agg(v) FILTER (WHERE v.id IS NOT NULL), '[]') AS variations
			FROM Ingridients i
			LEFT JOIN IngridientsVariations v ON v.ingridient_id = i.id
			WHERE i.id = ` + strconv.Itoa(id) + `
			GROUP BY i.id;
	`)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var record IngridientWithVariations
		var toappend Ingridient

		if err := rows.StructScan(&record); err != nil {
			panic(err)
		}

		in := []byte(record.Variations)
		variations := []Variation{}
		err := json.Unmarshal(in, &variations)

		if err != nil {
			log.Print(err)
		}

		toappend = Ingridient{record.Id, record.Name, variations}
		ingridients = append(ingridients, toappend)

	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return ingridients, nil

}

// All Ingridients with their Variations
func AllIngridients() ([]Ingridient, error) {

	ingridients := []Ingridient{}

	rows, err := DBpr.Queryx(`
		SELECT
			i.id,
			i.name,
			COALESCE(json_agg(v) FILTER (WHERE v.id IS NOT NULL), '[]') AS variations
			FROM Ingridients i
			LEFT JOIN IngridientsVariations v ON v.ingridient_id = i.id
			GROUP BY i.id;
	`)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var record IngridientWithVariations
		var toappend Ingridient

		if err := rows.StructScan(&record); err != nil {
			panic(err)
		}

		in := []byte(record.Variations)
		variations := []Variation{}
		err := json.Unmarshal(in, &variations)

		if err != nil {
			log.Print(err)
		}

		toappend = Ingridient{record.Id, record.Name, variations}
		ingridients = append(ingridients, toappend)

	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return ingridients, nil

}

// Drop tables #db
func DropDB() (string, error) {
	answer, err := DBpr.Queryx(`
		DROP TABLE IF EXISTS IngridientsVariations;
		DROP TABLE IF EXISTS Ingridients;
	`)
	if err != nil {
		return "", err
	}

	defer answer.Close()

	return "Drop table -done", err
}

// Create tables without anything #db
func InitDB() (string, error) {

	// Create tables
	DBpr.MustExec(schema)

	// Populate DB
	tx := DBpr.MustBegin()
	tx.MustExec(`
		INSERT INTO Ingridients
			(name)
		VALUES
			('Ganash'),
			('Salt'),
			('Pepper'),
			('Onion'),
			('Sugar'),
			('Milk');

		INSERT INTO IngridientsVariations
			(name, ingridient_id)
		VALUES 
			('Ganash of the north', 1),
			('Ganash of the south', 1),
			('Salt of the sea', 2),
			('Ganash of the west', 1),
			('Pepper of the Iron Man', 3),
			('Ganash of the east', 1),
			('Chilli pepper', 3),
			('Pepper', 3),
			('Red hot', 3),
			('Spicy pepper', 3),
			('Dr.Pepper', 3);
	`)

	tx.Commit()

	return "Create tables -done", nil
}

// Populate tables with some data #db
func PopulateDB(count int, minchild int, maxchild int) (string, error) {

	// TODO refactoring needed
	min := 4
	max := 16

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	tx := DBpr.MustBegin()

	for i := 0; i < count; i++ {
		randomChildCount := rand.Intn(maxchild-minchild) + minchild
		randomName := make([]byte, rand.Intn(max-min)+min)
		for li := range randomName {
			randomName[li] = letterBytes[rand.Int63()%int64(len(letterBytes))]
		}

		var ingridient Ingridient
		var variation Variation

		// Create some Ingridients
		err = tx.QueryRowx(`INSERT INTO Ingridients (name) VALUES ($1) RETURNING *;`, randomName).StructScan(&ingridient)
		if err != nil {
			tx.Rollback()
			return "", err
		}
		log.Println("Ingridient:: ", ingridient.Id, ingridient.Name)

		for ich := 0; ich < randomChildCount; ich++ {
			// Add to createt earlier Ingridients Variations
			err = tx.QueryRowx(`INSERT INTO IngridientsVariations (name, ingridient_id) VALUES ($1, $2) RETURNING *;`, string(randomName)+"_"+strconv.Itoa(ich), ingridient.Id).StructScan(&variation)
			if err != nil {
				tx.Rollback()
				return "", err
			}

		}

	}

	tx.Commit()

	return "Populate -done", nil
}

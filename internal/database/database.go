package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Database struct {
	path string
	mux  *sync.RWMutex
}

type DatabaseSchema struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

func (db *Database) loadDatabase() (DatabaseSchema, error) {
	var data DatabaseSchema

	db.mux.RLock()
	defer db.mux.RUnlock()
	fmt.Printf("Opening %v...\n", db.path)

	rawData, err := os.ReadFile(db.path)

	if err != nil {
		fmt.Printf("Error opening database file: %v\n", err.Error())
		return data, err
	}

	fmt.Printf("%v bytes read\n", len(rawData))

	if len(rawData) > 0 {
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			fmt.Println("There was an error unmarshalling JSON data.")
			fmt.Printf("Data: %v\n", rawData)
			return data, err
		}
	}

	return data, nil
}

func (db *Database) writeDatabase(data DatabaseSchema) error {
	var rawData []byte

	db.mux.Lock()
	defer db.mux.Unlock()

	rawData, err := json.Marshal(data)

	if err != nil {
		fmt.Printf("Error marshalling JSON data: %v\n", err.Error())
		return err
	}

	err = os.WriteFile(db.path, rawData, 0600) // 0600 is R/W?

	if err != nil {
		fmt.Printf("Error writing to database: %v\n", err.Error())
		return err
	}

	return nil
}

func NewDatabase(path string) (*Database, error) {
	database := Database{
		path: path,
		mux:  &sync.RWMutex{},
	}

	// Attempt to create file-based DB
	if _, err := os.Stat(path); err == nil {
		database.loadDatabase()
		return &database, err
	}

	f, err := os.Create(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()
	return &database, nil
}

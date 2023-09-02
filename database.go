package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

func (db *Database) CreateChirp(body string) (Chirp, error) {
	var chirp Chirp

	database, err := db.loadDatabase()

	if err != nil {
		return chirp, err
	}

	newId := len(database.Chirps) + 1

	chirp = Chirp{
		Id:   newId,
		Body: body,
	}

	fmt.Printf("New Chirp:\n")
	fmt.Printf("Id: %v\n", chirp.Id)
	fmt.Printf("Body: %v\n", chirp.Body)

	if database.Chirps == nil {
		database.Chirps = make(map[int]Chirp)
	}

	database.Chirps[newId] = chirp
	err = db.writeDatabase(database)

	if err != nil {
		fmt.Printf("Error writing database: %v\n", err.Error())
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *Database) ReadSingleChirp(id int) (Chirp, error) {
	var chirp Chirp
	database, err := db.loadDatabase()

	if err != nil {
		fmt.Printf("Error reading chirps: %v\n", err.Error())
		return chirp, err
	}

	fmt.Printf("Chirps: %v\n", database.Chirps)

	chirp, ok := database.Chirps[id]

	if !ok {
		return Chirp{}, os.ErrNotExist
	}

	return chirp, nil
}

func (db *Database) ReadChirps() ([]Chirp, error) {
	var chirps []Chirp
	database, err := db.loadDatabase()

	if err != nil {
		fmt.Printf("Error reading chirps: %v\n", err.Error())
		return chirps, err
	}

	fmt.Printf("Chirps: %v\n", database.Chirps)

	for _, val := range database.Chirps {
		fmt.Printf("Loaded chirp: %v\n", val)
		chirps = append(chirps, val)
	}

	// Sort asc by ID
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id })
	return chirps, nil
}

func (db *Database) CreateUser(email string) (User, error) {
	var user User

	database, err := db.loadDatabase()

	if err != nil {
		return user, err
	}

	newId := len(database.Users) + 1

	user = User{
		Id:    newId,
		Email: email,
	}

	fmt.Printf("New User:\n")
	fmt.Printf("Id: %v\n", user.Id)
	fmt.Printf("Email: %v\n", user.Email)

	if database.Users == nil {
		database.Users = make(map[int]User)
	}

	database.Users[newId] = user
	err = db.writeDatabase(database)

	if err != nil {
		fmt.Printf("Error writing database: %v\n", err.Error())
		return User{}, err
	}

	return user, nil
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

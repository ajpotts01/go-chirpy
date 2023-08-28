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

type ChirpData struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func (db *Database) CreateChirp(body string) (Chirp, error) {
	var chirp Chirp
	var maxId, newId int

	chirps, err := db.loadDatabase()

	if err != nil {
		return chirp, err
	}

	for maxId = range chirps.Chirps {
		break
	}

	for nextId := range chirps.Chirps {
		if nextId > maxId {
			maxId = nextId
		}
	}

	newId = maxId + 1

	chirp = Chirp{
		Id:   newId,
		Body: body,
	}

	fmt.Printf("New Chirp:\n")
	fmt.Printf("Id: %v\n", chirp.Id)
	fmt.Printf("Body: %v\n", chirp.Body)

	if chirps.Chirps == nil {
		chirps.Chirps = make(map[int]Chirp)
	}

	chirps.Chirps[newId] = chirp
	err = db.writeDatabase(chirps)

	if err != nil {
		fmt.Printf("Error writing database: %v\n", err.Error())
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *Database) ReadChirp() ([]Chirp, error) {
	var chirps []Chirp
	chirpData, err := db.loadDatabase()

	if err != nil {
		fmt.Printf("Error reading chirps: %v\n", err.Error())
		return chirps, err
	}

	fmt.Printf("Chirps: %v\n", chirpData)

	for _, val := range chirpData.Chirps {
		fmt.Printf("Loaded chirp: %v\n", val)
		chirps = append(chirps, val)
	}

	// Sort asc by ID
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id })
	return chirps, nil
}

func (db *Database) loadDatabase() (ChirpData, error) {
	var rawData []byte
	var chirpData ChirpData

	db.mux.Lock()
	defer db.mux.Unlock()
	fmt.Printf("Opening %v...\n", db.path)
	f, err := os.OpenFile(db.path, os.O_RDWR, 0666) // 0666 = read/write?

	defer f.Close()

	if err != nil {
		fmt.Printf("Error opening database file: %v\n", err.Error())
		return chirpData, err
	}

	bytesRead, err := f.Read(rawData)

	if err != nil {
		fmt.Printf("Error loading data: %v\n", err)
		return chirpData, err
	}

	fmt.Printf("%v bytes read\n", bytesRead)

	if bytesRead > 0 {
		err = json.Unmarshal(rawData, &chirpData)
		if err != nil {
			return chirpData, err
		}
	}

	return chirpData, nil
}

func (db *Database) writeDatabase(chirpData ChirpData) error {
	var rawData []byte

	db.mux.Lock()
	defer db.mux.Unlock()

	f, err := os.OpenFile(db.path, os.O_RDWR, 0666) // 0666 = read/write?
	defer f.Close()

	if err != nil {
		fmt.Printf("Error opening database path: %v\n", err.Error())
		return err
	}

	fmt.Printf("Chirp data to write: %v\n", chirpData.Chirps)
	rawData, err = json.Marshal(chirpData)

	if err != nil {
		return err
	}

	// Blow away contents
	f.Truncate(0)
	f.Seek(0, 0)
	_, err = f.Write(rawData)

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

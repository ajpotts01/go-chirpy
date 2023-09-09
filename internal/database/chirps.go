package database

import (
	"log"
	"os"
	"sort"
)

type Chirp struct {
	Body     string `json:"body"`
	Id       int    `json:"id"`
	AuthorId int    `json:"author_id"`
}

func (db *Database) CreateChirp(body string, authorId int) (Chirp, error) {
	var chirp Chirp

	database, err := db.loadDatabase()

	if err != nil {
		return chirp, err
	}

	newId := len(database.Chirps) + 1

	chirp = Chirp{
		Id:       newId,
		Body:     body,
		AuthorId: authorId,
	}

	log.Printf("New Chirp:\n")
	log.Printf("Id: %v\n", chirp.Id)
	log.Printf("Author Id: %v\n", chirp.AuthorId)
	log.Printf("Body: %v\n", chirp.Body)

	if database.Chirps == nil {
		database.Chirps = make(map[int]Chirp)
	}

	database.Chirps[newId] = chirp
	err = db.writeDatabase(database)

	if err != nil {
		log.Printf("Error writing database: %v\n", err.Error())
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *Database) ReadSingleChirp(id int) (Chirp, error) {
	var chirp Chirp
	database, err := db.loadDatabase()

	if err != nil {
		log.Printf("Error reading chirps: %v\n", err.Error())
		return chirp, err
	}

	log.Printf("Chirps: %v\n", database.Chirps)

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
		log.Printf("Error reading chirps: %v\n", err.Error())
		return chirps, err
	}

	log.Printf("Chirps: %v\n", database.Chirps)

	for _, val := range database.Chirps {
		log.Printf("Loaded chirp: %v\n", val)
		chirps = append(chirps, val)
	}

	// Sort asc by ID
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].Id < chirps[j].Id })
	return chirps, nil
}

func (db *Database) DeleteSingleChirp(id int) error {
	database, err := db.loadDatabase()

	if err != nil {
		log.Printf("Error loading database %v\n", err.Error())
		return err
	}

	if _, ok := database.Chirps[id]; ok {
		delete(database.Chirps, id)
		err = db.writeDatabase(database)

		if err != nil {
			log.Printf("Error writing database: %v\n", err.Error())
			return err
		}
	}

	return nil
}

package database

import (
	"log"
	"time"
)

func (db *Database) RevokeToken(token string) error {
	database, err := db.loadDatabase()

	if err != nil {
		return err
	}

	if database.RevokedTokens == nil {
		database.RevokedTokens = make(map[string]string)
	}

	database.RevokedTokens[token] = time.Now().UTC().String()
	err = db.writeDatabase(database)

	if err != nil {
		log.Printf("Error writing database: %v\n", err.Error())
		return err
	}

	return nil
}

func (db *Database) IsTokenRevoked(token string) (bool, error) {
	database, err := db.loadDatabase()

	if err != nil {
		return false, err
	}

	if database.RevokedTokens == nil {
		return false, nil
	}

	if _, ok := database.RevokedTokens[token]; ok {
		return true, nil
	} else {
		return false, nil
	}

}

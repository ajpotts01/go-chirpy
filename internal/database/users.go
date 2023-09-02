package database

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Password []byte `json:"password"`
	Email    string `json:"email"`
	Id       int    `json:"id"`
}

func (db *Database) CreateUser(email string, password string) (User, error) {
	var user User

	database, err := db.loadDatabase()

	if err != nil {
		return user, err
	}

	newId := len(database.Users) + 1

	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return User{}, err
	}

	user = User{
		Id:       newId,
		Email:    email,
		Password: hashPass,
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

func (db *Database) ReadUser(id int) (User, error) {
	var user User
	database, err := db.loadDatabase()

	if err != nil {
		return User{}, err
	}

	user, ok := database.Users[id]

	if !ok {
		return User{}, os.ErrNotExist
	}

	return user, nil
}

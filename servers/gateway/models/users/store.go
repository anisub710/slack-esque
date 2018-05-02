package users

import (
	"errors"

	"github.com/info344-s18/challenges-ask710/servers/gateway/indexes"
)

//ErrUserNotFound is returned when the user can't be found
var ErrUserNotFound = errors.New("user not found")

//Store represents a store for Users
type Store interface {
	//GetByID returns the User with the given ID
	GetByID(id int64) (*User, error)

	//GetByEmail returns the User with the given email
	GetByEmail(email string) (*User, error)

	//GetByUserName returns the User with the given Username
	GetByUserName(username string) (*User, error)

	//Insert inserts the user into the database, and returns
	//the newly-inserted User, complete with the DBMS-assigned ID
	Insert(user *User) (*User, error)

	//Update applies UserUpdates to the given user ID
	//and returns the newly-updated user
	Update(id int64, updates *Updates) (*User, error)

	//UpdatePhoto updates the photourl for a user
	UpdatePhoto(id int64, photourl string) (*User, error)

	//Delete deletes the user with the given ID
	Delete(id int64) error

	//InsertLogin inserts login activity
	InsertLogin(login *Login) (*Login, error)

	//UpdatePassword updates password after resetting it.
	UpdatePassword(id int64, passHash []byte) (*User, error)

	//LoadUsers gets all users to add to the trie
	LoadUsers() (*indexes.Trie, error)

	//GetSearchUsers gets all users based on the found Ids
	GetSearchUsers(found []int64) (*[]User, error)
}

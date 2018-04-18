package users

import (
	"database/sql"
	"fmt"
)

//MySQLStore represents a users.Store backed by MySQL
type MySQLStore struct {
	db *sql.DB
}

//NewMySQLStore constructs a new MySQLStore.
func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{
		db: db,
	}
}

func (s *MySQLStore) getBase(param string, value interface{}) (*User, error) {
	query := fmt.Sprintf("select id, email, passhash, username, firstname, lastname, photourl from users where %v=?", param)
	user := &User{}

	err := s.db.QueryRow(query, value).Scan(&user.ID, &user.Email, &user.PassHash,
		&user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL)
	switch {
	case err == sql.ErrNoRows:
		return nil, ErrUserNotFound
	case err != nil:
		return nil, err
	}

	return user, nil

}

//GetByID returns the User with the given ID
func (s *MySQLStore) GetByID(id int64) (*User, error) {
	return s.getBase("id", id)
}

//GetByEmail returns the User with the given email
func (s *MySQLStore) GetByEmail(email string) (*User, error) {
	return s.getBase("email", email)
}

//GetByUserName returns the User with the given Username
func (s *MySQLStore) GetByUserName(username string) (*User, error) {
	return s.getBase("username", username)
}

//Insert inserts the user into the database, and returns
//the newly-inserted User, complete with the DBMS-assigned ID
func (s *MySQLStore) Insert(user *User) (*User, error) {
	insq := "insert into users(email, passhash, username, firstname, lastname, photourl) values (?,?,?,?,?,?)"
	res, err := s.db.Exec(insq, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)

	if err != nil {
		return nil, fmt.Errorf("Error executing insert: %v", err)
	}

	id, err := res.LastInsertId()

	if err != nil {
		return nil, fmt.Errorf("Error getting last id: %v", err)
	}

	user.ID = id

	return user, nil
}

//Update applies UserUpdates to the given user ID
//and returns the newly-updated user
func (s *MySQLStore) Update(id int64, updates *Updates) (*User, error) {
	updateq := "update users set firstname = ?, lastname = ? where id = ?"
	updated, err := s.db.Exec(updateq, updates.FirstName, updates.LastName, id)
	if err != nil {
		return nil, fmt.Errorf("updating: %v", err)
	}

	if err := checkRowsAffected(updated); err != nil {
		return nil, err
	}

	return s.GetByID(id)

}

//Delete deletes the user with the given ID
func (s *MySQLStore) Delete(id int64) error {
	deleteq := "delete from users where id = ?"
	deleted, err := s.db.Exec(deleteq, id)
	if err != nil {
		return fmt.Errorf("Error deleting user: %v", err)
	}

	if err = checkRowsAffected(deleted); err != nil {
		return err
	}
	return nil
}

func checkRowsAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %v", err)
	}

	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

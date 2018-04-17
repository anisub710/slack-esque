package users

import (
	"database/sql"
	"fmt"
)

//MyPostGressStore represents a users.Store backed by MySQL
type MyPostGressStore struct {
	db *sql.DB
}

//NewMyPostGressStore constructs a new MySQLStore.
func NewMyPostGressStore(db *sql.DB) *MyPostGressStore {
	return &MyPostGressStore{
		db: db,
	}
}

func (s *MyPostGressStore) getBase(param string, value interface{}) (*User, error) {
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
func (s *MyPostGressStore) GetByID(id int64) (*User, error) {

	user, err := s.getBase("id", id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//GetByEmail returns the User with the given email
func (s *MyPostGressStore) GetByEmail(email string) (*User, error) {
	user, err := s.getBase("email", email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//GetByUserName returns the User with the given Username
func (s *MyPostGressStore) GetByUserName(username string) (*User, error) {
	user, err := s.getBase("username", username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//Insert inserts the user into the database, and returns
//the newly-inserted User, complete with the DBMS-assigned ID
func (s *MyPostGressStore) Insert(user *User) (*User, error) {

	var lastInsertID int64
	insq := "insert into users(email, passhash, userName, firstName, lastName, photourl) values (?,?,?,?,?,?) returning id;"
	err := s.db.QueryRow(insq, user.Email, user.PassHash,
		user.UserName, user.FirstName, user.LastName, user.PhotoURL).Scan(&lastInsertID)

	if err != nil {
		return nil, fmt.Errorf("Error executing insert: %v", err)
	}
	user.ID = lastInsertID

	return user, nil
}

//Update applies UserUpdates to the given user ID
//and returns the newly-updated user
func (s *MyPostGressStore) Update(id int64, updates *Updates) (*User, error) {
	//check if updates are valid?
	updateq := "update users set firstname = ?, lastname = ? where id = ?;"
	_, err := s.db.Exec(updateq, updates.FirstName, updates.LastName, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return s.GetByID(id)
}

//Delete deletes the user with the given ID
func (s *MyPostGressStore) Delete(id int64) error {
	deleteq := "delete from users where id = ?"
	_, err := s.db.Exec(deleteq, id)
	if err != nil {
		return fmt.Errorf("Error deleting user: %v", err)
	}
	return nil
}

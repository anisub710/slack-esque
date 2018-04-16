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

func (s *MySQLStore) getBase(param interface{}) (*User, error) {
	query := fmt.Sprintf(`select id, email, passhash, userName, firstName, lastName, photourl from users where %v=?`, param)
	user := &User{}
	err := s.db.QueryRow(query, param).Scan(&user.ID, &user.Email, &user.PassHash,
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

	user, err := s.getBase(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//GetByEmail returns the User with the given email
func (s *MySQLStore) GetByEmail(email string) (*User, error) {
	user, err := s.getBase(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//GetByUserName returns the User with the given Username
func (s *MySQLStore) GetByUserName(username string) (*User, error) {
	user, err := s.getBase(username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//Insert inserts the user into the database, and returns
//the newly-inserted User, complete with the DBMS-assigned ID
func (s *MySQLStore) Insert(user *User) (*User, error) {
	//check if users are valid?
	insq := "insert into users(email, passhash, userName, firstName, lastName, photourl) values (?,?,?,?,?,?)"
	res, err := s.db.Exec(insq, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)

	if err != nil {
		return nil, fmt.Errorf("Error executing insert: %v", err)
	}

	_, err = res.LastInsertId()

	if err != nil {
		return nil, fmt.Errorf("Error getting last id: %v", err)
	}

	return user, nil
}

//Update applies UserUpdates to the given user ID
//and returns the newly-updated user
func (s *MySQLStore) Update(id int64, updates *Updates) (*User, error) {
	//check if updates are valid?
	updateq := "update users set firstname = ?, lastname = ? where id = ?"
	_, err := s.db.Exec(updateq, updates.FirstName, updates.LastName, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return s.GetByID(id)
}

//Delete deletes the user with the given ID
func (s *MySQLStore) Delete(id int64) error {
	deleteq := "delete from users where id = ?"
	_, err := s.db.Exec(deleteq, id)
	if err != nil {
		return fmt.Errorf("Error deleting user: %v", err)
	}
	return nil
}

// id int not null auto_increment primary key,
// email varchar(255) not null unique,
// passhash binary(60) not null,
// username varchar(255) not null unique,
// firstname varchar(35) null,
// lastname varchar(35) null,
// photourl

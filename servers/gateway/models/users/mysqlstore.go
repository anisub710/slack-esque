package users

import (
	"database/sql"
	"fmt"

	"github.com/info344-s18/challenges-ask710/servers/gateway/indexes"
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

//getBase performs all select statements
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

// func (s *MySQLStore) updateBase(param string, value interface{}) (*User, error) {
// 	updateq := "update users set "
// }

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

//UpdatePhoto updates the photourl for a user
func (s *MySQLStore) UpdatePhoto(id int64, photourl string) (*User, error) {
	updateq := "update users set photourl = ? where id = ?"
	_, err := s.db.Exec(updateq, photourl, id)
	if err != nil {
		return nil, fmt.Errorf("updating: %v", err)
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

//InsertLogin inserts login activity
func (s *MySQLStore) InsertLogin(login *Login) (*Login, error) {
	insq := "insert into userslogin(userid, logintime, ipaddr) values (?,?,?)"
	res, err := s.db.Exec(insq, login.Userid, login.LoginTime, login.IPAddr)

	if err != nil {
		return nil, fmt.Errorf("Error executing insert: %v", err)
	}

	id, err := res.LastInsertId()

	if err != nil {
		return nil, fmt.Errorf("Error getting last id: %v", err)
	}

	login.ID = id

	return login, nil
}

//checkRowsAffected returns the number of affected rows
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

//UpdatePassword updates password after resetting it.
func (s *MySQLStore) UpdatePassword(id int64, passHash []byte) (*User, error) {
	updateq := "update users set passhash = ? where id = ?"
	_, err := s.db.Exec(updateq, passHash, id)
	if err != nil {
		return nil, fmt.Errorf("Error updating: %v", err)
	}

	return s.GetByID(id)
}

//LoadUsers gets all users to add to the trie
func (s *MySQLStore) LoadUsers() (*indexes.Trie, error) {
	query := "select * from users"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Error loading users for trie: %v", err)
	}
	defer rows.Close()

	users, err := extractUserRows(rows)
	if err != nil {
		return nil, err
	}

	trie := indexes.NewTrie()
	for _, u := range *users {
		trie.AddConvertedUsers(u.FirstName, u.LastName, u.UserName, u.ID)
	}
	return trie, nil
}

//GetSearchUsers gets all users based on the found Ids
func (s *MySQLStore) GetSearchUsers(found []int64) (*[]User, error) {
	if len(found) < 1 {
		return nil, nil
	}
	query := queryForSearch(found)
	selectq := "select id, email, passhash, username, firstname, lastname, photourl from users where id in " + query
	args := makeInterface(found)
	rows, err := s.db.Query(selectq, args...)
	if err != nil {
		return nil, fmt.Errorf("Error loading users for trie: %v", err)
	}
	defer rows.Close()

	users, err := extractUserRows(rows)

	return users, err
}

//makeInterface makes the interface to be passed in to the select
//statement with the user ids.
func makeInterface(found []int64) []interface{} {
	args := make([]interface{}, len(found))
	for i, f := range found {
		args[i] = f
	}
	return args
}

//iterates through rows and returns array of users
func extractUserRows(rows *sql.Rows) (*[]User, error) {
	users := &[]User{}
	for rows.Next() {
		user := User{}
		if err := rows.Scan(&user.ID, &user.Email, &user.PassHash,
			&user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL); err != nil {
			return nil, fmt.Errorf("Error scanning users for trie: %v", err)
		}
		*users = append(*users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error getting next row: %v", err)
	}
	return users, nil
}

//queryForSearch is a function for creating (?,?..) based on the length of
//input ids for the select query
func queryForSearch(found []int64) string {
	query := "(?"
	for i := 1; i < len(found); i++ {
		query += ",?"
	}
	query += ") order by username asc"
	return query
}

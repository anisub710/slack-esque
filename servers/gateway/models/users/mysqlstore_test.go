package users

import (
	"fmt"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const sqlGet = "select id, email, passhash, username, firstname, lastname, photourl from users "
const sqlInsert = "insert into users(email, passhash, userName, firstName, lastName, photourl)"

func TestGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := &User{
		ID:        1,
		Email:     "test123@uw.edu",
		PassHash:  []byte{36, 50, 97, 36, 49, 51, 36, 66, 78, 100},
		UserName:  "competentGopher",
		FirstName: "Competent",
		LastName:  "Gopher",
		PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
	}

	rows := sqlmock.NewRows([]string{"id", "email", "passhash", "username", "firstname", "lastname", "photourl"})
	rows.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)
	mock.ExpectQuery(sqlGet).WithArgs(1).WillReturnRows(rows)

	store := NewMySQLStore(db)

	user, err := store.GetByID(1)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	fmt.Printf("%d %s", user.ID, user.FirstName)

}

func TestGetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := &User{
		ID:        1,
		Email:     "test123@uw.edu",
		PassHash:  []byte{36, 50, 97, 36, 49, 51, 36, 66, 78, 100},
		UserName:  "competentGopher",
		FirstName: "Competent",
		LastName:  "Gopher",
		PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
	}

	rows := sqlmock.NewRows([]string{"id", "email", "passhash", "username", "firstname", "lastname", "photourl"})
	rows.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)
	mock.ExpectQuery(sqlGet).WithArgs("test123@uw.edu").WillReturnRows(rows)

	store := NewMySQLStore(db)

	user, err := store.GetByEmail("test123@uw.edu")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	fmt.Printf("%d %s", user.ID, user.FirstName)

}

func TestGetByUserName(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := &User{
		ID:        1,
		Email:     "test123@uw.edu",
		PassHash:  []byte{36, 50, 97, 36, 49, 51, 36, 66, 78, 100},
		UserName:  "competentGopher",
		FirstName: "Competent",
		LastName:  "Gopher",
		PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
	}

	rows := sqlmock.NewRows([]string{"id", "email", "passhash", "username", "firstname", "lastname", "photourl"})
	rows.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)
	mock.ExpectQuery(sqlGet).WithArgs("competentGopher").WillReturnRows(rows)

	store := NewMySQLStore(db)

	user, err := store.GetByUserName("competentGopher")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	fmt.Printf("%d %s", user.ID, user.FirstName)

}

// func TestInsert(t *testing.T) {
// 	db, mock, err := sqlmock.New()

// 	if err != nil {
// 		t.Fatalf("error creating sql mock: %v", err)
// 	}

// 	defer db.Close()

// 	expectedUser := &User{
// 		ID:        1,
// 		Email:     "test123@uw.edu",
// 		PassHash:  []byte{36, 50, 97, 36, 49, 51, 36, 66, 78, 100},
// 		UserName:  "competentGopher",
// 		FirstName: "Competent",
// 		LastName:  "Gopher",
// 		PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
// 	}

// 	rows := sqlmock.NewRows([]string{"id", "email", "passhash", "username", "firstname", "lastname", "photourl"})
// 	rows.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)
// 	mock.ExpectExec("insert").WithArgs(expectedUser.Email, expectedUser.PassHash,
// 		expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName,
// 		expectedUser.PhotoURL)

// 	store := NewMySQLStore(db)

// 	user, err := store.Insert(expectedUser)

// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}

// 	fmt.Printf("%d %s", user.ID, user.FirstName)

// }

func TestInsert(t *testing.T) {
	// 	db, mock, err := sqlmock.New()

	// 	if err != nil {
	// 		t.Fatalf("error creating sql mock: %v", err)
	// 	}

	// 	defer db.Close()

	// 	expectedUser := &User{
	// 		ID:        1,
	// 		Email:     "test123@uw.edu",
	// 		PassHash:  []byte{36, 50, 97, 36, 49, 51, 36, 66, 78, 100},
	// 		UserName:  "competentGopher",
	// 		FirstName: "Competent",
	// 		LastName:  "Gopher",
	// 		PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
	// 	}

	// 	rows := sqlmock.NewRows([]string{"id", "email", "passhash", "username", "firstname", "lastname", "photourl"})
	// 	rows.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)
	// 	mock.ExpectExec("insert").WithArgs(expectedUser.Email, expectedUser.PassHash,
	// 		expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName,
	// 		expectedUser.PhotoURL)

	// 	store := NewMySQLStore(db)

	// 	user, err := store.Insert(expectedUser)

	// 	if err != nil {
	// 		t.Errorf("unexpected error: %v", err)
	// 	}

	// 	fmt.Printf("%d %s", user.ID, user.FirstName)

}

func TestUpdate(t *testing.T) {

}

func TestDelete(t *testing.T) {

}

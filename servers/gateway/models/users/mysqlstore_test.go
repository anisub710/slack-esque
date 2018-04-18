package users

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const sqlGet = "select id, email, passhash, username, firstname, lastname, photourl from users where id=?"
const sqlGetEmail = "select id, email, passhash, username, firstname, lastname, photourl from users where email=?"
const sqlGetUserName = "select id, email, passhash, username, firstname, lastname, photourl from users where username=?"
const sqlInsert = "insert into users(email, passhash, username, firstname, lastname, photourl) values (?,?,?,?,?,?)"
const sqlUpdate = "update users set firstname = ?, lastname = ? where id = ?"
const sqlDelete = "delete from users where id = ?"

func createMock() (*sql.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	return db, mock, err
}

func checkMockExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func createTestUser(userType string) *User {
	var expectedUser *User
	switch userType {
	case "normal":
		expectedUser = &User{
			ID:        1,
			Email:     "test123@uw.edu",
			PassHash:  []byte{36, 50, 97, 36, 49, 51, 36, 66, 78, 100},
			UserName:  "competentGopher",
			FirstName: "Competent",
			LastName:  "Gopher",
			PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
		}
	case "insertError":
		expectedUser = &User{
			ID:        1,
			Email:     "test123@uw.edu",
			PassHash:  nil,
			UserName:  "competentGopher",
			FirstName: "Competent",
			LastName:  "Gopher",
			PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
		}
	case "updated":
		expectedUser = &User{
			ID:        1,
			Email:     "test123@uw.edu",
			PassHash:  nil,
			UserName:  "competentGopher",
			FirstName: "Incompetent",
			LastName:  "Shark",
			PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
		}

	}

	return expectedUser
}

func createRows(expectedUser *User) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "email", "passhash", "username", "firstname", "lastname", "photourl"})
	rows.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName,
		expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)
	return rows
}

func TestGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := createTestUser("normal")

	rows := createRows(expectedUser)
	mock.ExpectQuery(regexp.QuoteMeta(sqlGet)).WithArgs(1).WillReturnRows(rows)

	store := NewMySQLStore(db)

	user, err := store.GetByID(1)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if err == nil && !reflect.DeepEqual(user, expectedUser) {
		t.Errorf("Returned user not equal to expected user")
	}

	mock.ExpectQuery(regexp.QuoteMeta(sqlGet)).WithArgs(2).WillReturnError(ErrUserNotFound)

	_, err = store.GetByID(2)

	if err == nil {
		t.Errorf("Expected Error: %v, but got nothing", ErrUserNotFound)
	}

	checkMockExpectations(t, mock)

}

func TestGetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := createTestUser("normal")
	rows := createRows(expectedUser)
	mock.ExpectQuery(regexp.QuoteMeta(sqlGetEmail)).WithArgs("test123@uw.edu").WillReturnRows(rows)

	store := NewMySQLStore(db)

	user, err := store.GetByEmail("test123@uw.edu")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if err == nil && !reflect.DeepEqual(user, expectedUser) {
		t.Errorf("Returned user not equal to expected user")
	}

	mock.ExpectQuery(regexp.QuoteMeta(sqlGetEmail)).WithArgs("random@email.com").WillReturnError(ErrUserNotFound)

	_, err = store.GetByEmail("random@email.com")

	if err == nil {
		t.Errorf("Expected Error: %v, but got nothing", ErrUserNotFound)
	}
	checkMockExpectations(t, mock)

}

func TestGetByUserName(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := createTestUser("normal")

	rows := createRows(expectedUser)
	mock.ExpectQuery(regexp.QuoteMeta(sqlGetUserName)).WithArgs("competentGopher").WillReturnRows(rows)

	store := NewMySQLStore(db)

	user, err := store.GetByUserName("competentGopher")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if err == nil && !reflect.DeepEqual(user, expectedUser) {
		t.Errorf("Returned user not equal to expected user")
	}

	mock.ExpectQuery(regexp.QuoteMeta(sqlGetUserName)).WithArgs("incompetentGopher").WillReturnError(ErrUserNotFound)

	_, err = store.GetByUserName("incompetentGopher")

	if err == nil {
		t.Errorf("Expected Error: %v, but got nothing", ErrUserNotFound)
	}
	checkMockExpectations(t, mock)
}

func TestInsert(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := createTestUser("normal")

	mock.ExpectExec(regexp.QuoteMeta(sqlInsert)).WithArgs(expectedUser.Email, expectedUser.PassHash,
		expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName,
		expectedUser.PhotoURL).WillReturnResult(sqlmock.NewResult(1, 1))

	store := NewMySQLStore(db)

	user, err := store.Insert(expectedUser)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if err == nil && !reflect.DeepEqual(user, expectedUser) {
		t.Errorf("Returned user not equal to expected user")
	}

	errorExpectedUser := createTestUser("insertError")
	expectedError := fmt.Errorf("Error executing insert")
	mock.ExpectExec(regexp.QuoteMeta(sqlInsert)).WithArgs(errorExpectedUser.Email, errorExpectedUser.PassHash,
		errorExpectedUser.UserName, errorExpectedUser.FirstName, errorExpectedUser.LastName,
		errorExpectedUser.PhotoURL).WillReturnError(expectedError)

	_, err = store.Insert(errorExpectedUser)

	if err == nil {
		t.Errorf("Expected error: %v, but got nothing", expectedError)
	}
	checkMockExpectations(t, mock)
}

func TestUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	store := NewMySQLStore(db)
	expectedUser := createTestUser("updated")

	updates := &Updates{
		FirstName: "Incompetent",
		LastName:  "Shark",
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlUpdate)).WithArgs(updates.FirstName, updates.LastName, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	rows := createRows(expectedUser)
	mock.ExpectQuery(regexp.QuoteMeta(sqlGet)).WithArgs(1).WillReturnRows(rows)

	updated, err := store.Update(1, updates)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if err == nil && !reflect.DeepEqual(updated, expectedUser) {
		t.Errorf("Returned user not equal to expected user")
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlUpdate)).WithArgs(updates.FirstName, updates.LastName, 2).WillReturnError(ErrUserNotFound)

	if _, err = store.Update(2, updates); err == nil {
		t.Errorf("Expected error: %v", ErrUserNotFound)
	}
	checkMockExpectations(t, mock)
}

func TestDelete(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	store := NewMySQLStore(db)
	if err = store.Delete(1); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	deleteErr := fmt.Errorf("Error deleting user: %v", err)
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(2).WillReturnError(deleteErr)
	if err = store.Delete(2); err == nil {
		t.Errorf("Expected error: %v, but got nothing", deleteErr)
	}
	checkMockExpectations(t, mock)
}

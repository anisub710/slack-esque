package users

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const sqlPostgresInsert = "insert into users(email, passhash, username, firstname, lastname, photourl) values (?,?,?,?,?,?) returning id;"

func TestPostGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := createTestUser("normal")

	rows := createRows(expectedUser)
	mock.ExpectQuery(regexp.QuoteMeta(sqlGet)).WithArgs(1).WillReturnRows(rows)

	store := NewMyPostGressStore(db)

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

	mock.ExpectQuery(regexp.QuoteMeta(sqlGet)).WithArgs(3).WillReturnError(sql.ErrNoRows)
	_, err = store.GetByID(3)

	if err == nil {
		t.Errorf("Expected Error: %v but got nothing", sql.ErrNoRows)
	}
	checkMockExpectations(t, mock)

}

func TestPostGetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := createTestUser("normal")
	rows := createRows(expectedUser)
	mock.ExpectQuery(regexp.QuoteMeta(sqlGetEmail)).WithArgs("test123@uw.edu").WillReturnRows(rows)

	store := NewMyPostGressStore(db)

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

	mock.ExpectQuery(regexp.QuoteMeta(sqlGetEmail)).WithArgs("anotherrand@email.com").WillReturnError(sql.ErrNoRows)
	_, err = store.GetByEmail("anotherrand@email.com")

	if err == nil {
		t.Errorf("Expected Error: %v but got nothing", sql.ErrNoRows)
	}

	checkMockExpectations(t, mock)

}

func TestPostGetByUserName(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := createTestUser("normal")

	rows := createRows(expectedUser)
	mock.ExpectQuery(regexp.QuoteMeta(sqlGetUserName)).WithArgs("competentGopher").WillReturnRows(rows)

	store := NewMyPostGressStore(db)

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

	mock.ExpectQuery(regexp.QuoteMeta(sqlGetUserName)).WithArgs("randomGopher").WillReturnError(sql.ErrNoRows)
	_, err = store.GetByUserName("randomGopher")

	if err == nil {
		t.Errorf("Expected Error: %v but got nothing", sql.ErrNoRows)
	}

	checkMockExpectations(t, mock)

}
func TestPostInsert(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	expectedUser := createTestUser("normal")

	rows := sqlmock.NewRows([]string{"id"})
	rows.AddRow(expectedUser.ID)
	mock.ExpectQuery(regexp.QuoteMeta(sqlPostgresInsert)).WithArgs(expectedUser.Email, expectedUser.PassHash,
		expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName,
		expectedUser.PhotoURL).WillReturnRows(rows)

	store := NewMyPostGressStore(db)

	user, err := store.Insert(expectedUser)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if err == nil && !reflect.DeepEqual(user, expectedUser) {
		t.Errorf("Returned user not equal to expected user")
	}

	errorExpectedUser := createTestUser("insertError")
	expectedError := fmt.Errorf("Error executing insert")
	mock.ExpectQuery(regexp.QuoteMeta(sqlPostgresInsert)).WithArgs(errorExpectedUser.Email, errorExpectedUser.PassHash,
		errorExpectedUser.UserName, errorExpectedUser.FirstName, errorExpectedUser.LastName,
		errorExpectedUser.PhotoURL).WillReturnError(expectedError)

	_, err = store.Insert(errorExpectedUser)

	if err == nil {
		t.Errorf("Expected error: %v, but got nothing", expectedError)
	}
	checkMockExpectations(t, mock)

}

func TestPostUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	store := NewMyPostGressStore(db)
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
	mock.ExpectExec(regexp.QuoteMeta(sqlUpdate)).WithArgs(updates.FirstName, updates.LastName, 3).WillReturnResult(sqlmock.NewResult(0, 0))
	if _, err = store.Update(3, updates); err == nil {
		t.Errorf("Expected error: %v", ErrUserNotFound)
	}

	_, rowsError := driver.ResultNoRows.RowsAffected()
	mock.ExpectExec(regexp.QuoteMeta(sqlUpdate)).WithArgs(updates.FirstName, updates.LastName, 3).
		WillReturnResult(sqlmock.NewErrorResult(rowsError))
	if _, err = store.Update(3, updates); err == nil {
		t.Errorf("Expected error: %v but found nothing", rowsError)
	}
	checkMockExpectations(t, mock)
}

func TestPostDelete(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	store := NewMyPostGressStore(db)
	if err = store.Delete(1); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	deleteErr := fmt.Errorf("Error deleting user: %v", err)
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(2).WillReturnError(deleteErr)
	if err = store.Delete(2); err == nil {
		t.Errorf("Expected error: %v, but got nothing", deleteErr)
	}

	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(3).WillReturnResult(sqlmock.NewResult(0, 0))
	if err = store.Delete(3); err == nil {
		t.Errorf("Expected error: %v but got nothing", ErrUserNotFound)
	}

	_, rowsError := driver.ResultNoRows.RowsAffected()
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(3).
		WillReturnResult(sqlmock.NewErrorResult(rowsError))
	if err = store.Delete(3); err == nil {
		t.Errorf("Expected error: %v but found nothing", rowsError)
	}
	checkMockExpectations(t, mock)

}

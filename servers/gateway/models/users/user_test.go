package users

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

//TODO: add tests for the various functions in user.go, as described in the assignment.
//use `go test -cover` to ensure that you are covering all or nearly all of your code paths.

func TestUserValidate(t *testing.T) {
	cases := []struct {
		name          string
		nu            *NewUser
		expectError   bool
		expectedError string
	}{
		{
			"Valid New User",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "competentGopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			false,
			"",
		},
		{
			"Invalid Email Address",
			&NewUser{
				Email:        "test123_uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "competentGopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			true,
			"Should return error for invalid email address",
		},
		{
			"Invalid Password",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "",
				PasswordConf: "",
				UserName:     "competentGopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			true,
			"Should return error for invalid password length",
		},
		{
			"Passwords don't match",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "password123",
				PasswordConf: "password1234",
				UserName:     "competentGopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			true,
			"Should return error for passwords not matching",
		},
		{
			"Invalid User Name: Empty User Name",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			true,
			"Should return error for no User Name",
		},
		{
			"Invalid User Name: User Name with spaces",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "competent Gopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			true,
			"Should return error for invalid User Name",
		},
		{
			"Invalid User Name: User Name only space",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "  ",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			true,
			"Should return error for invalid User Name",
		},
	}

	for _, c := range cases {
		err := c.nu.Validate()
		switch {
		case c.expectError && err == nil:
			t.Errorf("case %s: expected error: %s,  but did not get any ", c.name, c.expectedError)
		case !c.expectError && err != nil:
			t.Errorf("case %s: unexpected error: %v", c.name, err)
		}
	}
}

func TestToUser(t *testing.T) {
	cases := []struct {
		name          string
		nu            *NewUser
		expectError   bool
		expectedError string
	}{
		{
			"Invalid New User",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "competent Gopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			true,
			"Invalid user: user name has spaces",
		},
		{
			"Invalid Email",
			&NewUser{
				Email:        "test 123@uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "competentGopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			true,
			"Invalid user: Email is invalid",
		},
		{
			"Valid Email for Photo URL",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "competentGopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			false,
			"",
		},
		{
			"Valid Email for Photo URL: Uppercase and space and mixed case Password",
			&NewUser{
				Email:        "TeSt123@uw.edu ",
				Password:     "pAsSword123",
				PasswordConf: "pAsSword123",
				UserName:     "competentGopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			false,
			"",
		},
	}

	for _, c := range cases {
		u, err := c.nu.ToUser()
		switch {
		case c.expectError && err == nil:
			t.Errorf("case %s: expected error: %s,  but did not get any ", c.name, c.expectedError)
		case !c.expectError && err != nil:
			t.Errorf("case %s: unexpected error: %v", c.name, err)
		case !c.expectError && err == nil:
			pass := u.PassHash
			fmt.Println(pass)
			if err := bcrypt.CompareHashAndPassword(pass, []byte(c.nu.Password)); err != nil {
				t.Errorf("case %s: unexpected error while comparing hash passwords: %v", c.name, err)
			}
			cleanEmail := strings.ToLower(strings.TrimSpace(c.nu.Email))
			hasher := md5.New()
			hasher.Write([]byte(cleanEmail))
			hashEmail := hasher.Sum(nil)
			expectedPhoto := gravatarBasePhotoURL + hex.EncodeToString(hashEmail)
			if expectedPhoto != u.PhotoURL {
				t.Errorf("case %s: Photo URLs don't match: expected %s, but got %s", c.name, expectedPhoto, u.PhotoURL)
			}
		}
	}

}

func TestFullName(t *testing.T) {
	cases := []struct {
		name         string
		u            *User
		expectedName string
	}{
		{
			"Both First Name and Last Name provided",
			&User{
				FirstName: "Competent",
				LastName:  "Gopher",
			},
			"Competent Gopher",
		},
		{
			"Both First Name and Last Name not provided",
			&User{
				FirstName: "",
				LastName:  "",
			},
			"",
		},
		{
			"Only First Name provided",
			&User{
				FirstName: "Competent",
				LastName:  "",
			},
			"Competent",
		},
		{
			"Only Last Name provided",
			&User{
				FirstName: "",
				LastName:  "Gopher",
			},
			"Gopher",
		},
	}

	for _, c := range cases {
		name := c.u.FullName()
		if name != c.expectedName {
			t.Errorf("case %s: Wrong result, expected %s, but got %s", c.name, c.expectedName, name)
		}
	}
}

func TestAuthenticate(t *testing.T) {
	cases := []struct {
		name          string
		nu            *NewUser
		testPass      string
		expectError   bool
		expectedError string
	}{
		{
			"Valid Authentication",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "competentGopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			"password123",
			false,
			"",
		},
		{
			"Invalid Authentication: passed in password doesn't match hash",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "competentGopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			"differentPass",
			true,
			"Passed in password doesn't match hash",
		},
		{
			"Invalid Authentication: passed in password is empty",
			&NewUser{
				Email:        "test123@uw.edu",
				Password:     "password123",
				PasswordConf: "password123",
				UserName:     "competentGopher",
				FirstName:    "Competent",
				LastName:     "Gopher",
			},
			"",
			true,
			"Passed in password doesn't match hash",
		},
	}

	for _, c := range cases {
		u, err := c.nu.ToUser()
		if err != nil {
			t.Errorf("case %s: Unexpected error in converting new user to user: %v", c.name, err)
		}
		err = u.Authenticate(c.testPass)
		switch {
		case c.expectError && err == nil:
			t.Errorf("case %s: Expected error: %s, but didn't get any", c.name, c.expectedError)
		case !c.expectError && err != nil:
			t.Errorf("case %s: Unexpected error: %v", c.name, err)
		}
	}
}

func TestApplyUpdates(t *testing.T) {
	cases := []struct {
		name          string
		updates       *Updates
		u             *User
		expectError   bool
		expectedError string
		expectedFName string
		expectedLName string
	}{
		{
			"Valid Update: Change First and Last Names",
			&Updates{
				FirstName: "Incompetent",
				LastName:  "Goer",
			},
			&User{
				FirstName: "Competent",
				LastName:  "Gopher",
			},
			false,
			"",
			"Incompetent",
			"Goer",
		},
		{
			"Valid Update: Change only First Name",
			&Updates{
				FirstName: "Incompetent",
				LastName:  "Gopher",
			},
			&User{
				FirstName: "Competent",
				LastName:  "Gopher",
			},
			false,
			"",
			"Incompetent",
			"Gopher",
		},
		{
			"Invalid Update: No first name or last name provided",
			&Updates{
				FirstName: "",
				LastName:  "",
			},
			&User{
				FirstName: "Competent",
				LastName:  "Gopher",
			},
			true,
			"Invalid update: no first name or last name provided",
			"Competent",
			"Gopher",
		},
		{
			"Invalid Update: No last name provided",
			&Updates{
				FirstName: "Incompetent",
				LastName:  "",
			},
			&User{
				FirstName: "Competent",
				LastName:  "Gopher",
			},
			true,
			"Invalid update: no last name provided",
			"Competent",
			"Gopher",
		},
	}

	for _, c := range cases {
		err := c.u.ApplyUpdates(c.updates)

		switch {
		case !c.expectError && err != nil:
			t.Errorf("case %s: Unexpected error: %v", c.name, err)
		case c.expectError && err == nil:
			t.Errorf("case %s: Expected error: %s, but got nothing", c.name, c.expectedError)
		case c.expectedFName != c.u.FirstName || c.expectedLName != c.u.LastName:
			t.Errorf("case %s: Did not update properly: expected %s, %s, got %s, %s", c.name,
				c.expectedFName, c.expectedLName, c.u.FirstName, c.u.LastName)
		}

	}
}

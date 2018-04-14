package users

import (
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
				Password:     "pass",
				PasswordConf: "pass",
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
		if c.expectError && err == nil {
			t.Errorf("case %s: expected error: %s,  but did not get any ", c.name, c.expectedError)
		} else if !c.expectError && err != nil {
			t.Errorf("case %s: unexpected error: %v", c.name, err)
		}
	}
}

func TestToUser(t *testing.T) {
	cases := []struct {
		name string
		nu   *NewUser
		// expectedURL      string
		// expectedPassHash []byte
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
			"Correct Photo URL",
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
	}

	for _, c := range cases {
		u, err := c.nu.ToUser()
		if c.expectError && err == nil {
			t.Errorf("case %s: expected error: %s,  but did not get any ", c.name, c.expectedError)
		} else if !c.expectError && err != nil {
			t.Errorf("case %s: unexpected error: %v", c.name, err)
		} else if !c.expectError && err == nil {
			pass := u.PassHash
			if err := bcrypt.CompareHashAndPassword(pass, []byte("")); err != nil {
				t.Errorf("case %s: unexpected error while comparing hash passwords: %v", c.name, err)
			}
		}
	}

}

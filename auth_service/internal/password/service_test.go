package password_test

import (
	"errors"
	"fmt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/password"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_ValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    *string
		expectedErr error
	}{
		{
			name:        "Error when nil password is passed",
			expectedErr: fmt.Errorf("password must be at least %d symbols long", password.MinimumPasswordLength),
		},
		{
			name:        "Error when password starts with leading space",
			password:    util.GenericPtr("  random-password"),
			expectedErr: errors.New("password must not start or end with spaces"),
		},
		{
			name:        "Error when password ends with leading space",
			password:    util.GenericPtr("random-password  "),
			expectedErr: errors.New("password must not start or end with spaces"),
		},
		{
			name:        "Error when password contains only one repetitive symbol",
			password:    util.GenericPtr("aaaaaaaaaaaaaaaaaaaaaaaaa"),
			expectedErr: errors.New("password must not be one repeated character"),
		},
		{
			name:        "Error when password is less than 8 symbols",
			password:    util.GenericPtr("notsec"),
			expectedErr: fmt.Errorf("password must be minumum %d symbols long", password.MinimumPasswordLength),
		},
		{
			name:     "Success when password is valid",
			password: util.GenericPtr("M!itahag56126fu"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			passwordService := password.NewService()
			err := passwordService.ValidatePassword(test.password)
			if test.expectedErr != nil {
				require.Error(t, err)
				require.EqualError(t, test.expectedErr, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

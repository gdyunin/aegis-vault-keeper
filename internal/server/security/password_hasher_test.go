package security

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPasswordHasherVerificator(t *testing.T) {
	t.Parallel()

	hasher := func(password string) (string, error) {
		return "hashed_" + password, nil
	}
	verificator := func(hashedData, verifyingData string) (bool, error) {
		return hashedData == "hashed_"+verifyingData, nil
	}

	phv := NewPasswordHasherVerificator(hasher, verificator)

	require.NotNil(t, phv)
	assert.NotNil(t, phv.hasher)
	assert.NotNil(t, phv.verificator)
}

func TestPasswordHasherVerificator_PasswordHash(t *testing.T) {
	t.Parallel()

	type fields struct {
		hasher hashFunc
	}
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "successful_hash",
			fields: fields{
				hasher: func(password string) (string, error) {
					return "hashed_" + password, nil
				},
			},
			args: args{
				password: "test123",
			},
			want:    "hashed_test123",
			wantErr: false,
		},
		{
			name: "empty_password",
			fields: fields{
				hasher: func(password string) (string, error) {
					if password == "" {
						return "", errors.New("password cannot be empty")
					}
					return "hashed_" + password, nil
				},
			},
			args: args{
				password: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "hasher_function_error",
			fields: fields{
				hasher: func(password string) (string, error) {
					return "", errors.New("hashing failed")
				},
			},
			args: args{
				password: "test123",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "complex_password",
			fields: fields{
				hasher: func(password string) (string, error) {
					return "hashed_" + password, nil
				},
			},
			args: args{
				password: "ComplexP@ssw0rd!#$%",
			},
			want:    "hashed_ComplexP@ssw0rd!#$%",
			wantErr: false,
		},
		{
			name: "unicode_password",
			fields: fields{
				hasher: func(password string) (string, error) {
					return "hashed_" + password, nil
				},
			},
			args: args{
				password: "пароль123",
			},
			want:    "hashed_пароль123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &PasswordHasherVerificator{
				hasher: tt.fields.hasher,
			}
			got, err := p.PasswordHash(tt.args.password)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPasswordHasherVerificator_PasswordVerify(t *testing.T) {
	t.Parallel()

	type fields struct {
		verificator veriFunc
	}
	type args struct {
		hashedData    string
		verifyingData string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "successful_verification",
			fields: fields{
				verificator: func(hashedData, verifyingData string) (bool, error) {
					return hashedData == "hashed_"+verifyingData, nil
				},
			},
			args: args{
				hashedData:    "hashed_test123",
				verifyingData: "test123",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "failed_verification_wrong_password",
			fields: fields{
				verificator: func(hashedData, verifyingData string) (bool, error) {
					return hashedData == "hashed_"+verifyingData, nil
				},
			},
			args: args{
				hashedData:    "hashed_test123",
				verifyingData: "wrong_password",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "verification_function_error",
			fields: fields{
				verificator: func(hashedData, verifyingData string) (bool, error) {
					return false, errors.New("verification failed")
				},
			},
			args: args{
				hashedData:    "hashed_test123",
				verifyingData: "test123",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "empty_hash_data",
			fields: fields{
				verificator: func(hashedData, verifyingData string) (bool, error) {
					if hashedData == "" {
						return false, errors.New("hash data cannot be empty")
					}
					return hashedData == "hashed_"+verifyingData, nil
				},
			},
			args: args{
				hashedData:    "",
				verifyingData: "test123",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "empty_verifying_data",
			fields: fields{
				verificator: func(hashedData, verifyingData string) (bool, error) {
					return hashedData == "hashed_"+verifyingData, nil
				},
			},
			args: args{
				hashedData:    "hashed_",
				verifyingData: "",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "complex_password_verification",
			fields: fields{
				verificator: func(hashedData, verifyingData string) (bool, error) {
					return hashedData == "hashed_"+verifyingData, nil
				},
			},
			args: args{
				hashedData:    "hashed_ComplexP@ssw0rd!#$%",
				verifyingData: "ComplexP@ssw0rd!#$%",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "unicode_password_verification",
			fields: fields{
				verificator: func(hashedData, verifyingData string) (bool, error) {
					return hashedData == "hashed_"+verifyingData, nil
				},
			},
			args: args{
				hashedData:    "hashed_пароль123",
				verifyingData: "пароль123",
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &PasswordHasherVerificator{
				verificator: tt.fields.verificator,
			}
			got, err := p.PasswordVerify(tt.args.hashedData, tt.args.verifyingData)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPasswordHasherVerificator_Integration(t *testing.T) {
	t.Parallel()

	// Integration test that verifies hash and verify work together
	hasher := func(password string) (string, error) {
		if password == "" {
			return "", errors.New("password cannot be empty")
		}
		return "secure_hash_" + password, nil
	}

	verificator := func(hashedData, verifyingData string) (bool, error) {
		if hashedData == "" || verifyingData == "" {
			return false, errors.New("data cannot be empty")
		}
		expectedHash := "secure_hash_" + verifyingData
		return hashedData == expectedHash, nil
	}

	phv := NewPasswordHasherVerificator(hasher, verificator)

	tests := []struct {
		name       string
		password   string
		wantHash   string
		wantVerify bool
		wantErr    bool
	}{
		{
			name:       "simple_password",
			password:   "test123",
			wantHash:   "secure_hash_test123",
			wantVerify: true,
			wantErr:    false,
		},
		{
			name:       "complex_password",
			password:   "Str0ng!P@ssw0rd",
			wantHash:   "secure_hash_Str0ng!P@ssw0rd",
			wantVerify: true,
			wantErr:    false,
		},
		{
			name:     "empty_password",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test hashing
			hashedPassword, err := phv.PasswordHash(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantHash, hashedPassword)

			// Test verification with correct password
			isValid, err := phv.PasswordVerify(hashedPassword, tt.password)
			require.NoError(t, err)
			assert.Equal(t, tt.wantVerify, isValid)

			// Test verification with wrong password
			isValidWrong, err := phv.PasswordVerify(hashedPassword, "wrong_password")
			require.NoError(t, err)
			assert.False(t, isValidWrong)
		})
	}
}

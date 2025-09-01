package credential

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/credential"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/credential"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock repository for testing.
type mockRepository struct {
	saveFunc func(ctx context.Context, params repository.SaveParams) error
	loadFunc func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error)
}

func (m *mockRepository) Save(ctx context.Context, params repository.SaveParams) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, params)
	}
	return nil
}

func (m *mockRepository) Load(
	ctx context.Context,
	params repository.LoadParams,
) ([]*credential.Credential, error) {
	if m.loadFunc != nil {
		return m.loadFunc(ctx, params)
	}
	return nil, nil
}

func TestNewService(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{}
	service := NewService(repo)

	require.NotNil(t, service)
	assert.Equal(t, repo, service.r)
}

func TestService_Pull(t *testing.T) {
	t.Parallel()

	testCredID := uuid.New()
	testUserID := uuid.New()
	testTime := time.Now()

	testCred := &credential.Credential{
		ID:          testCredID,
		UserID:      testUserID,
		Login:       []byte("testuser"),
		Password:    []byte("testpass123"),
		Description: []byte("Test credential"),
		UpdatedAt:   testTime,
	}

	type args struct {
		params PullParams
	}
	tests := []struct {
		setupMock      func(*mockRepository)
		name           string
		expectedErrMsg string
		args           args
		wantErr        bool
		expectCred     bool
	}{
		{
			name: "successful_pull",
			args: args{
				params: PullParams{
					ID:     testCredID,
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return []*credential.Credential{testCred}, nil
				}
			},
			expectCred: true,
			wantErr:    false,
		},
		{
			name: "credential_not_found",
			args: args{
				params: PullParams{
					ID:     testCredID,
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return []*credential.Credential{}, nil
				}
			},
			wantErr:        true,
			expectedErrMsg: "credential not found",
		},
		{
			name: "repository_load_failed",
			args: args{
				params: PullParams{
					ID:     testCredID,
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to load credentials",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			service := NewService(repo)
			cred, err := service.Pull(context.Background(), tt.args.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Nil(t, cred)
			} else {
				require.NoError(t, err)
				if tt.expectCred {
					require.NotNil(t, cred)
					assert.Equal(t, testCredID, cred.ID)
					assert.Equal(t, testUserID, cred.UserID)
					assert.Equal(t, "testuser", cred.Login)
					assert.Equal(t, "testpass123", cred.Password)
				}
			}
		})
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testTime := time.Now()

	testCreds := []*credential.Credential{
		{
			ID:          uuid.New(),
			UserID:      testUserID,
			Login:       []byte("user1"),
			Password:    []byte("pass1"),
			Description: []byte("Credential 1"),
			UpdatedAt:   testTime,
		},
		{
			ID:          uuid.New(),
			UserID:      testUserID,
			Login:       []byte("user2"),
			Password:    []byte("pass2"),
			Description: []byte("Credential 2"),
			UpdatedAt:   testTime,
		},
	}

	type args struct {
		params ListParams
	}
	tests := []struct {
		setupMock      func(*mockRepository)
		name           string
		expectedErrMsg string
		expectedCount  int
		args           args
		wantErr        bool
	}{
		{
			name: "successful_list",
			args: args{
				params: ListParams{
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return testCreds, nil
				}
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "empty_list",
			args: args{
				params: ListParams{
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return []*credential.Credential{}, nil
				}
			},
			expectedCount: 0,
			wantErr:       false,
		},
		{
			name: "repository_load_failed",
			args: args{
				params: ListParams{
					UserID: testUserID,
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to load credentials",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			service := NewService(repo)
			creds, err := service.List(context.Background(), tt.args.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Nil(t, creds)
			} else {
				require.NoError(t, err)
				assert.Len(t, creds, tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, testUserID, creds[0].UserID)
				}
			}
		})
	}
}

func TestService_Push(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testCredID := uuid.New()

	type args struct {
		params *PushParams
	}
	tests := []struct {
		args           args
		setupMock      func(*mockRepository)
		name           string
		expectedErrMsg string
		wantErr        bool
		expectID       bool
	}{
		{
			name: "successful_create",
			args: args{
				params: &PushParams{
					UserID:      testUserID,
					Login:       "testuser",
					Password:    "testpass123",
					Description: "Test credential",
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.saveFunc = func(ctx context.Context, params repository.SaveParams) error {
					return nil
				}
			},
			expectID: true,
			wantErr:  false,
		},
		{
			name: "successful_update",
			args: args{
				params: &PushParams{
					ID:          testCredID,
					UserID:      testUserID,
					Login:       "updated_user",
					Password:    "updated_pass",
					Description: "Updated credential",
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return []*credential.Credential{
						{
							ID:     testCredID,
							UserID: testUserID,
						},
					}, nil
				}
				repo.saveFunc = func(ctx context.Context, params repository.SaveParams) error {
					return nil
				}
			},
			expectID: true,
			wantErr:  false,
		},
		{
			name: "invalid_credential_data",
			args: args{
				params: &PushParams{
					UserID:      testUserID,
					Login:       "", // Invalid empty login
					Password:    "testpass123",
					Description: "Test credential",
				},
			},
			wantErr:        true,
			expectedErrMsg: "failed to create credential",
		},
		{
			name: "repository_save_failed",
			args: args{
				params: &PushParams{
					UserID:      testUserID,
					Login:       "testuser",
					Password:    "testpass123",
					Description: "Test credential",
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.saveFunc = func(ctx context.Context, params repository.SaveParams) error {
					return errors.New("database error")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to save credential",
		},
		{
			name: "update_access_denied",
			args: args{
				params: &PushParams{
					ID:          testCredID,
					UserID:      testUserID,
					Login:       "testuser",
					Password:    "testpass123",
					Description: "Test credential",
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return []*credential.Credential{}, nil // No credential found
				}
			},
			wantErr:        true,
			expectedErrMsg: "access check for updating credential failed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			service := NewService(repo)
			credID, err := service.Push(context.Background(), tt.args.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Equal(t, uuid.Nil, credID)
			} else {
				require.NoError(t, err)
				if tt.expectID {
					assert.NotEqual(t, uuid.Nil, credID)
				}
			}
		})
	}
}

func TestService_checkAccessToUpdate(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testCredID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		setupMock      func(*mockRepository)
		name           string
		expectedErrMsg string
		credID         uuid.UUID
		userID         uuid.UUID
		wantErr        bool
	}{
		{
			name:   "access_granted",
			credID: testCredID,
			userID: testUserID,
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return []*credential.Credential{
						{
							ID:     testCredID,
							UserID: testUserID,
						},
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:   "credential_not_found",
			credID: testCredID,
			userID: testUserID,
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return []*credential.Credential{}, nil
				}
			},
			wantErr:        true,
			expectedErrMsg: "credential for update not found",
		},
		{
			name:   "access_denied_different_user",
			credID: testCredID,
			userID: testUserID,
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return []*credential.Credential{
						{
							ID:     testCredID,
							UserID: otherUserID, // Different user
						},
					}, nil
				}
			},
			wantErr:        true,
			expectedErrMsg: "access denied to credential",
		},
		{
			name:   "repository_error",
			credID: testCredID,
			userID: testUserID,
			setupMock: func(repo *mockRepository) {
				repo.loadFunc = func(ctx context.Context, params repository.LoadParams) ([]*credential.Credential, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr:        true,
			expectedErrMsg: "failed to pull existing credential",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			service := NewService(repo)
			err := service.checkAccessToUpdate(context.Background(), tt.credID, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

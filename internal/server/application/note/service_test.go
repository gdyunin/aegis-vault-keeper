package note

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/note"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/note"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository implements Repository interface for testing.
type MockRepository struct {
	SaveFunc func(ctx context.Context, params repository.SaveParams) error
	LoadFunc func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error)
}

func (m *MockRepository) Save(ctx context.Context, params repository.SaveParams) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, params)
	}
	return nil
}

func (m *MockRepository) Load(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc(ctx, params)
	}
	return nil, nil
}

func TestNewService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		repo Repository
		want *Service
		name string
	}{
		{
			name: "success/creates_service_with_repository",
			repo: &MockRepository{},
			want: &Service{r: &MockRepository{}},
		},
		{
			name: "success/nil_repository",
			repo: nil,
			want: &Service{r: nil},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewService(tt.repo)
			require.NotNil(t, got)
			assert.Equal(t, tt.repo, got.r)
		})
	}
}

func TestService_Pull(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testNoteID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		setupMock   func(*MockRepository)
		want        *Note
		name        string
		wantErrText string
		params      PullParams
		wantErr     bool
	}{
		{
			name: "success/note_found",
			params: PullParams{
				ID:     testNoteID,
				UserID: testUserID,
			},
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					assert.Equal(t, testNoteID, params.ID)
					assert.Equal(t, testUserID, params.UserID)
					return []*note.Note{{
						ID:          testNoteID,
						UserID:      testUserID,
						Note:        []byte("test note"),
						Description: []byte("test description"),
						UpdatedAt:   testTime,
					}}, nil
				}
			},
			want: &Note{
				ID:          testNoteID,
				UserID:      testUserID,
				Note:        "test note",
				Description: "test description",
				UpdatedAt:   testTime,
			},
			wantErr: false,
		},
		{
			name: "error/note_not_found",
			params: PullParams{
				ID:     testNoteID,
				UserID: testUserID,
			},
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					return []*note.Note{}, nil
				}
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "note not found",
		},
		{
			name: "error/repository_error",
			params: PullParams{
				ID:     testNoteID,
				UserID: testUserID,
			},
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					return nil, errors.New("database error")
				}
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "failed to load notes",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &MockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewService(mockRepo)
			got, err := service.Pull(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		setupMock   func(*MockRepository)
		name        string
		wantErrText string
		want        []*Note
		params      ListParams
		wantErr     bool
	}{
		{
			name: "success/notes_found",
			params: ListParams{
				UserID: testUserID,
			},
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					assert.Equal(t, testUserID, params.UserID)
					return []*note.Note{
						{
							ID:          uuid.New(),
							UserID:      testUserID,
							Note:        []byte("note 1"),
							Description: []byte("desc 1"),
							UpdatedAt:   testTime,
						},
						{
							ID:          uuid.New(),
							UserID:      testUserID,
							Note:        []byte("note 2"),
							Description: []byte("desc 2"),
							UpdatedAt:   testTime,
						},
					}, nil
				}
			},
			want: []*Note{
				{
					ID:          uuid.UUID{},
					UserID:      testUserID,
					Note:        "note 1",
					Description: "desc 1",
					UpdatedAt:   testTime,
				},
				{
					ID:          uuid.UUID{},
					UserID:      testUserID,
					Note:        "note 2",
					Description: "desc 2",
					UpdatedAt:   testTime,
				},
			},
			wantErr: false,
		},
		{
			name: "success/empty_list",
			params: ListParams{
				UserID: testUserID,
			},
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					return []*note.Note{}, nil
				}
			},
			want:    []*Note{},
			wantErr: false,
		},
		{
			name: "error/repository_error",
			params: ListParams{
				UserID: testUserID,
			},
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					return nil, errors.New("database error")
				}
			},
			want:        nil,
			wantErr:     true,
			wantErrText: "failed to load notes",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &MockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewService(mockRepo)
			got, err := service.List(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				// Compare length first
				require.Len(t, got, len(tt.want))
				// Compare individual elements (excluding generated IDs)
				for i := range got {
					assert.Equal(t, tt.want[i].UserID, got[i].UserID)
					assert.Equal(t, tt.want[i].Note, got[i].Note)
					assert.Equal(t, tt.want[i].Description, got[i].Description)
					assert.Equal(t, tt.want[i].UpdatedAt, got[i].UpdatedAt)
				}
			}
		})
	}
}

func TestService_Push(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testNoteID := uuid.New()

	tests := []struct {
		params      *PushParams
		setupMock   func(*MockRepository)
		name        string
		wantErrText string
		wantID      uuid.UUID
		wantErr     bool
	}{
		{
			name: "success/create_new_note",
			params: &PushParams{
				UserID:      testUserID,
				Note:        "test note",
				Description: "test description",
			},
			setupMock: func(m *MockRepository) {
				m.SaveFunc = func(ctx context.Context, params repository.SaveParams) error {
					assert.NotNil(t, params.Entity)
					assert.Equal(t, testUserID, params.Entity.UserID)
					assert.Equal(t, []byte("test note"), params.Entity.Note)
					assert.Equal(t, []byte("test description"), params.Entity.Description)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "success/update_existing_note",
			params: &PushParams{
				ID:          testNoteID,
				UserID:      testUserID,
				Note:        "updated note",
				Description: "updated description",
			},
			setupMock: func(m *MockRepository) {
				loadCallCount := 0
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					loadCallCount++
					if loadCallCount == 1 {
						// First call from checkAccessToUpdate
						assert.Equal(t, testNoteID, params.ID)
						assert.Equal(t, testUserID, params.UserID)
						return []*note.Note{{
							ID:     testNoteID,
							UserID: testUserID,
							Note:   []byte("existing note"),
						}}, nil
					}
					return nil, nil
				}
				m.SaveFunc = func(ctx context.Context, params repository.SaveParams) error {
					assert.Equal(t, testNoteID, params.Entity.ID)
					assert.Equal(t, testUserID, params.Entity.UserID)
					return nil
				}
			},
			wantID:  testNoteID,
			wantErr: false,
		},
		{
			name: "error/invalid_note_text",
			params: &PushParams{
				UserID:      testUserID,
				Note:        "", // Invalid empty note
				Description: "test description",
			},
			setupMock:   func(m *MockRepository) {},
			wantErr:     true,
			wantErrText: "failed to create new note",
		},
		{
			name: "error/access_denied_update",
			params: &PushParams{
				ID:          testNoteID,
				UserID:      testUserID,
				Note:        "updated note",
				Description: "updated description",
			},
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					return []*note.Note{{
						ID:     testNoteID,
						UserID: uuid.New(), // Different user
						Note:   []byte("existing note"),
					}}, nil
				}
			},
			wantErr:     true,
			wantErrText: "access denied",
		},
		{
			name: "error/update_note_not_found",
			params: &PushParams{
				ID:          testNoteID,
				UserID:      testUserID,
				Note:        "updated note",
				Description: "updated description",
			},
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					return []*note.Note{}, nil // Note not found
				}
			},
			wantErr:     true,
			wantErrText: "note not found",
		},
		{
			name: "error/save_repository_error",
			params: &PushParams{
				UserID:      testUserID,
				Note:        "test note",
				Description: "test description",
			},
			setupMock: func(m *MockRepository) {
				m.SaveFunc = func(ctx context.Context, params repository.SaveParams) error {
					return errors.New("database error")
				}
			},
			wantErr:     true,
			wantErrText: "failed to save note",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &MockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewService(mockRepo)
			gotID, err := service.Push(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
				assert.Equal(t, uuid.Nil, gotID)
			} else {
				require.NoError(t, err)
				if tt.wantID != uuid.Nil {
					assert.Equal(t, tt.wantID, gotID)
				} else {
					assert.NotEqual(t, uuid.Nil, gotID)
				}
			}
		})
	}
}

func TestService_checkAccessToUpdate(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testNoteID := uuid.New()
	differentUserID := uuid.New()

	tests := []struct {
		setupMock   func(*MockRepository)
		name        string
		wantErrText string
		noteID      uuid.UUID
		userID      uuid.UUID
		wantErr     bool
	}{
		{
			name:   "success/access_granted",
			noteID: testNoteID,
			userID: testUserID,
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					return []*note.Note{{
						ID:     testNoteID,
						UserID: testUserID,
						Note:   []byte("test note"),
					}}, nil
				}
			},
			wantErr: false,
		},
		{
			name:   "error/note_not_found",
			noteID: testNoteID,
			userID: testUserID,
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					return []*note.Note{}, nil
				}
			},
			wantErr:     true,
			wantErrText: "note for update not found",
		},
		{
			name:   "error/access_denied",
			noteID: testNoteID,
			userID: testUserID,
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					return []*note.Note{{
						ID:     testNoteID,
						UserID: differentUserID, // Different user
						Note:   []byte("test note"),
					}}, nil
				}
			},
			wantErr:     true,
			wantErrText: "access denied to note",
		},
		{
			name:   "error/repository_error",
			noteID: testNoteID,
			userID: testUserID,
			setupMock: func(m *MockRepository) {
				m.LoadFunc = func(ctx context.Context, params repository.LoadParams) ([]*note.Note, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr:     true,
			wantErrText: "failed to pull existing note",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &MockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewService(mockRepo)
			err := service.checkAccessToUpdate(context.Background(), tt.noteID, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

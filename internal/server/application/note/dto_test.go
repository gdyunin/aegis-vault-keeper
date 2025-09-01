package note

import (
	"testing"
	"time"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/note"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoteFromDomain(t *testing.T) {
	t.Parallel()

	testTime := time.Now()
	testUserID := uuid.New()
	testNoteID := uuid.New()

	tests := []struct {
		input *note.Note
		want  *Note
		name  string
	}{
		{
			name: "success/valid_domain_note",
			input: &note.Note{
				ID:          testNoteID,
				UserID:      testUserID,
				Note:        []byte("test note content"),
				Description: []byte("test description"),
				UpdatedAt:   testTime,
			},
			want: &Note{
				ID:          testNoteID,
				UserID:      testUserID,
				Note:        "test note content",
				Description: "test description",
				UpdatedAt:   testTime,
			},
		},
		{
			name: "success/empty_note_content",
			input: &note.Note{
				ID:          testNoteID,
				UserID:      testUserID,
				Note:        []byte(""),
				Description: []byte(""),
				UpdatedAt:   testTime,
			},
			want: &Note{
				ID:          testNoteID,
				UserID:      testUserID,
				Note:        "",
				Description: "",
				UpdatedAt:   testTime,
			},
		},
		{
			name:  "nil_input/returns_nil",
			input: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newNoteFromDomain(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewNotesFromDomain(t *testing.T) {
	t.Parallel()

	testTime := time.Now()
	testUserID := uuid.New()
	testNoteID1 := uuid.New()
	testNoteID2 := uuid.New()

	tests := []struct {
		name  string
		input []*note.Note
		want  []*Note
	}{
		{
			name: "success/multiple_notes",
			input: []*note.Note{
				{
					ID:          testNoteID1,
					UserID:      testUserID,
					Note:        []byte("note 1"),
					Description: []byte("desc 1"),
					UpdatedAt:   testTime,
				},
				{
					ID:          testNoteID2,
					UserID:      testUserID,
					Note:        []byte("note 2"),
					Description: []byte("desc 2"),
					UpdatedAt:   testTime,
				},
			},
			want: []*Note{
				{
					ID:          testNoteID1,
					UserID:      testUserID,
					Note:        "note 1",
					Description: "desc 1",
					UpdatedAt:   testTime,
				},
				{
					ID:          testNoteID2,
					UserID:      testUserID,
					Note:        "note 2",
					Description: "desc 2",
					UpdatedAt:   testTime,
				},
			},
		},
		{
			name:  "success/empty_slice",
			input: []*note.Note{},
			want:  []*Note{},
		},
		{
			name:  "success/nil_slice",
			input: nil,
			want:  []*Note{},
		},
		{
			name: "success/slice_with_nil_element",
			input: []*note.Note{
				{
					ID:          testNoteID1,
					UserID:      testUserID,
					Note:        []byte("note 1"),
					Description: []byte("desc 1"),
					UpdatedAt:   testTime,
				},
				nil,
			},
			want: []*Note{
				{
					ID:          testNoteID1,
					UserID:      testUserID,
					Note:        "note 1",
					Description: "desc 1",
					UpdatedAt:   testTime,
				},
				nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newNotesFromDomain(tt.input)
			require.Len(t, got, len(tt.want))

			for i := range got {
				assert.Equal(t, tt.want[i], got[i])
			}
		})
	}
}

func TestPullParams(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testNoteID := uuid.New()

	tests := []struct {
		name       string
		params     PullParams
		wantID     uuid.UUID
		wantUserID uuid.UUID
	}{
		{
			name: "success/valid_params",
			params: PullParams{
				ID:     testNoteID,
				UserID: testUserID,
			},
			wantID:     testNoteID,
			wantUserID: testUserID,
		},
		{
			name: "success/nil_uuids",
			params: PullParams{
				ID:     uuid.Nil,
				UserID: uuid.Nil,
			},
			wantID:     uuid.Nil,
			wantUserID: uuid.Nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantID, tt.params.ID)
			assert.Equal(t, tt.wantUserID, tt.params.UserID)
		})
	}
}

func TestListParams(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()

	tests := []struct {
		name       string
		params     ListParams
		wantUserID uuid.UUID
	}{
		{
			name: "success/valid_params",
			params: ListParams{
				UserID: testUserID,
			},
			wantUserID: testUserID,
		},
		{
			name: "success/nil_uuid",
			params: ListParams{
				UserID: uuid.Nil,
			},
			wantUserID: uuid.Nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantUserID, tt.params.UserID)
		})
	}
}

func TestPushParams(t *testing.T) {
	t.Parallel()

	testUserID := uuid.New()
	testNoteID := uuid.New()

	tests := []struct {
		name            string
		wantNote        string
		wantDescription string
		params          PushParams
		wantID          uuid.UUID
		wantUserID      uuid.UUID
	}{
		{
			name: "success/valid_params",
			params: PushParams{
				Note:        "test note",
				Description: "test description",
				ID:          testNoteID,
				UserID:      testUserID,
			},
			wantNote:        "test note",
			wantDescription: "test description",
			wantID:          testNoteID,
			wantUserID:      testUserID,
		},
		{
			name: "success/empty_strings_nil_uuids",
			params: PushParams{
				Note:        "",
				Description: "",
				ID:          uuid.Nil,
				UserID:      uuid.Nil,
			},
			wantNote:        "",
			wantDescription: "",
			wantID:          uuid.Nil,
			wantUserID:      uuid.Nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantNote, tt.params.Note)
			assert.Equal(t, tt.wantDescription, tt.params.Description)
			assert.Equal(t, tt.wantID, tt.params.ID)
			assert.Equal(t, tt.wantUserID, tt.params.UserID)
		})
	}
}

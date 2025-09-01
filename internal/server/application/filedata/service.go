package filedata

import (
	"context"
	"errors"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/domain/filedata"
	repository "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/filedata"
	filestorage "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/filestorage"
	"github.com/google/uuid"
)

// Repository defines the interface for file metadata persistence operations.
type Repository interface {
	// Save persists file metadata using the provided parameters.
	Save(ctx context.Context, params repository.SaveParams) error
	// Load retrieves file metadata using the provided parameters.
	Load(ctx context.Context, params repository.LoadParams) ([]*filedata.FileData, error)
}

// FileStorageRepository defines the interface for actual file content storage operations.
type FileStorageRepository interface {
	// Save stores file content using the provided parameters.
	Save(ctx context.Context, params filestorage.SaveParams) error
	// Load retrieves file content using the provided parameters.
	Load(ctx context.Context, params filestorage.LoadParams) ([]byte, error)
	// Delete removes file content using the provided parameters.
	Delete(ctx context.Context, params filestorage.DeleteParams) error
}

// Service provides file data management business logic operations.
type Service struct {
	// r handles file metadata persistence.
	r Repository
	// fs handles actual file content storage.
	fs FileStorageRepository
}

// NewService creates a new file data service with the provided repositories.
func NewService(r Repository, fs FileStorageRepository) *Service {
	return &Service{r: r, fs: fs}
}

// Pull retrieves a specific file's metadata and content by ID.
func (s *Service) Pull(ctx context.Context, params PullParams) (*FileData, error) {
	fd, err := s.loadMetadata(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	fileData, err := s.fs.Load(ctx, filestorage.LoadParams{
		UserID:     fd.UserID,
		StorageKey: string(fd.StorageKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load file data: %w", mapError(err))
	}

	result := newFileFromDomain(fd)
	result.Data = fileData

	return result, nil
}

// List retrieves all files belonging to the specified user.
func (s *Service) List(ctx context.Context, params ListParams) ([]*FileData, error) {
	fds, err := s.r.Load(ctx, repository.LoadParams{
		UserID: params.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load files: %w", mapError(err))
	}
	return newFilesFromDomain(fds), nil
}

// Push creates or updates a file for the specified user with validation and encryption.
func (s *Service) Push(ctx context.Context, params *PushParams) (uuid.UUID, error) {
	if len(params.Data) == 0 {
		return uuid.Nil, fmt.Errorf("file data is required: %w", ErrFileDataRequired)
	}

	fd, err := filedata.NewFile(filedata.NewFileDataParams{
		UserID:      params.UserID,
		StorageKey:  params.StorageKey,
		HashSum:     params.calculateDataHashSum(),
		Description: params.Description,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create file: %w", mapError(err))
	}

	if params.ID != uuid.Nil {
		existing, err := s.findFileForUpdate(ctx, params)
		if err != nil {
			return uuid.Nil, fmt.Errorf("update file access error: %w", err)
		}
		fd.ID = params.ID
		if err := s.removeOldFileOnKeyChange(ctx, existing, params.StorageKey); err != nil {
			return uuid.Nil, fmt.Errorf("old file delete error: %w", err)
		}
	}

	if err := s.fs.Save(ctx, filestorage.SaveParams{
		UserID:     fd.UserID,
		StorageKey: string(fd.StorageKey),
		Data:       params.Data,
	}); err != nil {
		return uuid.Nil, fmt.Errorf("failed to save file data: %w", mapError(err))
	}

	if err := s.r.Save(ctx, repository.SaveParams{Entity: fd}); err != nil {
		if rollbackErr := s.rollbackFileSave(ctx, fd); rollbackErr != nil {
			return uuid.Nil, errors.Join(
				fmt.Errorf("failed to save file metadata: %w", mapError(err)),
				rollbackErr,
			)
		}
		return uuid.Nil, fmt.Errorf("failed to save file metadata: %w", mapError(err))
	}

	return fd.ID, nil
}

// findFileForUpdate retrieves and validates access to an existing file for update operations.
func (s *Service) findFileForUpdate(ctx context.Context, params *PushParams) (*filedata.FileData, error) {
	existing, err := s.loadMetadata(ctx, PullParams{ID: params.ID, UserID: params.UserID})
	if err != nil {
		return nil, fmt.Errorf("access check for updating file failed: %w", err)
	}
	if existing.UserID != params.UserID {
		return nil, fmt.Errorf("access denied to file: %w", ErrFileAccessDenied)
	}
	return existing, nil
}

// removeOldFileOnKeyChange deletes the old file from storage when storage key changes during update.
func (s *Service) removeOldFileOnKeyChange(
	ctx context.Context,
	existing *filedata.FileData,
	newStorageKey string,
) error {
	if string(existing.StorageKey) != newStorageKey {
		if err := s.fs.Delete(ctx, filestorage.DeleteParams{
			UserID:     existing.UserID,
			StorageKey: string(existing.StorageKey),
		}); err != nil {
			return fmt.Errorf("failed to delete old file data: %w", mapError(err))
		}
	}
	return nil
}

// rollbackFileSave removes a saved file from storage during transaction rollback operations.
func (s *Service) rollbackFileSave(ctx context.Context, fd *filedata.FileData) error {
	if deleteErr := s.fs.Delete(ctx, filestorage.DeleteParams{
		UserID:     fd.UserID,
		StorageKey: string(fd.StorageKey),
	}); deleteErr != nil {
		return errors.Join(ErrRollBackFileSaveFailed, deleteErr)
	}
	return nil
}

// loadMetadata retrieves file metadata for the specified file and user without loading file content.
func (s *Service) loadMetadata(ctx context.Context, params PullParams) (*filedata.FileData, error) {
	fds, err := s.r.Load(ctx, repository.LoadParams{
		ID:     params.ID,
		UserID: params.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load file metadata: %w", mapError(err))
	}
	if len(fds) == 0 {
		return nil, fmt.Errorf("file not found: %w", ErrFileNotFound)
	}

	return fds[0], nil
}

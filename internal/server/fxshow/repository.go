package fxshow

import (
	"context"

	applicationAuth "github.com/gdyunin/aegis-vault-keeper/internal/server/application/auth"
	applicationBankcard "github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	applicationCredential "github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	applicationFiledata "github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	applicationNote "github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/config"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/database"
	repositoryAuth "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/auth"
	repositoryBankcard "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/bankcard"
	repositoryCredential "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/credential"
	repositoryDB "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/db"
	repositoryFiledata "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/filedata"
	repositoryFilestorage "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/filestorage"
	repositoryKeyprv "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/keyprv"
	repositoryNote "github.com/gdyunin/aegis-vault-keeper/internal/server/repository/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/security"
	"go.uber.org/fx"
)

// repositoryModule provides all repository layer dependencies.
// Configures database client, storage repositories, and their interface implementations.
var repositoryModule = fx.Module("repository",
	provideWithInterfaces[*repositoryAuth.Repository](
		func(dbClient repositoryDB.DBClient, cfg *config.AuthConfig) *repositoryAuth.Repository {
			return repositoryAuth.NewRepository(dbClient, cfg.MasterKey)
		},
		new(applicationAuth.Repository),
		new(security.UserKeyRepository),
	),
	provideWithInterfaces[*security.UserKeyProvider](
		security.NewUserKeyProvider,
		new(repositoryKeyprv.UserKeyProvider),
	),
	provideWithInterfaces[*repositoryBankcard.Repository](
		repositoryBankcard.NewRepository,
		new(applicationBankcard.Repository),
	),
	provideWithInterfaces[*repositoryCredential.Repository](
		repositoryCredential.NewRepository,
		new(applicationCredential.Repository),
	),
	provideWithInterfaces[*repositoryNote.Repository](
		repositoryNote.NewRepository,
		new(applicationNote.Repository),
	),
	provideWithInterfaces[*repositoryFiledata.Repository](
		repositoryFiledata.NewRepository,
		new(applicationFiledata.Repository),
	),
	provideWithInterfaces[*repositoryFilestorage.Repository](
		func(cfg *config.FileStorageConfig, kprv repositoryKeyprv.UserKeyProvider) *repositoryFilestorage.Repository {
			return repositoryFilestorage.NewRepository(cfg.BasePath, kprv)
		},
		new(applicationFiledata.FileStorageRepository),
	),
	provideWithInterfaces[*database.Client](
		func(cfg *config.DBConfig) (*database.Client, error) {
			return database.NewClient(&database.Config{
				Host:     cfg.Host,
				User:     cfg.User,
				Password: cfg.Password,
				DBName:   cfg.DBName,
				SSLMode:  cfg.SSLMode,
				Port:     cfg.Port,
				Timeout:  cfg.Timeout,
			})
		},
		new(repositoryDB.DBClient),
		new(PingCloser),
	),
)

// PingCloser interface for database clients that support connectivity testing and graceful shutdown.
type PingCloser interface {
	Ping(context.Context) error
	Close(context.Context) error
}

// runDatabaseClient registers database client lifecycle hooks with fx.
func runDatabaseClient(lc fx.Lifecycle, pc PingCloser) {
	lc.Append(fx.Hook{
		OnStart: pc.Ping,
		OnStop:  pc.Close,
	})
}

package fxshow

import (
	authApp "github.com/gdyunin/aegis-vault-keeper/internal/server/application/auth"
	bankcardApp "github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	credentialApp "github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	datasyncApp "github.com/gdyunin/aegis-vault-keeper/internal/server/application/datasync"
	filedataApp "github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	noteApp "github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/config"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/crypto"
	authDelivery "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/auth"
	bankcardDelivery "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/bankcard"
	credentialDelivery "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/credential"
	datasyncDelivery "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/datasync"
	filedataDelivery "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/filedata"
	middlewareDelivery "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/middleware"
	noteDelivery "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/security"
	"go.uber.org/fx"
)

// applicationModule provides all application layer dependencies.
// Configures security components, business logic services, and their interfaces.
var applicationModule = fx.Module("application",
	provideWithInterfaces[*security.PasswordHasherVerificator](
		func() *security.PasswordHasherVerificator {
			return security.NewPasswordHasherVerificator(crypto.HashBcrypt, crypto.VerifyBcrypt)
		},
		new(authApp.PasswordHasherVerificator),
	),
	provideWithInterfaces[*security.CryptoKeyGenerator](
		security.NewCryptoKeyGenerator,
		new(authApp.CryptoKeyGenerator),
	),
	provideWithInterfaces[*security.TokenGenerateValidator](
		func(cfg *config.AuthConfig) (*security.TokenGenerateValidator, error) {
			return security.NewTokenGenerateValidator(cfg.MasterKey, cfg.AccessTokenLifeTime)
		},
		new(authApp.TokenGenerateValidator),
	),
	provideWithInterfaces[*bankcardApp.Service](
		bankcardApp.NewService,
		new(datasyncApp.BankCardService),
		new(bankcardDelivery.Service),
	),
	provideWithInterfaces[*credentialApp.Service](
		credentialApp.NewService,
		new(datasyncApp.CredentialService),
		new(credentialDelivery.Service),
	),
	provideWithInterfaces[*noteApp.Service](
		noteApp.NewService,
		new(datasyncApp.NoteService),
		new(noteDelivery.Service),
	),
	provideWithInterfaces[*filedataApp.Service](
		filedataApp.NewService,
		new(datasyncApp.FileDataService),
		new(filedataDelivery.Service),
	),
	provideWithInterfaces[*authApp.Service](
		authApp.NewService,
		new(authDelivery.Service),
		new(middlewareDelivery.AuthWithJWTService),
	),
	provideWithInterfaces[*datasyncApp.Service](
		datasyncApp.NewService,
		new(datasyncDelivery.Service),
	),
	fx.Provide(datasyncApp.NewServicesAggregator),
)

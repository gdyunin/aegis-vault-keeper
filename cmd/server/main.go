package main

import (
	_ "github.com/gdyunin/aegis-vault-keeper/docs"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/fxshow"
)

// main provides the entry point for the AegisVaultKeeper server application.
//
// @title                       AegisVaultKeeper API
// @version                     0.1.1
// @description                 AegisVaultKeeper is a secure personal data storage service that allows users to store
// @description                 and manage sensitive information including credentials, bank cards, notes, and files.
// @termsOfService              http://swagger.io/terms/
//
// @contact.name                German Dyunin
// @contact.url                 https://github.com/gdyunin/aegis_vault_keeper
// @contact.email               gdyunin@gmail.com
//
// @license.name                MIT
// @license.url                 https://opensource.org/licenses/MIT
//
// @host                        localhost:56789
// @BasePath                    /api
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Bearer token authentication. Use 'Bearer {token}' format.
//
// @tag.name                    Auth
// @tag.description             Authentication operations - user registration and login
//
// @tag.name                    BankCards
// @tag.description             Bank card management operations - store and retrieve bank card information
//
// @tag.name                    Credentials
// @tag.description             Credential management operations - store and retrieve login/password pairs
//
// @tag.name                    Notes
// @tag.description             Note management operations - store and retrieve text notes
//
// @tag.name                    Files
// @tag.description             File storage operations - upload and download files
//
// @tag.name                    DataSync
// @tag.description             Data synchronization operations - bulk push and pull data
//
// @tag.name                    System
// @tag.description             System operations - health check and application information
// .
func main() {
	app := fxshow.BuildApp()
	app.Run()
}

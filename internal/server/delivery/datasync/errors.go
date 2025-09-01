package datasync

import (
	bankcarddel "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/bankcard"
	credentialdel "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	filedatadel "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/filedata"
	notedel "github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/note"
	"github.com/gin-gonic/gin"
)

// DataSyncErrRegistry aggregates error registries from all data types for unified error handling.
var DataSyncErrRegistry = errutil.Merge(
	bankcarddel.BankCardErrRegistry,
	credentialdel.CredentialErrRegistry,
	notedel.NoteErrRegistry,
	filedatadel.FileDataErrRegistry,
)

// handleError processes errors using the consolidated data sync error registry.
func handleError(err error, c *gin.Context) (int, []string) {
	return errutil.HandleWithRegistry(DataSyncErrRegistry, err, c)
}

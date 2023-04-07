package pgstorage

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra-provider-sdk/storage/database_storage/postgresql_storage"
	"github.com/selefra/selefra/cli_ui"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/oci"
	"github.com/songzhibin97/gkit/tools/pointer"
	"strings"
	"sync"
)

type Option func(pgopts *postgresql_storage.PostgresqlStorageOptions)

//func DefaultPgStorageOpts() *postgresql_storage.PostgresqlStorageOptions {
//	dsn := getDsn()
//
//	pgopts := postgresql_storage.NewPostgresqlStorageOptions(dsn)
//
//	return pgopts
//}

func WithSearchPath(searchPath string) Option {
	return func(pgopts *postgresql_storage.PostgresqlStorageOptions) {
		pgopts.SearchPath = searchPath
	}
}

//func PgStorageWithMeta(ctx context.Context, meta *schema.ClientMeta, opts ...Option) (*postgresql_storage.PostgresqlStorage, error) {
//	pgopts := DefaultPgStorageOpts()
//
//	for _, opt := range opts {
//		opt(pgopts)
//	}
//
//	storage, diag := postgresql_storage.NewPostgresqlStorage(ctx, pgopts)
//	if diag != nil {
//		if diag != nil {
//			err := cli_ui.PrintDiagnostic(diag.GetDiagnosticSlice())
//			if err != nil {
//				return nil, errors.New(`The database maybe not ready.
//		You can execute the following command to install the official database image.
//		docker run --name selefra_postgres -p 5432:5432 -e POSTGRES_PASSWORD=pass -d postgres\n`)
//			}
//		}
//	}
//
//	storage.SetClientMeta(meta)
//
//	return storage, nil
//}
//
//func PgStorage(ctx context.Context, opts ...Option) (*postgresql_storage.PostgresqlStorage, *schema.Diagnostics) {
//	pgopts := DefaultPgStorageOpts()
//
//	for _, opt := range opts {
//		opt(pgopts)
//	}
//
//	return postgresql_storage.NewPostgresqlStorage(ctx, pgopts)
//}
//
//func Storage(ctx context.Context, opts ...Option) (storage.Storage, *schema.Diagnostics) {
//	pgopts := DefaultPgStorageOpts()
//
//	for _, opt := range opts {
//		opt(pgopts)
//	}
//
//	return storage_factory.NewStorage(ctx, storage_factory.StorageTypePostgresql, pgopts)
//}

func GetStorageValue(ctx context.Context, storage *postgresql_storage.PostgresqlStorage, key string) (string, error) {
	v, diag := storage.GetValue(ctx, key)
	if diag != nil {
		err := cli_ui.PrintDiagnostic(diag.GetDiagnosticSlice())
		if err != nil {
			return "", err
		}
	}
	return v, nil
}

func SetStorageValue(ctx context.Context, storage *postgresql_storage.PostgresqlStorage, key, value string) error {
	if diag := storage.SetKey(ctx, key, value); diag != nil {
		err := cli_ui.PrintDiagnostic(diag.GetDiagnosticSlice())
		if err != nil {
			return err
		}
	}

	return nil
}

// ------------------------------------------------- --------------------------------------------------------------------

var runOCIPostgreSQLOnce sync.Once

// DefaultPostgreSQL
// 1. If the default Postgresql is not installed, install it
// 2. If the default Postgresql is not started, start it
// 3. Return the default Postgresql DSN connection
func DefaultPostgreSQL(downloadWorkspace string, messageChannel *message.Channel[*schema.Diagnostics]) string {

	defer func() {
		messageChannel.SenderWaitAndClose()
	}()

	isRunSuccess := true

	//runOCIPostgreSQLOnce.Do(func() {
	//	messageChannel.Send(schema.NewDiagnostics().AddInfo("Use built-in PostgreSQL database..."))
	//	downloader := oci.NewPostgreSQLDownloader(&oci.PostgreSQLDownloaderOptions{
	//		MessageChannel:    messageChannel.MakeChildChannel(),
	//		DownloadDirectory: downloadWorkspace,
	//	})
	//	isRunSuccess = downloader.Run(context.Background())
	//})

	logo := " _____        _         __              \n/  ___|      | |       / _|             \n\\ `--.   ___ | |  ___ | |_  _ __   __ _ \n `--. \\ / _ \\| | / _ \\|  _|| '__| / _` |\n/\\__/ /|  __/| ||  __/| |  | |   | (_| |\n\\____/  \\___||_| \\___||_|  |_|    \\__,_|\n"
	cli_ui.Infof(logo)
	messageChannel.Send(schema.NewDiagnostics().AddInfo("Use built-in PostgreSQL database..."))
	downloader := oci.NewPostgreSQLDownloader(&oci.PostgreSQLDownloaderOptions{
		MessageChannel:    messageChannel.MakeChildChannel(),
		DownloadDirectory: downloadWorkspace,
	})
	isRunSuccess = downloader.Run(context.Background())

	// If the built-in Postgresql does not start successfully, a prompt is returned asking what to do next
	if !isRunSuccess {
		errorMsg := `

Sorry, the built-in Postgresql fails to start, please configure your own Postgresql connection
export SELEFRA_DATABASE_DSN='host=127.0.0.1 user=postgres password=pass port=15432 dbname=postgres sslmode=disable'

If you do not already have Postgresql installed, You can start an instance of Postgresql using Docker:
sudo docker run -d --name selefra-postgres -p 15432:5432 -e POSTGRES_PASSWORD=pass postgres:14

Or you can download and install Postgresql from its official website: 
https://www.postgresql.org/download/

You can check out our documentation: https://www.selefra.io/docs/faq#how-to-use-postgresql

`
		messageChannel.Send(schema.NewDiagnostics().AddErrorMsg(errorMsg))
		return ""
	}

	db := &module.ConnectionBlock{
		Type:     "postgres",
		Username: "postgres",
		Password: oci.DefaultPostgreSQLPasswd,
		Host:     "localhost",
		Port:     pointer.ToUint64Pointer(oci.DefaultPostgreSQLPort),
		Database: "postgres",
		SSLMode:  "disable",
		Extras:   nil,
	}
	return db.BuildDSN()
}

// ------------------------------------------------- --------------------------------------------------------------------

// GetSchemaKey return provider schema named <required.name>_<required_version>_<provider_name>
func GetSchemaKey(providerName, providerVersion string, providerConfigurationBlock *module.ProviderBlock) string {
	sourceArr := strings.Split(providerName, "/")
	var source string
	if len(sourceArr) > 1 {
		source = strings.Replace(sourceArr[1]+"@"+providerVersion, "/", "_", -1)
	} else {
		source = strings.Replace(sourceArr[0]+"@"+providerVersion, "/", "_", -1)
	}
	source = strings.Replace(source, "@", "_", -1)
	source = strings.Replace(source, ".", "", -1)
	if providerConfigurationBlock != nil {
		source = source + "_" + providerConfigurationBlock.Name
	}
	return source
}

// ------------------------------------------------- --------------------------------------------------------------------

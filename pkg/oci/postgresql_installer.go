package oci

import (
	"bufio"
	"context"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/global"
	"github.com/selefra/selefra/pkg/message"
	"github.com/selefra/selefra/pkg/utils"
	"io"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ------------------------------------------------- --------------------------------------------------------------------

const (

	// DefaultPostgreSQLPasswd The default password of the startup instance, Maybe I should use a stronger random password?
	DefaultPostgreSQLPasswd = "pass"

	DefaultPostgreSQLPort = 15432
)

// ------------------------------------------------ ---------------------------------------------------------------------

//type ProgressTracker interface {
//
//	// Begin Ready for installation
//	Begin()
//
//	// InstallBegin The system starts to install postgresql
//	InstallBegin(ctx context.Context, postgresqlInstallDirectory string, d *schema.Diagnostics)
//
//	// InstallEnd Installing postgresql ends
//	InstallEnd(ctx context.Context, postgresqlInstallDirectory string, d *schema.Diagnostics)
//
//	// RunCommand Execute the command
//	RunCommand(command string, args ...string)
//
//	// Start a postgresql instance
//	Start(stdout, stderr string, diagnostics *schema.Diagnostics)
//
//	// End of installation
//	End(isSuccess bool)
//}

// ------------------------------------------------ ---------------------------------------------------------------------

// PostgreSQLDownloaderOptions Download option
type PostgreSQLDownloaderOptions struct {

	// Which directory to store it in after downloading
	DownloadDirectory string

	//// Used to receive notifications when downloading progress updates to track progress
	//ProgressTracker ProgressTracker

	MessageChannel *message.Channel[*schema.Diagnostics]
}

// ------------------------------------------------ ---------------------------------------------------------------------

type PostgreSQLInstaller struct {
	options *PostgreSQLDownloaderOptions
}

func NewPostgreSQLDownloader(options *PostgreSQLDownloaderOptions) *PostgreSQLInstaller {
	return &PostgreSQLInstaller{
		options: options,
	}
}

func (x *PostgreSQLInstaller) Run(ctx context.Context) bool {

	defer func() {
		x.options.MessageChannel.SenderWaitAndClose()
	}()

	// Make sure that postgresql exists locally. If not, install one
	if !x.IsInstalled() {
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Postgresql is not installed, it will automatically install..."))
		if !x.Install(ctx) {
			return false
		}
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Postgresql is installed successfully"))
	}

	_ = x.Stop()

	return x.Start()
}

//func loadBar(doneFlag *bool) {
//	go func() {
//		dotLen := 0
//		for *doneFlag {
//			time.Sleep(1 * time.Second)
//			if *doneFlag {
//				dotLen++
//				cli_ui.Infof("\rWaiting for DB to download %s", strings.Repeat(".", dotLen%6)+strings.Repeat(" ", 6-dotLen%6))
//			}
//		}
//	}()
//}

func (x *PostgreSQLInstaller) DownloadOCIImage(ctx context.Context) bool {

	// postgresql oci file installation directory
	imageDownloadURL := global.PkgBasePath + runtime.GOOS + global.PkgTag

	// ensure install directory exists
	postgresqlDirectory := x.buildPgInstallDirectoryPath()
	_ = os.MkdirAll(postgresqlDirectory, 0755)

	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Download postgresql oci image from %s to %s ...", imageDownloadURL, postgresqlDirectory))

	fileStore := content.NewFile(postgresqlDirectory)
	dockerResolver := docker.NewResolver(docker.ResolverOptions{})
	_, err := oras.Copy(ctx, dockerResolver, imageDownloadURL, fileStore, postgresqlDirectory)
	if err != nil {
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg("OCI install postgresql failed, download OCI image error: %s", err.Error()))
		return false
	}

	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Download postgresql OCI image success"))

	return true
}

// IsInstalled Check whether postgresql is installed
func (x *PostgreSQLInstaller) IsInstalled() bool {
	// If the executable exists, it is considered installed
	return utils.ExistsFile(x.buildPgCtlExecutePath())
}

// Install postgresql locally
func (x *PostgreSQLInstaller) Install(ctx context.Context) bool {

	if x.IsInstalled() {
		return true
	}

	if !x.DownloadOCIImage(ctx) {
		return false
	}

	x.options.MessageChannel.Send(schema.NewDiagnostics().AddInfo("Begin init postgresql..."))

	diagnostics := schema.NewDiagnostics()
	_ = utils.EnsureDirectoryExists(x.buildDataDirectory())
	// for debug
	//fmt.Println("path: " + x.buildDataDirectory())
	stdout, stderr, err := utils.RunCommand(x.buildInitExecutePath(),
		"-D", x.buildDataDirectory(),
		"-U", "postgres",
		"-E", "UTF-8",
		"--locale", "en_US.UTF-8")
	if err != nil {
		diagnostics.AddErrorMsg("Init postgres failed: %s", err.Error())
	} else {
		diagnostics.AddInfo("Init postgres success")
	}
	diagnostics = x.fixDiagnostics(diagnostics)
	if stdout != "" {
		diagnostics.AddInfo(stdout)
	}
	if stderr != "" {
		diagnostics.AddErrorMsg(stderr)
	}

	diagnostics.AddDiagnostics(x.ChangeConfigFilePort(15432))
	diagnostics = x.fixDiagnostics(diagnostics)

	x.options.MessageChannel.Send(diagnostics)

	return utils.NotHasError(diagnostics)
}

// Start the postgresql database
func (x *PostgreSQLInstaller) Start() bool {
	diagnostics := schema.NewDiagnostics()
	stdout, stderr, err := utils.RunCommand(x.buildPgCtlExecutePath(), "-D", x.buildDataDirectory(), "-l", x.buildPgLogFilePath(), "start")
	if err != nil {
		diagnostics.AddErrorMsg("Start postgresql error: %s", err.Error())
	} else {
		//diagnostics.AddInfo("Start postgresql success")
	}
	if stdout != "" {
		//diagnostics.AddInfo(stdout)
	}
	if stderr != "" {
		diagnostics.AddErrorMsg(stderr)
	}
	x.options.MessageChannel.Send(diagnostics)
	return utils.NotHasError(diagnostics)
}

func (x *PostgreSQLInstaller) Stop() bool {
	diagnostics := schema.NewDiagnostics()
	stdout, stderr, err := utils.RunCommand(x.buildPgCtlExecutePath(), "-D", x.buildDataDirectory(), "stop")
	if err != nil {
		diagnostics.AddErrorMsg("Stop postgresql error: %s", err.Error())
	} else {
		diagnostics.AddInfo("Stop postgresql success")
	}
	if stderr != "" {
		diagnostics.AddErrorMsg(stderr)
	}
	if stdout != "" {
		diagnostics.AddInfo(stdout)
	}
	return utils.HasError(diagnostics)
}

// may be
// [ error ]
// WARNING: enabling "trust" authentication for local connections
// You can change this by editing pg_hba.conf or using the option -A, or
// --auth-local and --auth-host, the next time you run initdb.
func (x *PostgreSQLInstaller) fixDiagnostics(diagnostics *schema.Diagnostics) *schema.Diagnostics {
	if diagnostics == nil {
		return nil
	}
	newDiagnostics := schema.NewDiagnostics()
	// WARNING
	for _, d := range diagnostics.GetDiagnosticSlice() {
		level := d.Level()
		content := strings.TrimSpace(d.Content())
		if strings.HasPrefix(content, "WARNING:") {
			level = schema.DiagnosisLevelWarn
		}
		newDiagnostics.AddDiagnostic(schema.NewDiagnostic(level, content))
	}
	return newDiagnostics
}

// ------------------------------------------------ ---------------------------------------------------------------------

// get the postgresql installation directory
func (x *PostgreSQLInstaller) buildPgInstallDirectoryPath() string {
	return filepath.Join(x.options.DownloadDirectory, "oci/postgresql")
}

// postgresql data storage path
func (x *PostgreSQLInstaller) buildDataDirectory() string {
	return filepath.Join(x.buildPgInstallDirectoryPath(), "pgsql/data")
}

// get the location of the initdb exec file path
func (x *PostgreSQLInstaller) buildInitExecutePath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(x.buildPgInstallDirectoryPath(), "pgsql/bin/initdb.exe")
	} else {
		return filepath.Join(x.buildPgInstallDirectoryPath(), "pgsql/bin/initdb")
	}
}

// get the execution path of the postgresql ctl file
func (x *PostgreSQLInstaller) buildPgCtlExecutePath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(x.buildPgInstallDirectoryPath(), "pgsql/bin/pg_ctl.exe")
	} else {
		return filepath.Join(x.buildPgInstallDirectoryPath(), "pgsql/bin/pg_ctl")
	}
}

// get the postgresql data location
func (x *PostgreSQLInstaller) buildPgConfigFilePath() string {
	return filepath.Join(x.buildPgInstallDirectoryPath(), "pgsql/data/postgresql.conf")
}

// get the location where postgresql logs are stored
func (x *PostgreSQLInstaller) buildPgLogFilePath() string {
	return filepath.Join(x.buildPgInstallDirectoryPath(), "pgsql/logfile")
}

// ------------------------------------------------ ---------------------------------------------------------------------

// TODO Parameter is not used
// ChangeConfigFilePort Change the port number in the configuration file
func (x *PostgreSQLInstaller) ChangeConfigFilePort(port int) *schema.Diagnostics {

	// read config file
	diagnostics := schema.NewDiagnostics()
	file, err := os.OpenFile(x.buildPgConfigFilePath(), os.O_RDWR, 0666)
	if err != nil {
		return diagnostics.AddErrorMsg("Run postgresql error, open config file %s error: %s", x.buildPgConfigFilePath(), err.Error())
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	pos := int64(0)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return diagnostics.AddErrorMsg("OCI run postgresql error, open config file %s error: %s", x.buildPgConfigFilePath(), err.Error())
			}
		}
		if strings.Contains(line, "#port = 5432") {
			defaultPort := "15432"
			portBytes := []byte("port = " + defaultPort)
			_, err := file.WriteAt(portBytes, pos)
			if err != nil {
				return diagnostics.AddErrorMsg("OCI run postgresql error, change config file %s error: %s", x.buildPgConfigFilePath(), err.Error())
			}
		}
		pos += int64(len(line))
	}
	return nil
}

// ------------------------------------------------ ---------------------------------------------------------------------

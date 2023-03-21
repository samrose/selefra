package global

import (
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

type Option func(variable *Variable)

// Variable store some global variable
type Variable struct {
	readOnlyVariable

	mux sync.RWMutex

	// token is not empty when user is login
	token string

	// orgName is selefra cloud organization name
	orgName string

	// stage is the build stage for current project
	stage string

	// projectName is local project name
	projectName string

	// relvPrjName is the name of selefra cloud project name which is relevant to local project
	relvPrjName string

	logLevel string

	server string
}

// readOnlyVariable will only be set when programmer started
type readOnlyVariable struct {
	once sync.Once

	// workspace store where selefra worked
	workspace string

	// cmd store what command is running
	cmd string
}

var g = Variable{
	readOnlyVariable: readOnlyVariable{
		once: sync.Once{},
	},
	mux: sync.RWMutex{},
}

func WithWorkspace(workspace string) Option {
	return func(variable *Variable) {
		variable.workspace = workspace
	}
}

// Init the global variables with cmd and some options
func Init(cmd string, opts ...Option) {
	g.once.Do(func() {
		g.cmd = cmd

		cwd, err := os.Getwd()
		if err != nil {
			os.Exit(1)
		}

		g.workspace = cwd

		for _, opt := range opts {
			opt(&g)
		}

	})
}

// WrappedInit wrapper the Init function to a cobra func
func WrappedInit(workspace string) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		Init(parentCmdNames(cmd), WithWorkspace(workspace))
	}
}

// DefaultWrappedInit is a cobra func that will use default value to init Variable
func DefaultWrappedInit() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		Init(parentCmdNames(cmd))
	}
}

// parentCmdNames find cmd's parent cmd name and join their name
func parentCmdNames(cmd *cobra.Command) string {
	names := make([]string, 0)
	var fn func(cmd *cobra.Command)
	fn = func(cmd *cobra.Command) {
		if cmd.Parent() != nil {
			fn(cmd.Parent())
		}

		names = append(names, cmd.Name())
	}

	fn(cmd)

	return strings.Join(names, " ")
}

func SetToken(token string) {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.token = token
}

func SetStage(stage string) {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.stage = stage
}

func SetOrgName(orgName string) {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.orgName = orgName
}

func SetProjectName(prjName string) {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.projectName = prjName
}

func ProjectName() string {
	g.mux.RLock()
	defer g.mux.RUnlock()

	return g.projectName
}

func SetRelvPrjName(name string) {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.relvPrjName = name
}

func RelvPrjName() string {
	g.mux.RLock()
	defer g.mux.RUnlock()

	return g.relvPrjName
}

func SetLogLevel(level string) {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.logLevel = level
}

func WorkSpace() string {
	return g.workspace
}

func Token() string {
	g.mux.RLock()
	defer g.mux.RUnlock()

	return g.token
}

func OrgName() string {
	g.mux.RLock()
	defer g.mux.RUnlock()

	return g.orgName
}

func Cmd() string {
	return g.cmd
}

func Stage() string {
	g.mux.RLock()
	defer g.mux.RUnlock()

	return g.stage
}

func LogLevel() string {
	g.mux.RLock()
	defer g.mux.RUnlock()

	return g.logLevel
}

const PkgBasePath = "ghcr.io/selefra/postgre_"
const PkgTag = ":latest"

var SERVER = "main-api.selefra.io"

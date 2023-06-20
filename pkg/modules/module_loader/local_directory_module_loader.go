package module_loader

import (
	"context"
	"errors"
	"fmt"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/modules/module"
	"github.com/selefra/selefra/pkg/modules/parser"
	"github.com/selefra/selefra/pkg/registry"
	"github.com/selefra/selefra/pkg/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ------------------------------------------------- --------------------------------------------------------------------

// LocalDirectoryModuleLoaderOptions Option when loading modules from a local directory
type LocalDirectoryModuleLoaderOptions struct {
	*ModuleLoaderOptions
	Instruction map[string]interface{} `json:"instruction" yaml:"instruction"`
	// Directory where the module resides Directory
	ModuleDirectory string `json:"module-directory" yaml:"module-directory"`
}

//func (x *LocalDirectoryModuleLoaderOptions) Copy() *LocalDirectoryModuleLoaderOptions {
//	return &LocalDirectoryModuleLoaderOptions{
//		ModuleLoaderOptions: x.ModuleLoaderOptions.Copy(),
//		ModuleDirectory:     x.ModuleDirectory,
//	}
//}

//func (x *LocalDirectoryModuleLoaderOptions) CopyForModuleDirectory(source, moduleDirectory string) *LocalDirectoryModuleLoaderOptions {
//	options := x.Copy()
//	options.Source = source
//	options.ModuleDirectory = moduleDirectory
//	options.DependenciesTree = append(options.DependenciesTree, source)
//	return options
//}

// BuildFullName Gets the globally unique identity of the module
func (x *LocalDirectoryModuleLoaderOptions) BuildFullName() string {
	if x.Source == "" {
		return x.ModuleDirectory
	} else {
		return fmt.Sprintf("%s @ %s", x.Source, x.ModuleDirectory)
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

// LocalDirectoryModuleLoader Load the module from the local directory
type LocalDirectoryModuleLoader struct {
	options *LocalDirectoryModuleLoaderOptions
}

var _ ModuleLoader[*LocalDirectoryModuleLoaderOptions] = &LocalDirectoryModuleLoader{}

func NewLocalDirectoryModuleLoader(options *LocalDirectoryModuleLoaderOptions) (*LocalDirectoryModuleLoader, error) {

	// convert to abs path
	options.ModuleDirectory = utils.AbsPath(options.ModuleDirectory)

	if !utils.ExistsDirectory(options.ModuleDirectory) {
		return nil, fmt.Errorf("module %s does not exist or is not directory", options.BuildFullName())
	}

	return &LocalDirectoryModuleLoader{
		options: options,
	}, nil
}

func (x *LocalDirectoryModuleLoader) Name() ModuleLoaderType {
	return ModuleLoaderTypeLocalDirectory
}

func (x *LocalDirectoryModuleLoader) Load(ctx context.Context) (*module.Module, bool) {
	defer func() {
		x.options.MessageChannel.SenderWaitAndClose()
	}()

	// check path
	d := x.checkModuleDirectory()
	x.options.MessageChannel.Send(d)
	if utils.HasError(d) {
		return nil, false
	}

	// list all yaml file
	yamlFilePathSlice, d := x.listModuleDirectoryYamlFilePath()
	x.options.MessageChannel.Send(d)
	if utils.HasError(d) {
		return nil, false
	}

	// Read all files under the module as modules, these modules may be incomplete, may be some fragments of the module
	yamlFileModuleSlice := make([]*module.Module, len(yamlFilePathSlice))
	isHasError := false
	for index, yamlFilePath := range yamlFilePathSlice {
		yamlFileModule, d := parser.NewYamlFileToModuleParser(yamlFilePath, x.options.Instruction).Parse()
		x.options.MessageChannel.Send(d)
		if utils.HasError(d) {
			isHasError = true
		}
		yamlFileModuleSlice[index] = yamlFileModule
	}
	if isHasError {
		return nil, false
	}

	// Merge these modules
	finalModule := module.NewModule()
	hasError := false
	for _, yamlFileModule := range yamlFileModuleSlice {
		merge, d := finalModule.Merge(yamlFileModule)
		x.options.MessageChannel.Send(d)
		if utils.HasError(d) {
			hasError = true
		}
		if merge != nil {
			finalModule = merge
		}
	}
	if hasError {
		return nil, false
	}

	// load sub modules
	subModuleSlice, loadSuccess := x.loadSubModules(ctx, finalModule.ModulesBlock)
	if !loadSuccess {
		return nil, false
	}
	for _, subModule := range subModuleSlice {
		subModule.ProvidersBlock = finalModule.ProvidersBlock
		subModule.SelefraBlock = finalModule.SelefraBlock
		subModule.ParentModule = finalModule
		//subModule.VariablesBlock = finalModule.VariablesBlock
	}
	finalModule.SubModules = subModuleSlice
	finalModule.Source = x.options.Source
	finalModule.ModuleLocalDirectory = x.options.ModuleDirectory
	finalModule.DependenciesPath = x.options.DependenciesTree

	return finalModule, true
}

func (x *LocalDirectoryModuleLoader) loadSubModules(ctx context.Context, modulesBlock module.ModulesBlock) ([]*module.Module, bool) {
	subModuleSlice := make([]*module.Module, 0)
	for _, moduleBlock := range modulesBlock {
		for index, useModuleSource := range moduleBlock.Uses {
			useLocation := moduleBlock.GetNodeLocation(fmt.Sprintf("uses[%d]%s", index, module.NodeLocationSelfValue))
			//moduleDirectoryPath := filepath.Dir(useLocation.Path)

			switch NewModuleLoaderBySource(useModuleSource) {

			// Unsupported loading mode
			case ModuleLoaderTypeInvalid:
				errorReport := module.RenderErrorTemplate(fmt.Sprintf("invalid module uses source %s, unsupported module loader", useModuleSource), useLocation)
				x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg(errorReport))
				return nil, false

			// Load the module from the bucket in S3
			case ModuleLoaderTypeS3Bucket:
				subModule, ok := x.loadS3BucketModule(ctx, useLocation, useModuleSource)
				if !ok {
					return nil, false
				}
				subModuleSlice = append(subModuleSlice, subModule)

			case ModuleLoaderTypeGitHubRegistry:
				subModule, ok := x.loadGitHubRegistryModule(ctx, useLocation, useModuleSource)
				if !ok {
					return nil, false
				}
				subModuleSlice = append(subModuleSlice, subModule)

			case ModuleLoaderTypeLocalDirectory:
				subModule, ok := x.loadLocalDirectoryModule(ctx, useLocation, useModuleSource)
				if !ok {
					return nil, false
				}
				subModuleSlice = append(subModuleSlice, subModule)

			case ModuleLoaderTypeURL:
				subModule, ok := x.loadURLModule(ctx, useLocation, useModuleSource)
				if !ok {
					return nil, false
				}
				subModuleSlice = append(subModuleSlice, subModule)

			default:
				errorReport := module.RenderErrorTemplate(fmt.Sprintf("module source %s can cannot be assign loader", useModuleSource), useLocation)
				x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg(errorReport))
				return nil, false
			}
		}
	}
	return subModuleSlice, true
}

func (x *LocalDirectoryModuleLoader) loadURLModule(ctx context.Context, useLocation *module.NodeLocation, useModuleSource string) (*module.Module, bool) {
	urlModuleLoaderOptions := &URLModuleLoaderOptions{
		ModuleLoaderOptions: &ModuleLoaderOptions{
			Source: useModuleSource,
			// TODO
			Version:           "",
			DownloadDirectory: x.options.DownloadDirectory,
			// TODO
			ProgressTracker:  x.options.ProgressTracker,
			MessageChannel:   x.options.MessageChannel.MakeChildChannel(),
			DependenciesTree: x.options.DeepDependenciesTree(useModuleSource),
		},
		ModuleURL: useModuleSource,
	}
	loader, err := NewURLModuleLoader(urlModuleLoaderOptions)
	if err != nil {
		urlModuleLoaderOptions.MessageChannel.SenderWaitAndClose()
		errorReport := module.RenderErrorTemplate(fmt.Sprintf("create url module %s error: %s", useModuleSource, err.Error()), useLocation)
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg(errorReport))
		return nil, false
	}
	return loader.Load(ctx)
}

func (x *LocalDirectoryModuleLoader) loadLocalDirectoryModule(ctx context.Context, useLocation *module.NodeLocation, useModuleSource string) (*module.Module, bool) {

	// The path of the submodule should be from the current path
	subModuleDirectory := filepath.Join(utils.AbsPath(x.options.ModuleDirectory), useModuleSource)

	subModuleLocalDirectoryOptions := &LocalDirectoryModuleLoaderOptions{
		Instruction: x.options.Instruction,
		ModuleLoaderOptions: &ModuleLoaderOptions{
			Source: useModuleSource,
			// TODO
			Version:           "",
			DownloadDirectory: x.options.DownloadDirectory,
			// TODO
			ProgressTracker:  x.options.ProgressTracker,
			MessageChannel:   x.options.MessageChannel.MakeChildChannel(),
			DependenciesTree: x.options.DeepDependenciesTree(useModuleSource),
		},
		ModuleDirectory: subModuleDirectory,
	}

	loader, err := NewLocalDirectoryModuleLoader(subModuleLocalDirectoryOptions)
	if err != nil {
		subModuleLocalDirectoryOptions.MessageChannel.SenderWaitAndClose()
		errorReport := module.RenderErrorTemplate(fmt.Sprintf("create local directory module %s error: %s", subModuleLocalDirectoryOptions.BuildFullName(), err.Error()), useLocation)
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg(errorReport))
		return nil, false
	}
	return loader.Load(ctx)
}

func (x *LocalDirectoryModuleLoader) loadGitHubRegistryModule(ctx context.Context, useLocation *module.NodeLocation, useModuleSource string) (*module.Module, bool) {

	githubOptions := &GitHubRegistryModuleLoaderOptions{
		ModuleLoaderOptions: &ModuleLoaderOptions{
			Source: useModuleSource,
			// TODO
			Version: "",
			// TODO
			//ProgressTracker:   x.ProgressTracker,
			DownloadDirectory: x.options.DownloadDirectory,
			MessageChannel:    x.options.MessageChannel.MakeChildChannel(),
			DependenciesTree:  x.options.DeepDependenciesTree(useModuleSource),
		},
		RegistryRepoFullName: registry.ModuleGithubRegistryDefaultRepoFullName,
	}

	loader, err := NewGitHubRegistryModuleLoader(githubOptions)
	if err != nil {
		githubOptions.MessageChannel.SenderWaitAndClose()
		errorReport := module.RenderErrorTemplate(fmt.Sprintf("create github registry module %s error: %s", githubOptions.Source, err.Error()), useLocation)
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg(errorReport))
		return nil, false
	}

	return loader.Load(ctx)
}

func (x *LocalDirectoryModuleLoader) loadS3BucketModule(ctx context.Context, useLocation *module.NodeLocation, useModuleSource string) (*module.Module, bool) {

	s3Options := &S3BucketModuleLoaderOptions{
		ModuleLoaderOptions: &ModuleLoaderOptions{
			Source: useModuleSource,
			// TODO
			Version: "",
			// TODO
			//ProgressTracker:   x.ProgressTracker,
			DownloadDirectory: x.options.DownloadDirectory,
			MessageChannel:    x.options.MessageChannel.MakeChildChannel(),
			DependenciesTree:  x.options.DeepDependenciesTree(useModuleSource),
		},
		S3BucketURL: useModuleSource,
	}

	loader, err := NewS3BucketModuleLoader(s3Options)

	if err != nil {
		s3Options.MessageChannel.SenderWaitAndClose()
		errorReport := module.RenderErrorTemplate(fmt.Sprintf("create s3 module loader %s error: %s", s3Options.Source, err.Error()), useLocation)
		x.options.MessageChannel.Send(schema.NewDiagnostics().AddErrorMsg(errorReport))
		return nil, false
	}

	return loader.Load(ctx)
}

func (x *LocalDirectoryModuleLoader) Options() *LocalDirectoryModuleLoaderOptions {
	return x.options
}

// Check that the given module local path is correct
func (x *LocalDirectoryModuleLoader) checkModuleDirectory() *schema.Diagnostics {
	info, err := os.Stat(x.options.ModuleDirectory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return schema.NewDiagnostics().AddErrorMsg("module %s not found", x.options.BuildFullName())
		} else {
			return schema.NewDiagnostics().AddErrorMsg("module %s load error: %s", x.options.BuildFullName(), err.Error())
		}
	}

	if !info.IsDir() {
		return schema.NewDiagnostics().AddErrorMsg("module %s found, but not is directory", x.options.BuildFullName())
	}

	return nil
}

// Lists all yaml files in the directory where the module resides
func (x *LocalDirectoryModuleLoader) listModuleDirectoryYamlFilePath() ([]string, *schema.Diagnostics) {
	dir, err := os.ReadDir(x.options.ModuleDirectory)
	if err != nil {
		return nil, schema.NewDiagnostics().AddErrorMsg("module %s read error: %s", x.options.BuildFullName(), err.Error())
	}
	yamlFileSlice := make([]string, 0)
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}
		if IsYamlFile(entry) {
			yamlFilePath := filepath.Join(utils.AbsPath(x.options.ModuleDirectory), entry.Name())
			yamlFileSlice = append(yamlFileSlice, yamlFilePath)
		}
	}
	return yamlFileSlice, nil
}

func IsYamlFile(entry os.DirEntry) bool {
	if entry.IsDir() {
		return false
	}
	ext := strings.ToLower(path.Ext(entry.Name()))
	return strings.HasSuffix(ext, ".yaml") || strings.HasSuffix(ext, ".yml")
}

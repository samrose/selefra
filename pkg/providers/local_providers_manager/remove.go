package local_providers_manager

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"github.com/selefra/selefra/pkg/utils"
	"github.com/selefra/selefra/pkg/version"
	"os"
)

// RemoveProviders Delete the locally installed provider by name and version. If no version is specified, all versions of the provider are deleted by default
func (x *LocalProvidersManager) RemoveProviders(ctx context.Context, providerNameVersionSlice ...string) *schema.Diagnostics {

	diagnostics := schema.NewDiagnostics()

	if len(providerNameVersionSlice) == 0 {
		return diagnostics.AddErrorMsg("Must specify at least one provider version for remove, for example: aws@v0.0.1")
	}

	// Analyze whether the providers to be deleted are valid and exist
	deleteActionSlice := make([]func() *schema.Diagnostics, 0)
	for _, providerNameVersion := range providerNameVersionSlice {
		nameVersion := version.ParseNameAndVersion(providerNameVersion)

		if nameVersion.Version == "" {
			diagnostics.AddErrorMsg("The version number cannot be empty. Specify the version number in the format providerName@version, for example: aws@v0.0.1")
			continue
		} else if nameVersion.IsLatestVersion() {
			diagnostics.AddErrorMsg("The version number cannot be latest. Specify a version number, for example: aws@v0.0.1")
			continue
		}

		path := x.buildLocalProviderVersionPath(nameVersion.Name, nameVersion.Version)
		if !utils.Exists(path) {
			diagnostics.AddErrorMsg("Provider version %s not found in %s", providerNameVersion, x.buildLocalProvidersPath())
			continue
		}

		deleteActionSlice = append(deleteActionSlice, func() *schema.Diagnostics {
			err := os.RemoveAll(path)
			if err != nil {
				return schema.NewDiagnostics().AddErrorMsg("Remove provider %s at local directory %s failed: %s", nameVersion.String(), path, err.Error())
			} else {
				return schema.NewDiagnostics().AddInfo("Remove provider %s at local directory %s success", nameVersion.String(), path)
			}
		})

	}
	if diagnostics.HasError() {
		return diagnostics
	}

	// Perform a delete operation
	for _, action := range deleteActionSlice {
		diagnostics.AddDiagnostics(action())
	}

	return diagnostics
}

//func RemoveProviders(names []string) error {
//	argsMap := make(map[string]bool)
//	for i := range names {
//		argsMap[names[i]] = true
//	}
//	deletedMap := make(map[string]bool)
//	err := config.IsSelefraWorkspace()
//	if err != nil {
//		ui.PrintErrorLn(err.Error())
//		return err
//	}
//	var cof = &config.SelefraBlock{}
//
//	namespace, _, err := utils.Home()
//	if err != nil {
//		return err
//	}
//	provider := registry.NewProviderRegistry(namespace)
//	err = cof.UnmarshalConfig()
//	if err != nil {
//		return err
//	}
//	for _, p := range cof.Selefra.Providers {
//		name := *p.Source
//		path := utils.GetPathBySource(*p.Source, p.Version)
//		prov := registry.ProviderBinary{
//			Provider: registry.Provider{
//				Name:    name,
//				Version: p.Version,
//				Source:  "",
//			},
//			FilePath: path,
//		}
//		if !argsMap[p.Name] || deletedMap[p.Path] {
//			break
//		}
//
//		err := provider.DeleteProvider(prov)
//		if err != nil {
//			if !errors.Is(err, os.ErrNotExist) {
//				ui.PrintWarningF("Failed to remove  %s: %s", p.Name, err.Error())
//			}
//		}
//		_, jsonPath, err := utils.Home()
//		if err != nil {
//			return err
//		}
//		c, err := os.ReadFile(jsonPath)
//		if err == nil {
//			var configMap = make(map[string]string)
//			err = json.Unmarshal(c, &configMap)
//			if err != nil {
//				return err
//			}
//			delete(configMap, *p.Source+"@"+p.Version)
//			c, err = json.Marshal(configMap)
//			if err != nil {
//				return err
//			}
//			err = os.RemoveProviders(jsonPath)
//			if err != nil {
//				return err
//			}
//			err = os.WriteFile(jsonPath, c, 0644)
//			if err != nil {
//				return err
//			}
//			deletedMap[path] = true
//		}
//		ui.PrintSuccessF("Removed %s success", *p.Source)
//	}
//	return nil
//}

//// DeleteProvider Delete the provider of a given version
//DeleteProvider(binary *ProviderBinary) error

//func (x *ProviderGithubRegistry) DeleteProvider(binary *ProviderBinary) error {
//	return x.deleteProviderBinary(binary)
//}
//
//func (x *ProviderGithubRegistry) deleteProviderBinary(binary *ProviderBinary) error {
//	if _, err := os.Stat(binary.FilePath); err != nil {
//		return err
//	}
//	return os.RemoveAll(filepath.Dir(binary.FilePath))
//}

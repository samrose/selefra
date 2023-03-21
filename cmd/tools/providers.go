package tools

import (
	"context"
	"github.com/selefra/selefra-provider-sdk/storage/database_storage/postgresql_storage"
	"github.com/selefra/selefra/config"
	"github.com/selefra/selefra/pkg/storage/pgstorage"
	"strconv"
	"strings"
	"time"
)

//// ProviderConfigStrs find all selefra provider config by id from selefra config and return provider config in string format
//// TODO: deprecated
//func ProviderConfigStrs(config *config.RootConfig, id string) ([]string, error) {
//	var providerConfs []string
//	for _, group := range config.Providers.Content {
//		for i, node := range group.Content {
//			if node.Kind == yaml.ScalarNode && node.Value == "provider" && group.Content[i+1].Value == id {
//				b, err := yaml.Marshal(group)
//				if err != nil {
//					return nil, err
//				}
//				providerConfs = append(providerConfs, string(b))
//
//			}
//		}
//	}
//	return providerConfs, nil
//}
//
//// ProvidersByID find all provider in rootConfig by id
//func ProvidersByID(rootConfig *config.RootConfig, id string) []*config.ProviderBlock {
//	var prvds = make([]*config.ProviderBlock, 0)
//	for _, group := range rootConfig.Providers.Content {
//		for i, node := range group.Content {
//			if node.Kind == yaml.ScalarNode && node.Value == "provider" && group.Content[i+1].Value == id {
//				b, err := yaml.Marshal(group)
//				if err != nil {
//					continue
//				}
//
//				var prvd config.ProviderBlock
//				if err := yaml.Unmarshal(b, &prvd); err != nil {
//					continue
//				}
//
//				prvds = append(prvds, &prvd)
//			}
//		}
//	}
//
//	return prvds
//}
//
//// SetProviderTmpl set the provider yaml template
//func SetProviderTmpl(template string, provider registry.ProviderBinary, config *config.RootConfig) error {
//	if config.Providers.Kind != yaml.SequenceNode {
//		config.Providers.Kind = yaml.SequenceNode
//		config.Providers.Tag = "!!seq"
//		config.Providers.Value = ""
//		config.Providers.Content = []*yaml.Node{}
//	}
//
//	var node yaml.Node
//
//	err := yaml.Unmarshal([]byte(template), &node)
//	if err != nil {
//		return err
//	}
//	var provNode yaml.Node
//	if node.Content == nil {
//		provNode.Content = []*yaml.Node{
//			{
//				Kind: yaml.MappingNode,
//				Tag:  "!!map",
//				Content: append([]*yaml.Node{
//					{
//						Kind:  yaml.ScalarNode,
//						Value: "name",
//					},
//					{
//						Kind:  yaml.ScalarNode,
//						Value: provider.Name,
//					},
//					{
//						Kind:        yaml.ScalarNode,
//						Value:       "provider",
//						FootComment: template,
//					},
//					{
//						Kind:  yaml.ScalarNode,
//						Value: provider.Name,
//					},
//				}),
//			},
//		}
//	} else {
//		provNode.Content = []*yaml.Node{
//			{
//				Kind: yaml.MappingNode,
//				Tag:  "!!map",
//				Content: append([]*yaml.Node{
//					{
//						Kind:  yaml.ScalarNode,
//						Value: "name",
//					},
//					{
//						Kind:  yaml.ScalarNode,
//						Value: provider.Name,
//					},
//					{
//						Kind:        yaml.ScalarNode,
//						Value:       "provider",
//						FootComment: template,
//					},
//					{
//						Kind:  yaml.ScalarNode,
//						Value: provider.Name,
//					},
//				}),
//			},
//		}
//	}
//
//	config.Providers.Content = append(config.Providers.Content, provNode.Content...)
//
//	return nil
//}
//
//// AppendProviderDecl append a provider declare for rootConfig.Selefra.RequireProvidersBlock
//func AppendProviderDecl(provider registry.ProviderBinary, rootConfig *config.RootConfig, configVersion string) error {
//	source, latestSource := utils.CreateSource(provider.Name, provider.Version, configVersion)
//	_, configPath, err := utils.Home()
//	if err != nil {
//		cli_ui.Errorln("SetSelefraProviderError: " + err.Error())
//		return err
//	}
//
//	var pathMap = make(map[string]string)
//	file, err := os.ReadFile(configPath)
//	if err != nil {
//		cli_ui.Errorln("SetSelefraProviderError: " + err.Error())
//		return err
//	}
//	json.Unmarshal(file, &pathMap)
//	if latestSource != "" {
//		pathMap[latestSource] = provider.Filepath
//	}
//
//	pathMap[source] = provider.Filepath
//
//	pathMapJson, err := json.Marshal(pathMap)
//
//	if err != nil {
//		cli_ui.Errorln("SetSelefraProviderError: " + err.Error())
//	}
//
//	err = os.WriteFile(configPath, pathMapJson, 0644)
//	if rootConfig != nil {
//		rootConfig.Selefra.ProviderDecls = append(rootConfig.Selefra.ProviderDecls, &config.RequireProvider{
//			Name:    provider.Name,
//			Source:  &strings.Split(source, "@")[0],
//			Version: provider.Version,
//		})
//	}
//	return nil
//}

// CacheExpired check whether the cache time expires
func CacheExpired(ctx context.Context, storage *postgresql_storage.PostgresqlStorage, cacheTime string) (bool, error) {
	requireKey := config.GetCacheKey()
	fetchTime, err := pgstorage.GetStorageValue(ctx, storage, requireKey)
	if err != nil {
		return true, err
	}
	fetchTimeLocal, err := time.ParseInLocation(time.RFC3339, fetchTime, time.Local)
	if err != nil {
		return true, err
	}

	duration, err := parseDuration(cacheTime)
	if err != nil || duration == 0 {
		return true, err
	}
	if time.Now().Sub(fetchTimeLocal) > duration {
		return true, nil
	}

	return false, nil
}

func parseDuration(d string) (time.Duration, error) {
	d = strings.TrimSpace(d)
	dr, err := time.ParseDuration(d)
	if err == nil {
		return dr, nil
	}
	if strings.Contains(d, "d") {
		index := strings.Index(d, "d")

		hour, _ := strconv.Atoi(d[:index])
		dr = time.Hour * 24 * time.Duration(hour)
		ndr, err := time.ParseDuration(d[index+1:])
		if err != nil {
			return dr, nil
		}
		return dr + ndr, nil
	}

	dv, err := strconv.ParseInt(d, 10, 64)
	return time.Duration(dv), err
}

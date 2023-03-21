package version

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"sort"
	"strings"
)

const (

	// VersionLatest Indicates the latest version
	VersionLatest = "latest"

	// NameVersionDelimiter The separator character for name and version
	NameVersionDelimiter = "@"
)

// ------------------------------------------------- --------------------------------------------------------------------

// NameAndVersion A key-val ue pair representing a name and version
type NameAndVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewNameAndVersion(name, version string) *NameAndVersion {
	return &NameAndVersion{
		Name:    name,
		Version: version,
	}
}

// ParseNameAndVersion example: aws@v0.0.1
func ParseNameAndVersion(nameAndVersion string) *NameAndVersion {
	split := strings.Split(nameAndVersion, NameVersionDelimiter)
	var name, version string
	if len(split) > 1 {
		name = split[0]
		version = split[1]
	} else {
		name = split[0]
		version = VersionLatest
	}
	return &NameAndVersion{
		Name:    name,
		Version: version,
	}
}

// IsLatestVersion Check whether the version number indicates the latest version
func (x *NameAndVersion) IsLatestVersion() bool {
	return IsLatestVersion(x.Version)
}

func (x *NameAndVersion) String() string {
	if x.Version == "" {
		return x.Name
	}
	return fmt.Sprintf("%s%s%s", x.Name, NameVersionDelimiter, x.Version)
}

// ------------------------------------------------- --------------------------------------------------------------------

// Sort version numbers
func Sort(versionsRaw []string) []string {

	versions := make([]*version.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		v, _ := version.NewVersion(raw)
		versions[i] = v
	}

	// After this, the versions are properly sorted
	collection := version.Collection(versions)
	sort.Sort(collection)

	newVersions := make([]string, len(collection))
	for index, version := range collection {
		newVersions[index] = "v" + version.String()
	}
	return newVersions
}

// ------------------------------------------------- --------------------------------------------------------------------

// IsConstraintsAllow Determines whether the given version conforms to the version constraint
func IsConstraintsAllow(constraints version.Constraints, version *version.Version) bool {
	for _, c := range constraints {
		if c.Check(version) {
			return true
		}
	}
	return false
}

// ------------------------------------------------- --------------------------------------------------------------------

func IsLatestVersion(versionString string) bool {
	return versionString == "" || VersionLatest == strings.ToLower(versionString)
}

// ------------------------------------------------- --------------------------------------------------------------------

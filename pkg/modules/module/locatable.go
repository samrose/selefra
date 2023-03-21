package module

import (
	"github.com/golang-infrastructure/go-trie"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"unicode/utf8"
)

// ------------------------------------------------- --------------------------------------------------------------------

// Locatable Used to find the file and location of each block. All blocks should implement this interface
type Locatable interface {

	// GetNodeLocation Gets the location of the block
	GetNodeLocation(selector string) *NodeLocation

	// SetNodeLocation Set the location of the node
	SetNodeLocation(selector string, nodeLocation *NodeLocation) error
}

// NodeLocation A piece of location information used to represent a block
type NodeLocation struct {

	// The path from the root node of yaml to the current node
	YamlSelector string

	// path for find file, It is usually stored in a file system, which is the location of a file
	Path string

	// Represents a continuous piece of text in a file, with a starting position and an ending position
	Begin, End *Position
}

func BuildLocationFromYamlNode(yamlFilePath string, yamlSelector string, node *yaml.Node) *NodeLocation {
	endNode := rightLeafNode(node)
	if endNode == nil {
		return nil
	}
	return &NodeLocation{
		Path:         yamlFilePath,
		YamlSelector: yamlSelector,
		Begin:        NewPosition(node.Line, node.Column),
		End:          NewPosition(endNode.Line, endNode.Column+utf8.RuneCountInString(endNode.Value)),
	}
}

// Gets the end point of a node
func rightLeafNode(node *yaml.Node) *yaml.Node {
	if node == nil || node.Kind == yaml.ScalarNode || len(node.Content) == 0 {
		return node
	}
	return rightLeafNode(node.Content[len(node.Content)-1])
}

// ReadSourceString Read the source string content based on location information
func (x *NodeLocation) ReadSourceString() string {
	if x == nil {
		return ""
	}
	file, err := os.ReadFile(x.Path)
	if err != nil {
		return err.Error()
	}
	split := strings.Split(string(file), "\n")
	buff := strings.Builder{}
	inCollection := false
loop:
	for lineIndex, lineString := range split {
		for columnIndex, columnCharacter := range lineString {
			if (lineIndex+1) >= x.Begin.Line && (columnIndex+1) >= x.Begin.Column {
				inCollection = true
			}
			if inCollection {
				buff.WriteRune(columnCharacter)
			}
			if (lineIndex+1) >= x.End.Line && (columnIndex+1) >= x.End.Column {
				inCollection = false
				break loop
			}
		}
		if inCollection {
			buff.WriteRune('\n')
		}
	}
	return buff.String()
}

// ------------------------------------------------- --------------------------------------------------------------------

// Position Represents a point in a file
type Position struct {

	// which line
	Line int

	// which column
	Column int
}

func NewPosition(line, column int) *Position {
	return &Position{
		Line:   line,
		Column: column,
	}
}

// ------------------------------------------------- --------------------------------------------------------------------

type LocatableImpl struct {
	yamlSelectorTrie *trie.Trie[*NodeLocation]
}

var _ Locatable = &LocatableImpl{}

func NewLocatableImpl() *LocatableImpl {
	return &LocatableImpl{
		// TODO Improve the efficiency of the tree
		yamlSelectorTrie: trie.New[*NodeLocation](trie.DefaultPathSplitFunc),
	}
}

const (
	NodeLocationSelfKey   = "._key"
	NodeLocationSelfValue = "._value"
)

func (x *LocatableImpl) GetNodeLocation(relativeSelector string) *NodeLocation {

	// Example(with ._key or ._value):
	// foo._key
	// foo._value
	selectorPathLocation, err := x.yamlSelectorTrie.Query(relativeSelector)
	if err == nil {
		return selectorPathLocation
	}

	// Example(without ._key or ._value):
	// foo
	// bar
	keyLocation, keyErr := x.yamlSelectorTrie.Query(relativeSelector + NodeLocationSelfKey)
	valueLocation, valueErr := x.yamlSelectorTrie.Query(relativeSelector + NodeLocationSelfValue)
	if keyErr != nil && valueErr != nil {
		return nil
	}
	return MergeKeyValueLocation(keyLocation, valueLocation)
}

func MergeKeyValueLocation(keyLocation, valueLocation *NodeLocation) *NodeLocation {
	if keyLocation == nil {
		return valueLocation
	} else if valueLocation == nil {
		return keyLocation
	}

	return &NodeLocation{
		YamlSelector: keyLocation.YamlSelector,
		Path:         keyLocation.Path,
		Begin:        keyLocation.Begin,
		End:          valueLocation.End,
	}
}

// foo.bar.key --> foo.bar
// foo.bar[1] --> foo.bar
func baseYamlSelector(yamlSelector string) string {
	if len(yamlSelector) == 0 {
		return ""
	}

	// Look for boundary characters
	var delimiterCharacter byte
	switch yamlSelector[len(yamlSelector)-1] {
	case ']':
		delimiterCharacter = '['
	default:
		delimiterCharacter = '.'
	}

	for index := len(yamlSelector) - 2; index >= 0; index-- {
		if yamlSelector[index] == delimiterCharacter {
			return yamlSelector[0:index]
		}
	}

	return ""
}

func (x *LocatableImpl) SetNodeLocation(relativeSelector string, nodeLocation *NodeLocation) error {
	return x.yamlSelectorTrie.Add(relativeSelector, nodeLocation)
}

// ------------------------------------------------- --------------------------------------------------------------------

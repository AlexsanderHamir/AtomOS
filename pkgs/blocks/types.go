package blocks

import "gopkg.in/yaml.v3"

type BlockInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	Source      struct {
		Type string `yaml:"type"`
		Repo string `yaml:"repo"`
	} `yaml:"source"`
	Binary struct {
		From   string            `yaml:"from"`
		Assets map[string]string `yaml:"assets"`
	} `yaml:"binary"`
	Entries    map[string]Entry `yaml:"entries"`
	BinaryPath string           // Path to the downloaded binary
}

// Entry represents a CLI entry from the block
type Entry struct {
	Name        string   `yaml:"name"`
	Command     string   `yaml:"command"`
	Description string   `yaml:"description"`
	Inputs      []Input  `yaml:"inputs"`
	Outputs     []Output `yaml:"outputs"`
}

// Input represents an input parameter for an entry
type Input struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// Output represents an output from an entry
type Output struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// UnmarshalYAML implements custom YAML unmarshaling for BlockInfo
// to convert the entries array into a map keyed by entry name
func (b *BlockInfo) UnmarshalYAML(value *yaml.Node) error {
	// Define a temporary struct that matches the YAML structure
	type blockInfoYAML struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Version     string `yaml:"version"`
		Source      struct {
			Type string `yaml:"type"`
			Repo string `yaml:"repo"`
		} `yaml:"source"`
		Binary struct {
			From   string            `yaml:"from"`
			Assets map[string]string `yaml:"assets"`
		} `yaml:"binary"`
		Entries []Entry `yaml:"entries"`
	}

	var temp blockInfoYAML
	if err := value.Decode(&temp); err != nil {
		return err
	}

	// Copy all fields to the actual struct
	b.Name = temp.Name
	b.Description = temp.Description
	b.Version = temp.Version
	b.Source = temp.Source
	b.Binary = temp.Binary

	// Convert entries slice to map
	b.Entries = make(map[string]Entry)
	for _, entry := range temp.Entries {
		b.Entries[entry.Name] = entry
	}

	return nil
}

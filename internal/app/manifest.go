package app

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// StringOrSlice is a custom type that allows unmarshalling a YAML field as either a single string or a slice of strings.
//
// # Usage
//
//	var s StringOrSlice
//	err := yaml.Unmarshal([]byte("- foo\n- bar"), &s)
//	// s == []string{"foo", "bar"}
//
// # Example
//
//	s := StringOrSlice{"foo", "bar"}
type StringOrSlice []string

// UnmarshalYAML implements the yaml.Unmarshaler interface for StringOrSlice.
// It allows the field to be unmarshalled from either a single string or a sequence of strings.
//
// # Parameters
//   - value: the YAML node to decode
//
// # Returns
//   - error: if decoding fails or the node kind is unsupported
func (s *StringOrSlice) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var str string
		if err := value.Decode(&str); err != nil {
			return err
		}
		*s = []string{str}
		return nil
	case yaml.SequenceNode:
		var arr []string
		if err := value.Decode(&arr); err != nil {
			return err
		}
		*s = arr
		return nil
	default:
		return &yaml.TypeError{Errors: []string{"unsupported YAML node kind for StringOrSlice"}}
	}
}

// SoftwareEntry represents a single software entry in the manifest, including metadata and installation methods.
//
// # Fields
//   - Bin, Desc, Docs, Github, Home, Name, Short, Groups: metadata fields
//   - Brew, Apt, Pacman, etc.: installation methods for various package managers
//
// # Example
//
//	entry := SoftwareEntry{Name: "bat", Brew: StringOrSlice{"bat"}}
type SoftwareEntry struct {
	Bin           StringOrSlice `yaml:"_bin"`
	Desc          string        `yaml:"_desc"`
	Docs          string        `yaml:"_docs"`
	Github        string        `yaml:"_github"`
	Home          string        `yaml:"_home"`
	Name          string        `yaml:"_name"`
	Short         string        `yaml:"_short"`
	Groups        StringOrSlice `yaml:"_groups"`
	Brew          StringOrSlice `yaml:"brew"`
	Apt           StringOrSlice `yaml:"apt"`
	Pacman        StringOrSlice `yaml:"pacman"`
	Choco         StringOrSlice `yaml:"choco"`
	Go            StringOrSlice `yaml:"go"`
	Snap          StringOrSlice `yaml:"snap"`
	Port          StringOrSlice `yaml:"port"`
	Scoop         StringOrSlice `yaml:"scoop"`
	Yay           StringOrSlice `yaml:"yay"`
	Apk           StringOrSlice `yaml:"apk"`
	Dnf           StringOrSlice `yaml:"dnf"`
	Pkg           StringOrSlice `yaml:"pkg"`
	Cask          StringOrSlice `yaml:"cask"`
	Flatpak       StringOrSlice `yaml:"flatpak"`
	Mas           StringOrSlice `yaml:"mas"`
	Nix           StringOrSlice `yaml:"nix"`
	PkgTermux     StringOrSlice `yaml:"pkg-termux"`
	Emerge        StringOrSlice `yaml:"emerge"`
	NixEnv        StringOrSlice `yaml:"nix-env"`
	BinaryDarwin  StringOrSlice `yaml:"binary:darwin"`
	BinaryLinux   StringOrSlice `yaml:"binary:linux"`
	BinaryWindows StringOrSlice `yaml:"binary:windows"`
	Xbps          StringOrSlice `yaml:"xbps"`
	Zypper        StringOrSlice `yaml:"zypper"`
	Cargo         StringOrSlice `yaml:"cargo"`
	Pipx          StringOrSlice `yaml:"pipx"`
	// Add more fields as needed
}

// Manifest represents the full manifest mapping software names to their entries.
//
// # Example
//
//	m := Manifest{"bat": SoftwareEntry{...}}
type Manifest map[string]SoftwareEntry

// LoadManifest loads a manifest from a YAML file at the given path.
//
// # Parameters
//   - path: the path to the YAML manifest file
//
// # Returns
//   - Manifest: the loaded manifest
//   - error: if the file cannot be opened or decoded
//
// # Example
//
//	m, err := LoadManifest("software.yml")
func LoadManifest(path string) (Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	var m Manifest
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

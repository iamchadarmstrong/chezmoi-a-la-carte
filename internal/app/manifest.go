package app

import (
	"os"

	"gopkg.in/yaml.v3"
)

type StringOrSlice []string

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
	}
	return nil
}

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

type Manifest map[string]SoftwareEntry

func LoadManifest(path string) (Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var m Manifest
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

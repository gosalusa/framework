package util

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
	"gosalusa.com/stream"
)

type Package struct {
	Dir    string `yaml:"dir"`
	Pkg    string `yaml:"package"`
	Import string `yaml:"import"`
}

type Config struct {
	Root      string   `yaml:"-"`
	Module    string   `yaml:"module"`
	Model     *Package `yaml:"model"`
	Migration *Package `yaml:"migration"`
}

var ErrNotFound = errors.New("file not found")

const configFileName = "spice.yml"

func defaultConfig() *Config {
	return &Config{
		Model:     &Package{Dir: "app/models"},
		Migration: &Package{Dir: "migrations"},
	}
}

func LoadConfig(from string) (*Config, error) {
	root, err := FileDir(from, configFileName)
	if err != nil {
		return nil, err
	}

	pkgRoot, err := FileDir(from, "go.mod")
	if err != nil {
		return nil, err
	}
	pkg, err := PkgInfo(pkgRoot, root)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path.Join(root, configFileName))
	if err != nil {
		return nil, err
	}

	c := defaultConfig()
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	c.prep(root, pkg)

	return c, nil
}

func (c *Config) prep(root string, pkgInfo *PackageInfo) {
	c.Root = root
	v := reflect.ValueOf(c).Elem()

	for i := 0; i < v.NumField(); i++ {
		pkg, ok := v.Field(i).Interface().(*Package)
		if !ok {
			continue
		}
		pkg.Dir = path.Join(root, pkg.Dir)
		if pkg.Pkg == "" {
			pkg.Pkg = path.Base(pkg.Dir)
		}
		if pkg.Import == "" {
			imp, ok := strings.CutPrefix(pkg.Dir, root)
			if ok {
				pkg.Import = path.Join(pkgInfo.SpicePackage, imp)
			}
		}
	}
}

func FileDir(from, file string) (string, error) {
	cwd, err := filepath.Abs(from)
	if err != nil {
		return "", err
	}
	current := cwd
	name := ""
	relPackage := ""
	for {
		files, err := os.ReadDir(current)
		if err != nil {
			return "", err
		}

		_, ok := stream.Of(files).Find(func(f fs.DirEntry) bool {
			return f.Name() == file
		})
		if ok {
			return current, nil
		}

		if current == "/" {
			return "", fmt.Errorf("no file %s: %w", file, ErrNotFound)
		}
		current = strings.TrimSuffix(current, "/")
		current, name = path.Split(current)
		relPackage = path.Join(name, relPackage)
	}
}

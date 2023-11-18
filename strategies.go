package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
)

var strategies = []Strategy{
	newComposerStrategy(),
	newNPMStrategy(),
}

type Strategy struct {
	Name        string
	matchFiles  []string
	getPackages func(r io.Reader) (map[string]Package, error)
}

func (s Strategy) GetPackages(r io.Reader, dev bool) (map[string]Package, error) {
	packages, err := s.getPackages(r)
	if err != nil {
		return nil, err
	}
	if !dev {
		for name, pkg := range packages {
			if pkg.IsDev {
				delete(packages, name)
			}
		}
	}
	return packages, nil
}

var ErrStrategyNotFound = errors.New("strategy not found")

func findStrategyByName(name string) (Strategy, error) {
	for _, s := range strategies {
		if s.Name == name {
			return s, nil
		}
	}

	return Strategy{}, ErrStrategyNotFound
}

func findStrategyByPath(path string) (Strategy, error) {
	target := filepath.Base(path)
	for _, s := range strategies {
		for _, fileName := range s.matchFiles {
			if fileName == target {
				return s, nil
			}
		}
	}

	return Strategy{}, ErrStrategyNotFound
}

type composerJSON struct {
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
}

func newComposerStrategy() Strategy {
	return Strategy{
		Name:        "composer",
		matchFiles:  []string{"composer.json"},
		getPackages: getComposerPackages,
	}
}

func getComposerPackages(r io.Reader) (map[string]Package, error) {
	c := &composerJSON{}
	if err := json.NewDecoder(r).Decode(c); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	totalPkgCount := len(c.Require) + len(c.RequireDev)
	packages := make(map[string]Package, totalPkgCount)

	for pkg, ver := range c.Require {
		packages[pkg] = Package{
			Name:    pkg,
			Version: ver,
			IsDev:   false,
		}
	}
	for pkg, ver := range c.RequireDev {
		packages[pkg] = Package{
			Name:    pkg,
			Version: ver,
			IsDev:   true,
		}
	}

	return packages, nil
}

type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func newNPMStrategy() Strategy {
	return Strategy{
		Name:        "npm",
		matchFiles:  []string{"package.json"},
		getPackages: getNPMPackages,
	}
}

func getNPMPackages(r io.Reader) (map[string]Package, error) {
	p := &packageJSON{}
	if err := json.NewDecoder(r).Decode(p); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	totalPkgCount := len(p.Dependencies) + len(p.DevDependencies)
	packages := make(map[string]Package, totalPkgCount)

	for pkg, ver := range p.Dependencies {
		packages[pkg] = Package{
			Name:    pkg,
			Version: ver,
			IsDev:   false,
		}
	}
	for pkg, ver := range p.DevDependencies {
		packages[pkg] = Package{
			Name:    pkg,
			Version: ver,
			IsDev:   true,
		}
	}

	return packages, nil
}

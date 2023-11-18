package main

import "time"

type Package struct {
	Name    string `json:"name" csv:"name"`
	Version string `json:"version" csv:"version"`
	IsDev   bool   `json:"is_dev" csv:"is_dev"`
}

type ChangeEvent string

const (
	PackageAdded   ChangeEvent = "added"
	PackageUpdated ChangeEvent = "updated"
	PackageDeleted ChangeEvent = "deleted"
)

type PackageChange struct {
	Package

	OldVersion string      `json:"old_version,omitempty" csv:"old_version,omitempty"`
	Author     string      `json:"author,omitempty" csv:"author,omitempty"`
	Email      string      `json:"email,omitempty" csv:"email,omitempty"`
	Time       time.Time   `json:"time" csv:"time"`
	Commit     string      `json:"commit,omitempty" csv:"commit,omitempty"`
	Event      ChangeEvent `json:"event,omitempty" csv:"event,omitempty"`
}

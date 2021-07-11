package schemadance

import (
	"context"
	"io/fs"
)

type Tx interface {
	Exec(ctx context.Context, sql string) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Version(ctx context.Context, prefix string) (int, error)
	SetVersion(ctx context.Context, prefix string, version int) error
}

type Db interface {
	Initialize(ctx context.Context) error
	Begin(ctx context.Context) (Tx, error)
	Close(ctx context.Context) error
}

type Patch struct {
	Prefix      string
	From        int
	To          int
	Description string
	Filename    string
	Before      func(ctx context.Context, tx Tx, from, to int) error
	After       func(ctx context.Context, tx Tx, from, to int) error
}

type PatchSet struct {
	Name  string
	Up    []Patch
	Down  []Patch
}

type PatchSets struct {
	Sets []PatchSet
	Files fs.FS
}

type StatusPhase int

const (
	Start StatusPhase = iota
	End
	Failed
)

type Status struct {
	Prefix      string
	From        int
	To          int
	Description string
	Filename    string
	Phase       StatusPhase
}

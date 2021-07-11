package schemadance

import (
	"context"
	"fmt"
	"io/fs"
)

type Migrator struct {
	Database Db
	Patches  PatchSets
	Partial  bool
	Status   chan Status
}

func (m Migrator) Version(ctx context.Context, set string) (ver int, rerr error) {
	tx, err := m.Database.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		err := tx.Rollback(ctx)
		if rerr == nil {
			rerr = err
		}
	}()
	return tx.Version(ctx, set)
}

func (m Migrator) Migrate(ctx context.Context, set string, to int) error {
	for _, p := range m.Patches.Sets {
		if p.Name == set {
			currentVer, err := m.Version(ctx, set)
			if err != nil {
				return err
			}
			patches, err := Plan(currentVer, to, p)
			if err != nil {
				return err
			}
			return m.apply(ctx, patches, m.Patches.Files)
		}
	}
	return fmt.Errorf("no such patch set as %s", set)
}

func (m Migrator) Update(ctx context.Context) error {
	var patches []Patch
	for _, set := range m.Patches.Sets {
		currentVer, err := m.Version(ctx, set.Name)
		if err != nil {
			return err
		}
		maxVer := set.EndVersion()
		p, err := Plan(currentVer, maxVer, set)
		if err != nil {
			return err
		}
		patches = append(patches, p...)
	}
	return m.apply(ctx, patches, m.Patches.Files)
}

func (m Migrator) apply(ctx context.Context, patches []Patch, filesystem fs.FS) (rerr error) {
	if len(patches) == 0 {
		return nil
	}
	err := m.Database.Initialize(ctx)
	if err != nil {
		return err
	}
	tx, err := m.Database.Begin(ctx)
	if err != nil {
		return err
	}

	defer func(){
		err := tx.Rollback(ctx)
		if rerr == nil {
			rerr = err
		}
	}()

	for _, p := range patches {
		if m.Status != nil {
			m.Status <- Status{
				Prefix:      p.Prefix,
				From:        p.From,
				To:          p.To,
				Description: p.Description,
				Filename:    p.Filename,
				Phase:       Start,
			}
		}

		currentVersion, err := tx.Version(ctx, p.Prefix)
		if err != nil {
			return fmt.Errorf("while retrieving version: %w", err)
		}
		if currentVersion != p.From {
			return fmt.Errorf("found current version %d, expected %d while processing %s", currentVersion, p.From, p.Filename)
		}

		if p.Before != nil {
			err = p.Before(ctx, tx, p.From, p.To)
			if err != nil {
				return fmt.Errorf("before applying %s: %w", p.Filename, err)
			}
		}

		sql, err := fs.ReadFile(filesystem, p.Filename)
		if err != nil {
			return err
		}

		err = tx.Exec(ctx, string(sql))
		if err != nil {
			return fmt.Errorf("while applying %s: %w", p.Filename, err)
		}

		if p.After != nil {
			err = p.After(ctx, tx, p.From, p.To)
			if err != nil {
				return fmt.Errorf("after applying %s: %w", p.Filename, err)
			}
		}

		err = tx.SetVersion(ctx, p.Prefix, p.To)

		if m.Partial {
			err = tx.Commit(ctx)
			if err != nil {
				return fmt.Errorf("while committing %s: %w", p.Filename, err)
			}
			tx, err = m.Database.Begin(ctx)
			if err != nil {
				return err
			}
		}
		if m.Status != nil {
			m.Status <- Status{
				Prefix:      p.Prefix,
				From:        p.From,
				To:          p.To,
				Description: p.Description,
				Filename:    p.Filename,
				Phase:       End,
			}
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("at final commit: %w", err)
	}
	return nil
}

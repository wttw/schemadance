package schemadance

import "fmt"

func Plan(from int, to int, set PatchSet) ([]Patch, error) {
	if from == to {
		return []Patch{}, nil
	}
	if from < to {
		return planUp(from, to, set.Up)
	}
	return planDown(from, to, set.Down)
}

func planUp(from int, to int, patches []Patch) ([]Patch, error) {
	at := from
	var plan []Patch
	for _, p := range patches {
		if p.From < at {
			continue
		}
		if p.From > at {
			return nil, fmt.Errorf("failed to plan migration from %d to %d", at, to)
		}
		if p.To > to {
			continue
		}
		plan = append(plan, p)
		at = p.To
		if at == to {
			return plan, nil
		}
	}
	return nil, fmt.Errorf("failed to plan migration from %d to %d", from, to)
}

func planDown(from int, to int, patches []Patch) ([]Patch, error) {
	at := from
	var plan []Patch
	for _, p := range patches {
		if p.From > at {
			continue
		}
		if p.From < at {
			return nil, fmt.Errorf("failed to plan migration from %d to %d", at, to)
		}
		if p.To < to {
			continue
		}
		plan = append(plan, p)
		at = p.To
		if at == to {
			return plan, nil
		}
	}
	return nil, fmt.Errorf("failed to plan migration from %d to %d", from, to)
}


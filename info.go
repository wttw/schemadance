package schemadance

func (p PatchSet) EndVersion() int {
	max := 0
	for _, step := range p.Up {
		if step.To > max {
			max = step.To
		}
	}
	return max
}

func (p PatchSet) StartVersion() int {
	if len(p.Up) == 0 {
		return 0
	}
	return p.Up[0].From
}

func sqlOnly(p []Patch) bool {
	for _, step := range p {
		if step.Before != nil {
			return false
		}
		if step.After != nil {
			return false
		}
	}
	return true
}

func (p PatchSet) SQLOnly() bool {
	return sqlOnly(p.Up) && sqlOnly(p.Down)
}

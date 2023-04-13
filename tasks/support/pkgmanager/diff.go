package pkgmanager

type Diff struct {
	Added   []string
	Removed []string
}

func (d *Diff) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0
}

func CalcDiff(before, after []string) *Diff {
	diff := &Diff{
		Added:   make([]string, 0),
		Removed: make([]string, 0),
	}
	beforeMap := make(map[string]int, len(before))
	afterMap := make(map[string]int, len(after))

	for i := range after {
		afterMap[after[i]] = i
	}

	for i, beforeItem := range before {
		if beforeItem == "" {
			continue
		}

		beforeMap[beforeItem] = i
		if _, ok := afterMap[beforeItem]; ok {
			continue
		}
		diff.Removed = append(diff.Removed, beforeItem)
	}

	for _, afterItem := range after {
		if afterItem == "" {
			continue
		}

		if _, ok := beforeMap[afterItem]; !ok {
			diff.Added = append(diff.Added, afterItem)
			continue
		}
	}

	if diff.IsEmpty() {
		return nil
	}

	return diff
}

package names

type TaskNames struct { // Targets
	Name  string
	Names []string
}

func (nt *TaskNames) GetNames() []string {
	names := make([]string, 0, len(nt.Names)+1)
	if nt.Name != "" {
		names = append(names, nt.Name)
	}

	for _, name := range nt.Names {
		if name != "" {
			names = append(names, name)
		}
	}

	return names
}

package database

func ListModelToIDs[T Identifiable](models []T) []uint32 {
	if len(models) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(models))
	for i, model := range models {
		ids[i] = model.GetID()
	}
	return ids
}

package node

type Map map[string]Node
type Indexes map[int]string

func (m Map) Add(nodeToAdd Node) Map {
	// copy the map
	mapCopy := make(Map)
	for k, v := range m {
		mapCopy[k] = v
	}

	// add entry
	if _, ok := mapCopy[nodeToAdd.ID]; !ok {
		nextIndex := len(mapCopy)
		nodeToAdd.Index = nextIndex
		mapCopy[nodeToAdd.ID] = nodeToAdd
	}

	return mapCopy
}

func (m Map) Delete(nodeToRemove Node) Map {
	// copy the map
	mapCopy := make(Map)
	for k, v := range m {
		mapCopy[k] = v
	}

	// delete entry
	delete(mapCopy, nodeToRemove.ID)

	// shift indexes left to fill the gap
	for k, v := range mapCopy {
		if v.Index > nodeToRemove.Index {
			v.Index--
			mapCopy[k] = v
		}
	}
	return mapCopy
}

func (m Map) CreateIndexes() Indexes {
	indexes := make(Indexes)
	for k, v := range m {
		indexes[v.Index] = k
	}
	return indexes
}

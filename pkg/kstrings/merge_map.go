package kstrings

func MergeMaps(a, b map[string]string) map[string]string {
	newMap := make(map[string]string)
	for ka, va := range a {
		newMap[ka] = va
	}

	for kb, vb := range b {
		newMap[kb] = vb
	}

	return newMap
}

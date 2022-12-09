package rule

type Inherits []Info

func (i Inherits) Contains(key string) bool {
	for _, r := range i {
		if r.Name == key {
			return true
		}
	}

	return false
}

func (i Inherits) IndexOf(key string) int {
	for j, item := range i {
		if item.Name == key {
			return j
		}
	}

	return -1
}

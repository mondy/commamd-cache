package collection

func EveryWithError[C ~[]V, V any](collection C, predicate func(value V, index int) (bool, error)) (bool, error) {
	for index, value := range collection {
		if ok, err := predicate(value, index); err != nil {
			return false, err
		} else if !ok {
			return false, nil
		}
	}

	return true, nil
}

func SomeWithError[C ~[]V, V any](collection C, predicate func(value V, index int) (bool, error)) (bool, error) {
	for index, value := range collection {
		if ok, err := predicate(value, index); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}

	return false, nil
}

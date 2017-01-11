package jzon

type flag uintptr

// func (f flag) contains(mask ...flag) bool {

// 	if len(mask) == 0 {
// 		return true
// 	}

// 	for _, ff := range mask {
// 		f &= ff
// 	}

// 	if f > 0 {
// 		return true
// 	}
// 	return false
// }

// func (f *flag) add(mask ...flag) *flag {
// 	for _, ff := range mask {
// 		*f |= ff
// 	}
// 	return f
// }

// func (f *flag) remove(mask ...flag) *flag {
// 	for _, ff := range mask {
// 		*f &= ^ff
// 	}
// 	return f
// }

func contains(f flag, mask ...flag) bool {

	if len(mask) == 0 {
		return true
	}

	for _, ff := range mask {
		f &= ff
	}

	if f > 0 {
		return true
	}
	return false
}

func add(f flag, mask ...flag) flag {
	for _, ff := range mask {
		f |= ff
	}
	return f
}

func remove(f flag, mask ...flag) flag {
	for _, ff := range mask {
		f &= ^ff
	}
	return f
}

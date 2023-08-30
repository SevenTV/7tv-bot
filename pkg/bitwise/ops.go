package bitwise

func Set(flags, flag uint32) uint32 {
	return flags | flag
}

func UnSet(flags, flag uint32) uint32 {
	return flags &^ flag
}

func Has(flags, flag uint32) bool {
	return flags&flag == flag
}

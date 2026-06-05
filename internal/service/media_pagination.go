package service

func mediaListOffset(useCursor bool, page, pageSize int) int {
	if useCursor {
		return 0
	}
	return (page - 1) * pageSize
}

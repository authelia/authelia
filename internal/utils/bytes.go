package utils

// BytesJoin is an alternate form of bytes.Join which doesn't use a sep.
func BytesJoin(s ...[]byte) (dst []byte) {
	if len(s) == 0 {
		return []byte{}
	}

	if len(s) == 1 {
		return append([]byte(nil), s[0]...)
	}

	var (
		n, dstp int
	)

	for _, v := range s {
		n += len(v)
	}

	dst = make([]byte, n)

	for _, v := range s {
		dstp += copy(dst[dstp:], v)
	}

	return dst
}

package utils

// IsBoolCountLessThanN takes an int (n), bool (v), and then a variadic slice of bool (vals). If the number of bools in vals with
// the value v is more than n, it returns false, otherwise it returns true.
func IsBoolCountLessThanN(n int, v bool, vals ...bool) bool {
	lvals := len(vals)

	// If lvals (len of vals) is less than n it can't possibly have more than n so we can short circuit here.
	if lvals < n {
		return true
	}

	j := 0

	for i, val := range vals {
		if val == v {
			j++
		}

		// If lvals (len of vals) minus the current index (the remainder) plus the number of positives
		// is less than n we can short circuit here.
		if lvals-i+j < n {
			return true
		}

		if j > n {
			return false
		}
	}

	return true
}

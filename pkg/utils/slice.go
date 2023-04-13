package utils

func FindFirstSameKeyInTwoStringArray(a []string, b []string) string {
	for _, v := range a {
		for _, v2 := range b {
			if v2 == v {
				return v
			}
		}
	}
	return ""
}

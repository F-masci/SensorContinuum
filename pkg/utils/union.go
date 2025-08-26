package utils

func UniqueUnion(slices ...[]string) []string {
	m := make(map[string]struct{})
	for _, sl := range slices {
		for _, v := range sl {
			m[v] = struct{}{}
		}
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

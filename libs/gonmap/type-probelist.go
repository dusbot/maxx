package gonmap

type ProbeList []string

var emptyProbeList []string

func (p ProbeList) removeDuplicate() ProbeList {
	result := make([]string, 0, len(p))
	temp := map[string]struct{}{}
	for _, item := range p {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func (p ProbeList) exist(probeName string) bool {
	for _, name := range p {
		if name == probeName {
			return true
		}
	}
	return false
}

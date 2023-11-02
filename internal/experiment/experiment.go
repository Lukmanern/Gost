package experiment

func BuildBitGroups(permIDs ...int) map[int]int {
	groups := make(map[int]int)
	for _, id := range permIDs {
		group := (id - 1) / 8
		bitPosition := uint(id - 1 - (group * 8))
		groups[group+1] |= 1 << bitPosition
	}
	return groups
}

func CheckHasPermission(endpointPermID int, userPermissions map[int]int) bool {
	endpointBits := BuildBitGroups(endpointPermID)
	for key, requiredBits := range endpointBits {
		userBits, ok := userPermissions[key]
		if !ok || requiredBits&userBits == 0 {
			return false
		}
	}
	return true
}

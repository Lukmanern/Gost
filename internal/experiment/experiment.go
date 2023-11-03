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
	// it seems O(n), but it's actually O(1)
	// because length of $endpointBits is 1
	for key, requiredBits := range endpointBits {
		userBits, ok := userPermissions[key]
		if !ok || requiredBits&userBits == 0 {
			return false
		}
	}
	return true
}

package repository

//отсутсвует роль владельца (owner), т.к. владелец может быть только один
var rolesList = []string{"admin", "cashier"}

func roleIsExists(role string) bool {
	for _, r := range rolesList {
		if r == role {
			return true
		}
	}
	return false
}

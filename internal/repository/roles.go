package repository

var rolesList = []string{"owner", "admin", "cashier"}

func roleIsExists(role string) bool {
	for _, r := range rolesList {
		if r == role {
			return true
		}
	}
	return false
}

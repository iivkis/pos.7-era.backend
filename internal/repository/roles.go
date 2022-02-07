package repository

//roles
const (
	R_OWNER   = "owner"
	R_ADMIN   = "admin"
	R_CASHIER = "cashier"
)

var (
	rolesList = []string{R_OWNER, R_ADMIN, R_CASHIER}
	rolesMap  map[string]int
)

func init() {
	rolesMap = make(map[string]int)
	for i, role := range rolesList {
		rolesMap[role] = i
	}
}

func roleIsExists(role string) (ok bool) {
	_, ok = rolesMap[role]
	return
}

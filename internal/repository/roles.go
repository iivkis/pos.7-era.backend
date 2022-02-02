package repository

//roles
const (
	R_OWNER   = "owner"
	R_ADMIN   = "admin"
	R_CASHIER = "cashier"
)

var rolesList = map[string]int{
	R_OWNER:   0,
	R_ADMIN:   1,
	R_CASHIER: 2,
}

func roleIsExists(role string) (ok bool) {
	_, ok = rolesList[role]
	return
}

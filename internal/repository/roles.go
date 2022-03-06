package repository

//roles
const (
	R_ROOT     = "root" // используется только внутри приложения (например, для регистрации владельца). Пользователи не могут иметь роль root.
	R_OWNER    = "owner"
	R_DIRECTOR = "director"
	R_ADMIN    = "admin"
	R_CASHIER  = "cashier"
)

var (
	//роли, которые могу быть занесены в БД
	rolesMap map[string]int = map[string]int{
		R_OWNER:    1,
		R_DIRECTOR: 2,
		R_ADMIN:    3,
		R_CASHIER:  4,
	}
)

func roleIsExists(role string) (ok bool) {
	_, ok = rolesMap[role]
	return
}

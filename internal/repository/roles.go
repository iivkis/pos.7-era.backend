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
	rolesMapByString map[string]int = map[string]int{
		R_OWNER:    1,
		R_DIRECTOR: 2,
		R_ADMIN:    3,
		R_CASHIER:  4,
	}

	rolesMapByInt map[int]string = make(map[int]string)
)

func init() {
	for k, v := range rolesMapByString {
		rolesMapByInt[v] = k
	}
}

func roleIsExists(role string) (ok bool) {
	_, ok = rolesMapByString[role]
	return
}

func RoleNameToID(role string) int {
	return rolesMapByString[role]
}

func RoleIDToName(roleID int) string {
	return rolesMapByInt[roleID]
}

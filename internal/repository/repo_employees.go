package repository

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)

type EmployeeModel struct {
	ID        uint
	CreatedAt time.Time

	Name     string
	Password string
	Role     string
	Online   bool

	OrgID    uint
	OutletID uint

	DeletedAt gorm.DeletedAt `gorm:"index"`

	OrganizationModel OrganizationModel `gorm:"foreignKey:OrgID"`
	OutletModel       OutletModel       `gorm:"foreignKey:OutletID"`
}

func (m *EmployeeModel) GetRoleID() int {
	return RoleNameToID(m.Role)
}

func (m *EmployeeModel) passwordValidation() bool {
	n, err := strconv.Atoi(m.Password)
	if err != nil || n < 0 {
		return false
	}
	return true
}

//проверка, имеет ли сотрудник какую-либо роль из массива roles
func (m *EmployeeModel) hasRole(roles ...string) bool {
	for _, role := range roles {
		if role == m.Role {
			return true
		}
	}
	return false
}

type EmployeesRepo struct {
	db *gorm.DB
}

func newEmployeesRepo(db *gorm.DB) *EmployeesRepo {
	return &EmployeesRepo{
		db: db,
	}
}

func (r *EmployeesRepo) Create(m *EmployeeModel, myRole string) error {

	if !roleIsExists(m.Role) {
		return ErrUndefinedRole
	}

	if !m.passwordValidation() {
		return ErrOnlyNumCanBeInPassword
	}

	switch myRole {
	case R_ROOT: //нет ограничений
	case R_OWNER: //разрешено создавать директоров, админов и кассиров
		if !m.hasRole(R_DIRECTOR, R_ADMIN, R_CASHIER) {
			return ErrPermissionDenided
		}
	case R_DIRECTOR: //разрешено создавать админов и кассиров
		if !m.hasRole(R_ADMIN, R_CASHIER) {
			return ErrPermissionDenided
		}
	case R_ADMIN: //разрешено создавать кассиров
		if !m.hasRole(R_CASHIER) {
			return ErrPermissionDenided
		}
	default:
		return ErrPermissionDenided
	}

	return r.db.Create(m).Error
}

func (r *EmployeesRepo) SignIn(id uint, password string, orgID uint) (empl EmployeeModel, err error) {
	err = r.db.Where("id = ? AND org_id = ? AND password = ?", id, orgID, password).First(&empl).Error
	return
}

func (r *EmployeesRepo) Updates(m *EmployeeModel, employeeID interface{}, outletID interface{}, myRole string) error {
	var employee EmployeeModel
	if err := r.db.Where("id = ? AND outlet_id = ?", employeeID, outletID).First(&employee).Error; err != nil {
		return err
	}

	if !roleIsExists(m.Role) {
		return ErrUndefinedRole
	}

	if !m.passwordValidation() {
		return ErrOnlyNumCanBeInPassword
	}

	var updatedData *EmployeeModel

	//поля, которые могут изменять разные роли
	switch myRole {
	case R_OWNER: // может изменить данные любого
		updatedData = &EmployeeModel{
			Name:     m.Name,
			Password: m.Password,
		}

	case R_DIRECTOR: //может изменить свой пароль и любые данные нижестоящих ролей
		if employee.hasRole(R_DIRECTOR) && employee.ID == employeeID { //если пытается изменить себя
			updatedData = &EmployeeModel{
				Password: m.Password,
			}
		} else if employee.hasRole(R_ADMIN, R_CASHIER) {
			updatedData = &EmployeeModel{
				Name:     m.Name,
				Password: m.Password,
			}
		} else {
			return ErrPermissionDenided
		}

	case R_ADMIN: //может изменить свой пароль и любые данные нижестоящих ролей
		if employee.hasRole(R_ADMIN) && employee.ID == employeeID { //если пытается изменить себя
			updatedData = &EmployeeModel{
				Password: m.Password,
			}
		} else if employee.hasRole(R_CASHIER) {
			updatedData = &EmployeeModel{
				Name:     m.Name,
				Password: m.Password,
			}
		} else {
			return ErrPermissionDenided
		}

	default:
		return ErrPermissionDenided
	}

	return r.db.Where("id = ? AND outlet_id = ?", employeeID, outletID).Updates(updatedData).Error
}

func (r *EmployeesRepo) Delete(employeeID interface{}, outletID interface{}, myRole string) error {
	var employee EmployeeModel
	if err := r.db.Where("id = ? AND outlet_id = ?", employeeID, outletID).First(&employee).Error; err != nil {
		return err
	}

	switch myRole {
	case R_ROOT:
	case R_OWNER: //может удалить только перечисленные роли
		if !employee.hasRole(R_DIRECTOR, R_ADMIN, R_CASHIER) {
			return ErrPermissionDenided
		}
	case R_DIRECTOR: //может удалить только перечисленные роли
		if !employee.hasRole(R_ADMIN, R_CASHIER) {
			return ErrPermissionDenided
		}
	case R_ADMIN: //может удалить только перечисленные роли
		if !employee.hasRole(R_CASHIER) {
			return ErrPermissionDenided
		}
	default:
		return ErrPermissionDenided
	}

	return r.db.Where("id = ? AND outlet_id = ?", employeeID, outletID).Delete(&EmployeeModel{}).Error
}

func (r *EmployeesRepo) FindAllByOrgID(orgID interface{}, whereOutletID uint) (employees []EmployeeModel, err error) {
	if whereOutletID == 0 {
		err = r.db.Where("org_id = ?", orgID).Find(&employees).Error
	} else {
		err = r.db.Where("org_id = ? AND outlet_id = ?", orgID, whereOutletID).Find(&employees).Error
	}
	return
}

func (r *EmployeesRepo) SetOnline(employeeID interface{}) error {
	return r.db.Model(&EmployeeModel{}).Where("id = ?", employeeID).UpdateColumn("online", true).Error
}

func (r *EmployeesRepo) SetOffline(employeeID interface{}) error {
	return r.db.Model(&EmployeeModel{}).Where("id = ?", employeeID).UpdateColumn("online", false).Error
}

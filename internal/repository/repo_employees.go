package repository

import (
	"strconv"

	"gorm.io/gorm"
)

type EmployeeModel struct {
	gorm.Model

	Name     string
	Password string
	Role     string
	Online   bool

	OrgID    uint
	OutletID uint

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
func (m *EmployeeModel) HasRole(roles ...string) bool {
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

//actual
func (r *EmployeesRepo) SignIn(id uint, password string, orgID uint) (empl EmployeeModel, err error) {
	err = r.db.Where("id = ? AND org_id = ? AND password = ?", id, orgID, password).First(&empl).Error
	return
}

func (r *EmployeesRepo) Create(model *EmployeeModel) (err error) {
	if !model.passwordValidation() {
		return ErrOnlyNumCanBeInPassword
	}
	return r.db.Create(model).Error
}

func (r *EmployeesRepo) Updates(updatedFields *EmployeeModel, where *EmployeeModel) error {
	if updatedFields.Password != "" && !updatedFields.passwordValidation() {
		return ErrOnlyNumCanBeInPassword
	}
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *EmployeesRepo) Find(where *EmployeeModel) (result *[]EmployeeModel, err error) {
	err = r.db.Model(&EmployeeModel{}).Where(where).Find(&result).Error
	return
}

func (r *EmployeesRepo) FindFirst(where *EmployeeModel) (result *EmployeeModel, err error) {
	err = r.db.Model(&EmployeeModel{}).Where(where).First(&result).Error
	return
}

func (r *EmployeesRepo) Delete(where *EmployeeModel) (err error) {
	err = r.db.Where(where).Delete(&EmployeeModel{}).Error
	return
}

func (r *EmployeesRepo) Count(where *EmployeeModel) (n int64, err error) {
	err = r.db.Model(&EmployeeModel{}).Where(where).Count(&n).Error
	return
}

func (r *EmployeesRepo) SetOnline(employeeID interface{}) error {
	return r.db.Model(&EmployeeModel{}).Where("id = ?", employeeID).UpdateColumn("online", true).Error
}

func (r *EmployeesRepo) SetOffline(employeeID interface{}) error {
	return r.db.Model(&EmployeeModel{}).Where("id = ?", employeeID).UpdateColumn("online", false).Error
}

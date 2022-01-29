package myservice

import (
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
	"github.com/iivkis/pos-ninja-backend/pkg/mailagent"
	"github.com/iivkis/strcode"
)

type MyService struct {
	Authorization AuthorizationService
	Employees     EmployeesService
	Outlets       OutletsService
	Session       SessionService
	Category      CategoryService
}

func NewMyService(repo repository.Repository, strcode *strcode.Strcode, mailagent *mailagent.MailAgent, authjwt authjwt.AuthJWT) MyService {
	return MyService{
		Authorization: newAuthorizationService(repo, strcode, mailagent, authjwt),
		Employees:     newEmployeesService(repo),
		Outlets:       newOutletsService(repo),
		Session:       newSessionService(repo),
		Category:      newCategoryService(repo),
	}
}

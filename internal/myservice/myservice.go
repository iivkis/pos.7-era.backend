package myservice

import (
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"github.com/iivkis/pos-ninja-backend/pkg/authjwt"
	"github.com/iivkis/pos-ninja-backend/pkg/mailagent"
	"github.com/iivkis/strcode"
)

type MyService struct {
	Authorization           *AuthorizationService
	Employees               *EmployeesService
	Outlets                 *OutletsService
	Sessions                *SessionsService
	Categories              *CategoriesService
	Products                *ProductsService
	Ingredients             *IngredientsService
	OrdersList              *OrdersListService
	OrdersInfo              *OrdersInfoService
	ProductsWithIngredients *ProductsWithIngredientsService
}

func NewMyService(repo *repository.Repository, strcode *strcode.Strcode, mailagent *mailagent.MailAgent, authjwt *authjwt.AuthJWT) MyService {
	return MyService{
		Authorization:           newAuthorizationService(repo, strcode, mailagent, authjwt),
		Employees:               newEmployeesService(repo),
		Outlets:                 newOutletsService(repo),
		Sessions:                newSessionsService(repo),
		Categories:              newCategoriesService(repo),
		Products:                newProductsService(repo),
		Ingredients:             newIngredientsService(repo),
		OrdersList:              newOrderListService(repo),
		OrdersInfo:              newOrdersInfoService(repo),
		ProductsWithIngredients: newProductsWithIngredientsService(repo),
	}
}

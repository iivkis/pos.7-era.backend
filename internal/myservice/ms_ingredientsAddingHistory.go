package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type IngredientsAddingHistoryOutputModel struct {
	ID uint

	Count  float64 `json:"count"`  //кол-во продукта, который не сходится
	Total  float64 `json:"total"`  //сумма, на которую не сходится
	Status int     `json:"status"` // 1 - инвенторизация

	Date int64 `json:"date"` //unixmilli

	IngredientID uint `json:"ingredient_id"`
	EmployeeID   uint `json:"employee_id"` //сотрудник, который делал инветаризацию
	OutletID     uint `json:"outlet_id"`
}

type IngredientsAddingHistoryService struct {
	repo *repository.Repository
}

func newIngredientsAddingHistoryService(repo *repository.Repository) *IngredientsAddingHistoryService {
	return &IngredientsAddingHistoryService{
		repo: repo,
	}
}

type IngredientsAddingHistoryCreateInput struct {
	Count  float64 `json:"count"`                        //кол-во продукта, который не сходится
	Total  float64 `json:"total"`                        //сумма, на которую не сходится
	Status int     `json:"status" binding:"min=1,max=2"` // 1 - инвенторизация

	Date int64 `json:"date"` //unixmilli

	IngredientID uint `json:"ingredient_id" binding:"min=1"`
}

//@Summary Добавить новый отчёт об ингредиентах
//@param type body IngredientsAddingHistoryCreateInput false "Принимаемый объект"
//@Accept json
//@Produce json
//@Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /ingredients.History [post]
func (s *IngredientsAddingHistoryService) Create(c *gin.Context) {
	var input IngredientsAddingHistoryCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	model := repository.IngredientsAddingHistoryModel{
		Count:        input.Count,
		Total:        input.Total,
		Status:       input.Status,
		IngredientID: input.IngredientID,
		Date:         input.Date,
		EmployeeID:   claims.EmployeeID,
		OutletID:     claims.OutletID,
		OrgID:        claims.OrganizationID,
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		if stdQuery.OutletID != 0 && s.repo.Outlets.ExistsInOrg(stdQuery.OutletID, claims.OrganizationID) {
			model.OutletID = stdQuery.OutletID
		}
	}

	if !s.repo.Ingredients.Exists(&repository.IngredientModel{ID: model.IngredientID, OutletID: model.OutletID}) {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData("undefined `ingredient_id` in outlet with this id"))
		return
	}

	if err := s.repo.IngredientsAddingHistory.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type IngredientsAddingHistorytGetAllInput struct {
	Start uint64 `form:"start"` //in unixmilli
	End   uint64 `form:"end"`   //in unixmilli
}
type IngredientsAddingHistorytGetAllOutput []IngredientsAddingHistoryOutputModel

//@Summary Получить историю добавления ингредиентов
//@Accept json
//@Produce json
//@Success 200 {object} IngredientsAddingHistorytGetAllOutput "возвращаемый объект"
//@Failure 400 {object} serviceError
//@Router /ingredients.History [get]
func (s *IngredientsAddingHistoryService) GetAll(c *gin.Context) {
	var query IngredientsAddingHistorytGetAllInput
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}
	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

	where := &repository.IngredientsAddingHistoryModel{
		OrgID:    claims.OrganizationID,
		OutletID: claims.OutletID,
	}

	if claims.HasRole(repository.R_OWNER) {
		if stdQuery.OrgID != 0 && s.repo.Invitation.Exists(&repository.InvitationModel{OrgID: claims.OrganizationID, AffiliateOrgID: stdQuery.OrgID}) {
			where.OrgID = stdQuery.OrgID
		}
	}

	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
		where.OutletID = stdQuery.OutletID
	}

	histories, err := s.repo.IngredientsAddingHistory.FindWithPeriod(where, query.Start, query.End)
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	var output = make(IngredientsAddingHistorytGetAllOutput, len(*histories))
	for i, item := range *histories {
		output[i] = IngredientsAddingHistoryOutputModel{
			ID:           item.ID,
			Count:        item.Count,
			Total:        item.Total,
			Status:       item.Status,
			Date:         item.Date,
			IngredientID: item.IngredientID,
			EmployeeID:   item.EmployeeID,
			OutletID:     item.OutletID,
		}
	}
	NewResponse(c, http.StatusOK, output)
}

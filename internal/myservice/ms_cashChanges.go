package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type CashChangesService struct {
	repo *repository.Repository
}

func newCashChangesService(repo *repository.Repository) *CashChangesService {
	return &CashChangesService{
		repo: repo,
	}
}

type CashChangesCreateInput struct {
	Date      int64   `json:"date"` //unixmilli
	Total     float64 `json:"total"`
	Reason    int     `json:"reason" binding:"min=0,max=2"`
	Comment   string  `json:"comment"`
	SessionID uint    `json:"session_id"`
}

//@Summary Добавить информацию о снятии\вкладе денежных средств в кассу
//@param type body CashChangesCreateInput false "Принимаемый объект"
//@Success 201 {object} DefaultOutputModel "возвращает id созданной записи"
//@Accept json
//@Produce json
//@Failure 400 {object} serviceError
//@Failure 500 {object} serviceError
//@Router /products [post]
func (s *CashChangesService) Create(c *gin.Context) {
	var input CashChangesCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	m := repository.CashChangesModel{
		Date:       input.Date,
		Total:      input.Total,
		Reason:     input.Reason,
		Comment:    input.Comment,
		SessionID:  input.SessionID,
		EmployeeID: c.MustGet("claims_employee_id").(uint),
		OutletID:   c.MustGet("claims_outlet_id").(uint),
		OrgID:      c.MustGet("claims_org_id").(uint),
	}

	if err := s.repo.CashChanges.Create(&m); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, DefaultOutputModel{ID: m.ID})
}

type CashChangesGetAllForOutletInput struct {
	Start int64 `json:"start"` //in unixmilli
	End   int64 `json:"end"`   //in unixmilli
}

func GetAllForOutlet(c *gin.Context) {
	var query CashChangesGetAllForOutletInput
	if err := c.ShouldBindQuery(&query); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

}

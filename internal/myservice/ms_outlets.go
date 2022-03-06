package myservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type OutletsService struct {
	repo *repository.Repository
}

type outletOutputModel struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func newOutletsService(repo *repository.Repository) *OutletsService {
	return &OutletsService{
		repo: repo,
	}
}

type OutletCreateInput struct {
	Name string `json:"name" binding:"required,max=100"`
}

//@Summary Добавить торговую точку (токен юзера)
//@Description Метод позволяет добавить торговую точку
//@Param json body OutletCreateInput true "Объект для добавления торговой точки."
//@Accept json
//@Produce json
//@Success 200 {object} DefaultOutputModel "возвращает id созданной записи"
//@Failure 500 {object} serviceError
//@Router /outlets [post]
func (s *OutletsService) Create(c *gin.Context) {
	var input OutletCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	model := repository.OutletModel{
		Name:  input.Name,
		OrgID: c.MustGet("claims_org_id").(uint),
	}

	if err := s.repo.Outlets.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	emplModel := repository.EmployeeModel{
		Name:     "Администратор",
		Password: "000000",
		Role:     repository.R_ADMIN,
		OutletID: model.ID,
		OrgID:    c.MustGet("claims_org_id").(uint),
	}

	if err := s.repo.Employees.Create(&emplModel, repository.R_OWNER); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type OutletGetAllForOrgOutput []outletOutputModel

//@Summary Список всех торговых точек (токен организации)
//@Description Метод позволяет получить список всех торговых точек
//@Produce json
//@Success 200 {object} OutletGetAllForOrgOutput "Возвращает массив торговых точек"
//@Failure 500 {object} serviceError
//@Router /outlets [get]
func (s *OutletsService) GetAllForOrg(c *gin.Context) {
	outlets, err := s.repo.Outlets.FindAllByOrgID(c.MustGet("claims_org_id"))
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownServer(err.Error()))
		return
	}

	output := make(OutletGetAllForOrgOutput, len(outlets))
	for i, outlet := range outlets {
		output[i] = outletOutputModel{
			ID:   outlet.ID,
			Name: outlet.Name,
		}
	}
	NewResponse(c, http.StatusOK, output)
}

type OutletUpdateFieldsInput struct {
	Name string `json:"name"`
}

//@Summary Обновить точку (токен юзера)
//@Param json body OutletCreateInput false "Обновляемые поля"
//@Accept json
//@Produce json
//@Success 200 {object} object "возвращает пустой объект"
//@Failure 500 {object} serviceError
//@Router /outlets/:id [put]
func (s *OutletsService) UpdateFields(c *gin.Context) {
	var input OutletUpdateFieldsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	m := repository.OutletModel{
		Name: input.Name,
	}

	if !s.repo.Outlets.ExistsInOrg(c.Param("id"), c.MustGet("claims_org_id")) {
		NewResponse(c, http.StatusBadRequest, errRecordNotFound("undefined outlet with this `id` in your organization"))
		return
	}

	if err := s.repo.Outlets.Updates(&m, c.Param("id")); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

//@Summary Удалить точку (токен юзера)
//@Accept json
//@Produce json
//@Success 200 {object} object "возвращает пустой объект"
//@Failure 500 {object} serviceError
//@Router /outlets/:id [delete]
func (s *OutletsService) Delete(c *gin.Context) {
	if !s.repo.Outlets.ExistsInOrg(c.Param("id"), c.MustGet("claims_org_id")) {
		NewResponse(c, http.StatusBadRequest, errRecordNotFound("undefined outlet with this `id` in your organization"))
		return
	}

	if err := s.repo.Outlets.Delete(c.MustGet("claims_outlet_id")); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}

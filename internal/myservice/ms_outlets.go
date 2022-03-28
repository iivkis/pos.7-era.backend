package myservice

import (
	"net/http"
	"strconv"

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

	claims := mustGetEmployeeClaims(c)

	model := repository.OutletModel{
		Name:  input.Name,
		OrgID: claims.OrganizationID,
	}

	if err := s.repo.Outlets.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	employeeModel := repository.EmployeeModel{
		Name:     "Администратор",
		Password: "000000",
		Role:     repository.R_ADMIN,
		OutletID: model.ID,
		OrgID:    claims.OrganizationID,
	}

	if err := s.repo.Employees.Create(&employeeModel); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type OutletGetAllOutput []outletOutputModel

//@Summary Список всех торговых точек (токен организации)
//@Description Метод позволяет получить список всех торговых точек
//@Produce json
//@Success 200 {object} OutletGetAllOutput "Возвращает массив торговых точек"
//@Failure 500 {object} serviceError
//@Router /outlets [get]
func (s *OutletsService) GetAllForOrg(c *gin.Context) {
	claims := mustGetOrganizationClaims(c)
	outlets, err := s.repo.Outlets.Find(&repository.OutletModel{OrgID: claims.OrganizationID})
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}

	output := make(OutletGetAllOutput, len(*outlets))
	for i, outlet := range *outlets {
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

	claims := mustGetEmployeeClaims(c)

	updatedFields := &repository.OutletModel{
		Name: input.Name,
	}

	outletID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	if !s.repo.Outlets.ExistsInOrg(uint(outletID), claims.OrganizationID) {
		NewResponse(c, http.StatusBadRequest, errRecordNotFound("undefined outlet with this `id` in your organization"))
		return
	}

	if err := s.repo.Outlets.Updates(&repository.OutletModel{ID: uint(outletID), OrgID: claims.OrganizationID}, updatedFields); err != nil {
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
	claims := mustGetEmployeeClaims(c)

	outletID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}
	if !s.repo.Outlets.ExistsInOrg(uint(outletID), claims.OrganizationID) {
		NewResponse(c, http.StatusBadRequest, errRecordNotFound("undefined outlet with this `id` in your organization"))
		return
	}

	if err := s.repo.Outlets.Delete(&repository.OutletModel{ID: uint(outletID), OrgID: claims.OrganizationID}); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}

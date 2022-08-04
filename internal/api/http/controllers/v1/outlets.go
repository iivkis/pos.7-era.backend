package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type outletResponseModel struct {
	ID   uint   `json:"id" mapstructure:"id"`
	Name string `json:"name" mapstructure:"name"`
}

type outlets struct {
	repo *repository.Repository
}

func newOutlets(repo *repository.Repository) *outlets {
	return &outlets{
		repo: repo,
	}
}

type outletsCreateBody struct {
	Name string `json:"name" binding:"required,max=100"`
}

// @Summary Добавить торговую точку (токен юзера)
// @Description Метод позволяет добавить торговую точку
// @Param json body outletsCreateBody true "object"
// @Success 201 {object} DefaultOutputModel "возвращает id новой торговой точки"
// @Router /outlets [post]
func (s *outlets) Create(c *gin.Context) {
	var body outletsCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	model := repository.OutletModel{
		Name:  body.Name,
		OrgID: claims.OrganizationID,
	}

	if err := s.repo.Outlets.Create(&model); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
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
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusCreated, DefaultOutputModel{ID: model.ID})
}

type outletsGetAllResponse []outletResponseModel

// @Summary Список всех торговых точек (токен организации)
// @Description Метод позволяет получить список всех торговых точек
// @Produce json
// @Success 200 {object} outletsGetAllResponse "Возвращает массив торговых точек"
// @Failure 500 {object} serviceError
// @Router /outlets [get]
func (s *outlets) GetAll(c *gin.Context) {
	claims := mustGetOrganizationClaims(c)
	outlets, err := s.repo.Outlets.Find(&repository.OutletModel{OrgID: claims.OrganizationID})
	if err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	output := make(outletsGetAllResponse, len(*outlets))
	for i, outlet := range *outlets {
		output[i] = outletResponseModel{
			ID:   outlet.ID,
			Name: outlet.Name,
		}
	}
	NewResponse(c, http.StatusOK, output)
}

type outletsUpdateBody struct {
	Name string `json:"name" mapstructure:"name"`
}

// @Summary Обновить точку (токен юзера)
// @Param json body OutletCreateInput false "Обновляемые поля"
// @Accept json
// @Produce json
// @Success 200 {object} object "возвращает пустой объект"
// @Failure 500 {object} serviceError
// @Router /outlets/:id [put]
func (s *outlets) Update(c *gin.Context) {
	outletID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	var body outletsUpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
		return
	}

	claims := mustGetEmployeeClaims(c)

	updated := &repository.OutletModel{
		Name: body.Name,
	}

	if !s.repo.Outlets.ExistsInOrg(uint(outletID), claims.OrganizationID) {
		NewResponse(c, http.StatusBadRequest, errRecordNotFound("undefined outlet with this `id` in your organization"))
		return
	}

	if err := s.repo.Outlets.Updates(&repository.OutletModel{Model: gorm.Model{ID: uint(outletID)}, OrgID: claims.OrganizationID}, updated); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}

	NewResponse(c, http.StatusOK, nil)
}

// @Summary Удалить точку (токен юзера)
// @Accept json
// @Produce json
// @Success 200 {object} object "возвращает пустой объект"
// @Failure 500 {object} serviceError
// @Router /outlets/:id [delete]
func (s *outlets) Delete(c *gin.Context) {
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

	// ошибка, если владелец привязан к точке
	if s.repo.Employees.Exists(&repository.EmployeeModel{Role: "owner", OutletID: uint(outletID)}) {
		NewResponse(c, http.StatusBadRequest, errPermissionDenided("you don't can delete main outlet"))
		return
	}

	if err := s.repo.Outlets.Delete(&repository.OutletModel{Model: gorm.Model{ID: uint(outletID)}, OrgID: claims.OrganizationID}); err != nil {
		NewResponse(c, http.StatusInternalServerError, errUnknown(err.Error()))
		return
	}
	NewResponse(c, http.StatusOK, nil)
}

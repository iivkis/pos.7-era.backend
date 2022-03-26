package myservice

// import (
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"github.com/iivkis/pos.7-era.backend/internal/repository"
// )

// type ReportRevenueOutputModel struct {
// 	ID uint `json:"id"`

// 	BankEarned  float64 `json:"bank_earned"`
// 	CashEarned  float64 `json:"cash_earned"`
// 	TotalAmount float64 `json:"total_amount"`

// 	NumberOfReceipts int     `json:"number_of_receipts"`
// 	AverageReceipt   float64 `json:"average_receipt"`

// 	Date int64 `json:"date"` // (in unixmilli) за какое число отчёт

// 	OutletID uint
// }

// type ReportRevenueService struct {
// 	repo *repository.Repository
// }

// func newReportRevenueService(repo *repository.Repository) *ReportRevenueService {
// 	return &ReportRevenueService{
// 		repo: repo,
// 	}
// }

// type ReportRevenueGetAllQuery struct {
// 	Start uint64 `form:"start"` //in unixmilli
// 	End   uint64 `form:"end"`   //in unixmilli
// }

// type ReportRevenueGetAllOutput []ReportRevenueOutputModel

// //@Summary Отчёты по дневной выручке всех точек
// //@param type query ReportRevenueGetAllQuery false "Принимаемый объект"
// //@Accept json
// //@Produce json
// //@Success 200 {object} ReportRevenueOutputModel "возвращаемый объект"
// //@Failure 400 {object} serviceError
// //@Router /report.Revenue [get]
// func (s *ReportRevenueService) GetAll(c *gin.Context) {
// 	var query InventoryHistoryGetAllQuery
// 	if err := c.ShouldBindQuery(&query); err != nil {
// 		NewResponse(c, http.StatusBadRequest, errIncorrectInputData(err.Error()))
// 		return
// 	}

// 	claims, stdQuery := mustGetEmployeeClaims(c), mustGetStdQuery(c)

// 	where := &repository.InventoryHistoryModel{
// 		OrgID:    claims.OrganizationID,
// 		OutletID: claims.OutletID,
// 	}

// 	if claims.HasRole(repository.R_OWNER, repository.R_DIRECTOR) {
// 		where.OutletID = stdQuery.OutletID
// 	}

// 	reports, err := s.repo.ReportRevenue.FindWithPeriod(where, query.Start, query.End)
// 	if err != nil {
// 		NewResponse(c, http.StatusInternalServerError, errUnknownDatabase(err.Error()))
// 		return
// 	}

// 	var output = make(ReportRevenueGetAllOutput, len(*reports))
// 	for i, item := range *reports {
// 		output[i] = ReportRevenueOutputModel{
// 			ID:          item.ID,
// 			BankEarned:  item.BankEarned,
// 			CashEarned:  item.CashEarned,
// 			TotalAmount: item.TotalAmount,
// 			Date:        item.Date,
// 			OutletID:    item.OutletID,
// 		}
// 	}
// 	NewResponse(c, http.StatusOK, output)
// }

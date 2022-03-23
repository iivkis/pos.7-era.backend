package autoreport

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/iivkis/pos.7-era.backend/internal/repository"
	"gorm.io/gorm"
)

type AutoReport struct {
	repo   *repository.Repository
	errlog *log.Logger
}

func NewAutoReport(repo *repository.Repository) *AutoReport {
	errlogFile, err := os.OpenFile("./autoreport.log", os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}

	return &AutoReport{
		repo:   repo,
		errlog: log.New(errlogFile, "", 0),
	}
}

func (a *AutoReport) Run() {
	var lastDay int
	go func() {
		for {
			// МСК +3 GTM.
			if time.Now().UTC().Day() != lastDay {
				lastDay = time.Now().UTC().Day()
				a.createReport()
				time.Sleep(time.Hour * 23)
			}
			time.Sleep(time.Minute * 30)
		}
	}()
}

func (a *AutoReport) createReport() {
	//поиск точек, где есть новые сессии
	outletIDs, err := a.repo.Sessions.FindOutletIDForReport(&repository.SessionModel{})

	fmt.Println("Create report for outlets (", len(*outletIDs), ")")

	if err != nil {
		a.errlog.Println(err)
	}

	//для каждой точки берем новые сессии
	for _, outletID := range *outletIDs {
		sessions, err := a.repo.Sessions.FindSessionsForReport(&repository.SessionModel{OutletID: outletID})
		if err != nil {
			a.errlog.Println(err)
			continue
		}

		sessionIDs := make([]uint, len(*sessions))

		newReport := &repository.ReportRevenueModel{OutletID: outletID}
		for i, sess := range *sessions {
			sessionIDs[i] = sess.ID

			newReport.BankEarned += sess.BankEarned
			newReport.CashEarned += sess.CashEarned

			newReport.Date = sess.DateClose
			newReport.OrgID = sess.OrgID
		}

		//общая сумама
		newReport.TotalAmount = newReport.BankEarned + newReport.CashEarned

		//округляем дату до дня
		date := time.UnixMilli(newReport.Date)
		newReport.Date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()

		report, err := a.repo.ReportRevenue.FindFirts(&repository.ReportRevenueModel{Date: newReport.Date, OutletID: outletID})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := a.repo.ReportRevenue.Create(newReport); err != nil {
					a.errlog.Println(err.Error())
					continue
				}
			} else {
				a.errlog.Println(err.Error())
				continue
			}
		} else {
			report.BankEarned += newReport.BankEarned
			report.CashEarned += newReport.CashEarned
			report.TotalAmount = report.BankEarned + report.CashEarned

			if err := a.repo.ReportRevenue.Updates(&repository.ReportRevenueModel{Model: gorm.Model{ID: report.ID}}, report); err != nil {
				a.errlog.Println(err.Error())
				continue
			}
		}

		if err := a.repo.Sessions.SetFieldAddedToReport(true, sessionIDs); err != nil {
			a.errlog.Println(err.Error())
		}
	}
}

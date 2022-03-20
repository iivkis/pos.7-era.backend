package autoreport

import (
	"errors"
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
	errlogFile, err := os.OpenFile("./log/autoreport.log", os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}

	return &AutoReport{
		repo:   repo,
		errlog: log.New(errlogFile, "", 0),
	}
}

//1) Каждый день в 00:00 по МСК запускаем отчет по выручке (cashEarn и bankEarn записываем в соотвествующие поля).
//2) Ищем id точек, в которых есть новые зыкрытые сессии (added_to_report = false AND date_close != 0)
//3) запрашиваем новые закрытые сессии для каждой точки. Суммируем соответсвующие поля, задаем дату отчёта прошлым днём.
func (a *AutoReport) Run() {
	var lastDay int
	go func() {
		for {
			if time.Now().UTC().Hour() >= 3 && time.Now().UTC().Day() != lastDay {
				lastDay = time.Now().UTC().Day()
				a.createReport()
				time.Sleep(time.Hour * 23)
			}
			time.Sleep(time.Minute * 30)
		}
	}()
}

func (a *AutoReport) createReport() {
	outletIDs, err := a.repo.Sessions.FindOutletIDForReport(&repository.SessionModel{})
	if err != nil {
		a.errlog.Println(err)
	}

	for _, outletID := range *outletIDs {
		sessions, err := a.repo.Sessions.FindSessionsForReport(&repository.SessionModel{OutletID: outletID})
		if err != nil {
			a.errlog.Println(err)
			continue
		}

		sessionIDs := make([]uint, len(*sessions))

		newRreport := &repository.ReportRevenueModel{OutletID: outletID}
		for i, sess := range *sessions {
			sessionIDs[i] = sess.ID

			newRreport.BankEarned += sess.BankEarned
			newRreport.CashEarned += sess.CashEarned
			newRreport.Date = sess.DateClose
		}

		date := time.UnixMilli(newRreport.Date)
		newRreport.Date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()

		report, err := a.repo.ReportRevenue.FindFirts(&repository.ReportRevenueModel{Date: newRreport.Date})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := a.repo.ReportRevenue.Create(newRreport); err != nil {
					a.errlog.Println(err.Error())
					continue
				}
			} else {
				a.errlog.Println(err.Error())
				continue
			}
		} else {
			report.BankEarned += newRreport.BankEarned
			report.CashEarned += newRreport.CashEarned

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

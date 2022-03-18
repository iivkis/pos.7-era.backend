package autoreport

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/iivkis/pos.7-era.backend/internal/repository"
)

type AutoReport struct {
	repo   *repository.Repository
	errlog *log.Logger
}

func NewAutoReport(repo *repository.Repository) *AutoReport {
	errlogFile, err := os.OpenFile("./log/autoreport.log", os.O_CREATE|os.O_WRONLY, 777)
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
	fmt.Println(outletIDs, err)

	for _, outletID := range *outletIDs {
		sessions, err := a.repo.Sessions.FindSessionsForReport(&repository.SessionModel{OutletID: outletID})
		if err != nil {
			a.errlog.Println(err)
		}

		report := &repository.ReportRevenueModel{}
		for _, sess := range *sessions {
			report.BankEarned += sess.BankEarned
			report.CashEarned += sess.CashEarned
			report.Date = sess.DateClose
		}
	}
}

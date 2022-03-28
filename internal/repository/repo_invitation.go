package repository

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/iivkis/pos.7-era.backend/internal/config"
	"gorm.io/gorm"
)

type InvitationModel struct {
	ID uint

	Code      string `gorm:"index"`     // рандомный код
	ExpiresIn int64  `gorm:"default:0"` // истекает в (unixmilli)

	OrgID          uint // организация, которой принадлежит данный код
	AffiliateOrgID uint `gorm:"default:NULL"` // приглашенная организация

	OrganizationModel          OrganizationModel `gorm:"foreignKey:OrgID"`
	OrganizationAffiliateModel OrganizationModel `gorm:"foreignKey:AffiliateOrgID"`
}

type InvitationRepo struct {
	db       *gorm.DB
	rand     *rand.Rand
	alphabet []byte
}

func newInvitationRepo(db *gorm.DB) *InvitationRepo {
	r := &InvitationRepo{
		db:       db,
		alphabet: []byte("abcdefqxyzkrABCDEFQXYZKR1234567890"),
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if *config.Flags.Main {
		go func() {
			for {
				r.DeleteExpired()
				time.Sleep(time.Hour)
			}
		}()
	}

	return r
}

func (r *InvitationRepo) generateRandomString(length int) string {
	bstr := make([]byte, length)
	for i := range bstr {
		bstr[i] = r.alphabet[r.rand.Intn(len(r.alphabet))]
	}
	return string(bstr)
}

func (r *InvitationRepo) Create(m *InvitationModel) error {
	m.ExpiresIn = time.Now().Add(time.Hour * 24).UnixMilli()

	for {
		m.Code = r.generateRandomString(9)
		if _, err := r.FindFirts(&InvitationModel{Code: m.Code}); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			} else {
				return err
			}
		}
	}

	return r.db.Create(m).Error
}

func (r *InvitationRepo) Find(where *InvitationModel) (result *[]InvitationModel, err error) {
	err = r.db.Where("expires_in > ? OR expires_in is NULL", time.Now().UTC().UnixMilli()).Find(&result, where).Error
	return
}

func (r *InvitationRepo) FindActivated(where *InvitationModel) (result *[]InvitationModel, err error) {
	err = r.db.Where("affiliate_org_id IS NOT NULL").Find(&result, where).Error
	return
}

func (r *InvitationRepo) FindNotActivated(where *InvitationModel) (result *[]InvitationModel, err error) {
	err = r.db.Where("expires_in > ? AND affiliate_org_id is NULL", time.Now().UTC().UnixMilli()).Find(&result, where).Error
	return
}

func (r *InvitationRepo) FindFirts(where *InvitationModel) (result *InvitationModel, err error) {
	err = r.db.Where("expires_in > ? OR expires_in is NULL", time.Now().UTC().UnixMilli()).First(&result, where).Error
	return
}

func (r *InvitationRepo) Updates(where *InvitationModel, updatedFields *InvitationModel) error {
	return r.db.Where(where).Updates(updatedFields).Error
}

func (r *InvitationRepo) Delete(where *InvitationModel) (err error) {
	err = r.db.Where(where).Delete(&InvitationModel{}).Error
	return
}

func (r *InvitationRepo) DeleteExpired() (err error) {
	err = r.db.Where("expires_in <= ? AND expires_in IS NOT NULL", time.Now().UTC().UnixMilli()).Delete(&InvitationModel{}).Error
	return
}

func (r *InvitationRepo) Exists(where *InvitationModel) bool {
	return r.db.Where("expires_in > ? OR expires_in IS NULL", time.Now().UTC().UnixMilli()).First(&InvitationModel{}, where).Error == nil
}

func (r *InvitationRepo) Activate(code string, AffiliateOrgID uint) error {
	invite, err := r.FindFirts(&InvitationModel{Code: code})
	if err != nil {
		return err
	}

	fmt.Println(invite)

	r.db.Model(&InvitationModel{}).Where(invite.ID).Updates(map[string]interface{}{
		"code":             nil,
		"expires_in":       nil,
		"affiliate_org_id": AffiliateOrgID,
	})

	return nil
}

func (r *InvitationRepo) CountNotActivated(ordID interface{}) (n int64, err error) {
	err = r.db.Model(&InvitationModel{}).Where("expires_in IS NOT NULL AND org_id = ?", ordID).Count(&n).Error
	return
}

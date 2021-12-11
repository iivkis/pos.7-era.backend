package myservice

import (
	"github.com/iivkis/pos-ninja-backend/internal/repository"
	"github.com/iivkis/pos-ninja-backend/pkg/mailagent"
	"github.com/iivkis/strcode"
)

type MyService struct {
	Authorization AuthorizationService
}

func NewMyService(repo repository.Repository, strcode *strcode.Strcode, mailagent *mailagent.MailAgent) MyService {
	return MyService{
		Authorization: newAuthorizationService(repo, strcode, mailagent),
	}
}

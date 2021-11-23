package myservice

import "github.com/iivkis/pos-ninja-backend/internal/repository"

type MyService struct {
	Authorization AuthorizationService
}

func NewMyService(repo repository.Repository) MyService {
	return MyService{
		Authorization: newAuthorizationService(repo),
	}
}

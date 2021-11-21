package myservice

import "github.com/iivkis/pos-ninja-backend/internal/repository"

type MyService struct {
	Organizations iOrganizations
}

func NewMyService(repo repository.Repository) MyService {
	return MyService{
		Organizations: newOrganizations(repo),
	}
}

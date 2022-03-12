package repository

import "errors"

var (
	ErrOnlyNumCanBeInPassword = errors.New("only numbers can be used in an employee's password")
	ErrSessionAlreadyOpen     = errors.New("this user already has a covered session")
)

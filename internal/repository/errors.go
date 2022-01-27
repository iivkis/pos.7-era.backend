package repository

import "errors"

var (
	ErrUndefinedRole     = errors.New("undefined role")
	ErrOnlyNumInPassword = errors.New("only numbers can be used in an employee's password")
)

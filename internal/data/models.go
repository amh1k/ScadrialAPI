package data

import (
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Movies      MovieModelInterface
	Users       UserModelInterface
	Tokens      TokenModel
	Permissions PermissionModel
}

func NewModels(movies MovieModelInterface, users UserModelInterface, tokens TokenModel, permission PermissionModel) Models {
	return Models{
		Movies:      movies,
		Users:       users,
		Tokens:      tokens,
		Permissions: permission,
	}
}

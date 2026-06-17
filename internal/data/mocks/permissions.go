package mocks

import "scadrialapi.abdulmoiz.net/internal/data"

type PermissionModel struct {
	GetAllForUserFn func(userID int64) (data.Permissions, error)
	AddForUserFn func(userID int64, codes ...string) error
}
func(m *PermissionModel) AddForUser(userID int64, codes ...string) error  {
	return m.AddForUserFn(userID, codes...)
}
func(m *PermissionModel)GetAllForUser(userID int64) (data.Permissions, error) {
	return m.GetAllForUserFn(userID)	
}


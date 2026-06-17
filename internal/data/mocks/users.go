package mocks

import (
	"time"

	"scadrialapi.abdulmoiz.net/internal/data"
)
type UserModel struct {}


func(m *UserModel)GetForToken(tokenScope, tokenPlaintext string)(*data.User, error) {
	var password data.Password
	pointerString := "12345678"
	err := password.Set(pointerString)
	if err != nil {
		return nil, err
	}
	userMock := data.User{
		ID: 1,     
		CreatedAt :time.Now(), 
		Name:     "abc",
		Email:     "abc@gmail.com",   
		Password: password  ,
		Activated: true,      
		Version:   1 ,      


	}
	return &userMock, nil
}
func (m *UserModel)GetByEmail(email string) (*data.User, error) {
	return &data.User{}, nil
}
func (m *UserModel)Insert(*data.User) error {
	return nil
}


func(m *UserModel)Update(*data.User) error {
	return nil
}

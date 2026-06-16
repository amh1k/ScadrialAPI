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
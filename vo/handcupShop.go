package vos

import (
	"encoding/json"
	"time"
)

// User ...
type User struct {
	ID                   int            `json:"id"`
	CreateDate           time.Time      `json:"created"`
	UpdateDate           time.Time      `json:"updated"`
	Email                string         `json:"email" validate:"omitempty,email"`
	Name                 string         `json:"name"`
	Password             string         `json:"password"`
	Token                string         `json:"token"`
	ValidCode            string         `json:"validCode"`
	ValidCodeExpiredTime time.Time      `json:"validCodeExpiredTime"`
	Status               string         `json:"status"`
	Type                 string         `json:"type" validate:"omitempty,UserTypeConstFunc"`
	Avator               string         `json:"avator"`
	TokenExpiredTime     time.Time      `json:"tokenExpiredTime"`
	UserDetail           UserDetail     `json:"userDetail" validate:"omitempty,dive"`
	UserThirdparty       UserThirdparty `json:"userThirdparty" validate:"omitempty,dive"`
}
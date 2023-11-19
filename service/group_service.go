package service

import (
	req "github.com/elabosak233/pgshub/model/request/account"
	"github.com/elabosak233/pgshub/model/response"
)

type GroupService interface {
	Create(req req.CreateGroupRequest)
	Update(req req.UpdateGroupRequest)
	Delete(id string)
	FindById(id string) response.GroupResponse
	FindAll() []response.GroupResponse
	AddUserToGroup(req req.AddUserToGroupRequest)
}
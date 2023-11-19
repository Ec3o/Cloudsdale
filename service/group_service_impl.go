package service

import (
	model "github.com/elabosak233/pgshub/model/data"
	req "github.com/elabosak233/pgshub/model/request/account"
	"github.com/elabosak233/pgshub/model/response"
	"github.com/elabosak233/pgshub/repository"
	"github.com/elabosak233/pgshub/utils"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type GroupServiceImpl struct {
	GroupRepository repository.GroupRepository
	UserRepository  repository.UserRepository
	Validate        *validator.Validate
}

func NewGroupServiceImpl(appRepository repository.AppRepository) GroupService {
	return &GroupServiceImpl{
		GroupRepository: appRepository.GroupRepository,
		UserRepository:  appRepository.UserRepository,
		Validate:        validator.New(),
	}
}

// Create implements UserService
func (t *GroupServiceImpl) Create(req req.CreateGroupRequest) {
	err := t.Validate.Struct(req)
	utils.ErrorPanic(err)
	groupModel := model.Group{
		Id:   uuid.NewString(),
		Name: req.Name,
	}
	t.GroupRepository.Insert(groupModel)
}

// Delete implements UserService
func (t *GroupServiceImpl) Delete(id string) {
	t.GroupRepository.Delete(id)
}

// FindAll implements UserService
func (t *GroupServiceImpl) FindAll() []response.GroupResponse {
	result := t.GroupRepository.FindAll()

	var groups []response.GroupResponse
	for _, value := range result {
		group := response.GroupResponse{
			Id:      value.Id,
			Name:    value.Name,
			UserIds: value.UserIds,
		}
		groups = append(groups, group)
	}

	return groups
}

// FindById implements UserService
func (t *GroupServiceImpl) FindById(id string) response.GroupResponse {
	groupData, err := t.GroupRepository.FindById(id)
	utils.ErrorPanic(err)

	group := response.GroupResponse{
		Id:      groupData.Id,
		Name:    groupData.Name,
		UserIds: groupData.UserIds,
	}
	return group
}

// Update implements UserService
func (t *GroupServiceImpl) Update(req req.UpdateGroupRequest) {
	groupData, err := t.GroupRepository.FindById(req.Id)
	utils.ErrorPanic(err)
	groupData.Name = req.Name
	t.GroupRepository.Update(groupData)
}

func (t *GroupServiceImpl) AddUserToGroup(req req.AddUserToGroupRequest) {
	user, err := t.UserRepository.FindById(req.UserId)
	if err != nil || user.Id == "" {
		utils.ErrorPanic(err)
		return
	}
	group, err := t.GroupRepository.FindById(req.GroupId)
	if err != nil || group.Id == "" {
		utils.ErrorPanic(err)
		return
	}
	if !contains(group.UserIds, user.Id) {
		group.UserIds = append(group.UserIds, user.Id)
		t.GroupRepository.Update(group)
	}

	if !contains(user.GroupIds, group.Id) {
		user.GroupIds = append(user.GroupIds, group.Id)
		t.UserRepository.Update(user)
	}
}

func contains(slice []string, element string) bool {
	for _, e := range slice {
		if e == element {
			return true
		}
	}
	return false
}
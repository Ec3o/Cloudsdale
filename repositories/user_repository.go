package repositories

import (
	model "github.com/elabosak233/pgshub/models/entity"
	"github.com/elabosak233/pgshub/models/request"
	"github.com/elabosak233/pgshub/models/response"
	"github.com/xormplus/xorm"
)

type UserRepository interface {
	Insert(user model.User) error
	Update(user model.User) error
	Delete(id int64) error
	FindById(id int64) (user model.User, err error)
	FindByUsername(username string) (user model.User, err error)
	FindByEmail(email string) (user model.User, err error)
	Find(req request.UserFindRequest) (user []response.UserResponse, count int64, err error)
	BatchFindByTeamId(req request.UserBatchFindByTeamIdRequest) (users []response.UserResponseWithTeamId, err error)
}

type UserRepositoryImpl struct {
	Db *xorm.Engine
}

func NewUserRepositoryImpl(Db *xorm.Engine) UserRepository {
	return &UserRepositoryImpl{Db: Db}
}

// Insert implements UserRepository
func (t *UserRepositoryImpl) Insert(user model.User) error {
	_, err := t.Db.Table("users").Insert(&user)
	return err
}

// Delete implements UserRepository
func (t *UserRepositoryImpl) Delete(id int64) error {
	_, err := t.Db.Table("users").ID(id).Delete(&model.User{})
	return err
}

func (t *UserRepositoryImpl) Update(user model.User) error {
	_, err := t.Db.Table("users").ID(user.UserId).Update(&user)
	return err
}

func (t *UserRepositoryImpl) Find(req request.UserFindRequest) (users []response.UserResponse, count int64, err error) {
	applyFilter := func(q *xorm.Session) *xorm.Session {
		if req.UserId != 0 {
			q = q.Where("id = ?", req.UserId)
		}
		if req.Email != "" {
			q = q.Where("email LIKE ?", "%"+req.Email+"%")
		}
		if req.Name != "" {
			q = q.Where("name LIKE ? OR username LIKE ?", "%"+req.Name+"%", "%"+req.Name+"%")
		}
		if req.Role != 0 {
			q = q.Where("role = ?", req.Role)
		}
		return q
	}
	db := applyFilter(t.Db.Table("users"))
	ct := applyFilter(t.Db.Table("users"))
	count, err = ct.Count(&model.User{})
	if len(req.SortBy) > 0 {
		sortKey := req.SortBy[0]
		sortOrder := req.SortBy[1]
		if sortOrder == "asc" {
			db = db.Asc("users." + sortKey)
		} else if sortOrder == "desc" {
			db = db.Desc("users." + sortKey)
		}
	}
	if req.Page != 0 && req.Size > 0 {
		offset := (req.Page - 1) * req.Size
		db = db.Limit(req.Size, offset)
	}
	err = db.Find(&users)
	return users, count, err
}

func (t *UserRepositoryImpl) BatchFindByTeamId(req request.UserBatchFindByTeamIdRequest) (users []response.UserResponseWithTeamId, err error) {
	err = t.Db.Table("users").
		Join("INNER", "user_team", "users.id = user_team.user_id").
		In("user_team.team_id", req.TeamId).
		Find(&users)
	return users, err
}

func (t *UserRepositoryImpl) FindById(id int64) (user model.User, err error) {
	_, err = t.Db.Table("users").ID(id).Get(&user)
	return user, err
}

// FindByUsername implements UserRepository
func (t *UserRepositoryImpl) FindByUsername(username string) (user model.User, err error) {
	_, err = t.Db.Table("users").Where("username = ?", username).Get(&user)
	return user, err
}

func (t *UserRepositoryImpl) FindByEmail(email string) (user model.User, err error) {
	_, err = t.Db.Table("users").Where("email = ?", email).Get(&user)
	return user, err
}
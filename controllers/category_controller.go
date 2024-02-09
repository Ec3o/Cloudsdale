package controllers

import (
	"github.com/elabosak233/pgshub/models/entity"
	"github.com/elabosak233/pgshub/services"
	"github.com/elabosak233/pgshub/utils/validator"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CategoryController interface {
	Create(ctx *gin.Context)
}

type CategoryControllerImpl struct {
	CategoryService services.CategoryService
}

func NewCategoryController(appService *services.Services) CategoryController {
	return &CategoryControllerImpl{
		CategoryService: appService.CategoryService,
	}
}

// Create
// @Summary create new category
// @Description
// @Tags Category
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param req body request.CategoryCreateRequest true "CategoryCreateRequest"
// @Router /api/challenges/ [post]
func (c *CategoryControllerImpl) Create(ctx *gin.Context) {
	req := entity.Category{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusBadRequest,
			"msg":  validator.GetValidMsg(err, &req),
		})
		return
	}
	err := c.CategoryService.Create(req)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusBadRequest,
			"msg":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
	})
}

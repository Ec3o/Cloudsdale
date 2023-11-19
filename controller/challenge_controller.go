package controller

import (
	req "github.com/elabosak233/pgshub/model/request/challenge"
	"github.com/elabosak233/pgshub/service"
	"github.com/elabosak233/pgshub/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ChallengeController struct {
	challengeService service.ChallengeService
}

func NewChallengeController(appService service.AppService) *ChallengeController {
	return &ChallengeController{
		challengeService: appService.ChallengeService,
	}
}

func (c *ChallengeController) Create(ctx *gin.Context) {
	createChallengeRequest := req.CreateChallengeRequest{}
	err := ctx.ShouldBindJSON(&createChallengeRequest)
	if err != nil {
		utils.FormatErrorResponse(ctx)
		return
	}
	_ = c.challengeService.Create(createChallengeRequest)
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
	})
}

func (c *ChallengeController) Update(ctx *gin.Context) {
	var updateChallengeRequest map[string]interface{}
	err := ctx.ShouldBindJSON(&updateChallengeRequest)
	if err != nil {
		utils.FormatErrorResponse(ctx)
		return
	}
	id := ctx.Param("id")
	updateChallengeRequest["id"] = id

	err = c.challengeService.Update(updateChallengeRequest)
	if err != nil {
		ctx.Header("Content-Type", "application/json")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "更新失败",
		})
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
	})
}

func (c *ChallengeController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.challengeService.Delete(id)
	if err != nil {
		ctx.Header("Content-Type", "application/json")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "删除失败",
		})
	}
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
	})
}

func (c *ChallengeController) FindById(ctx *gin.Context) {
	id := ctx.Param("id")
	challengeData := c.challengeService.FindById(id)
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": challengeData,
	})
}

func (c *ChallengeController) FindAll(ctx *gin.Context) {
	challengeData := c.challengeService.FindAll()
	ctx.Header("Content-Type", "application/json")
	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": challengeData,
	})

}
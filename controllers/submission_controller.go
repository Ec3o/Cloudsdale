package controllers

import (
	"github.com/elabosak233/pgshub/models/request"
	"github.com/elabosak233/pgshub/services"
	"github.com/elabosak233/pgshub/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SubmissionController interface {
	Find(ctx *gin.Context)
	Create(ctx *gin.Context)
	BatchFind(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type SubmissionControllerImpl struct {
	SubmissionService services.SubmissionService
}

func NewSubmissionControllerImpl(appService *services.AppService) SubmissionController {
	return &SubmissionControllerImpl{
		SubmissionService: appService.SubmissionService,
	}
}

// Find
// @Summary 提交记录查询
// @Description
// @Tags 提交
// @Accept json
// @Produce json
// @Param PgsToken header string true "PgsToken"
// @Param 查找请求 query request.SubmissionFindRequest false "SubmissionFindRequest"
// @Router /api/submissions/ [get]
func (c *SubmissionControllerImpl) Find(ctx *gin.Context) {
	isDetailed := func() int {
		if ctx.GetInt64("UserRole") <= 3 && utils.ParseIntParam(ctx.Query("is_detailed"), 0) == 1 {
			return 1
		}
		return 0
	}
	if ctx.Query("id") == "" {
		submissions, pageCount, total, _ := c.SubmissionService.Find(request.SubmissionFindRequest{
			UserId:      int64(utils.ParseIntParam(ctx.Query("user_id"), 0)),
			Status:      utils.ParseIntParam(ctx.Query("status"), 0),
			TeamId:      int64(utils.ParseIntParam(ctx.Query("team_id"), 0)),
			GameId:      int64(utils.ParseIntParam(ctx.Query("game_id"), 0)),
			IsDetailed:  isDetailed(),
			ChallengeId: int64(utils.ParseIntParam(ctx.Query("challenge_id"), 0)),
			SortBy:      ctx.QueryArray("sort_by"),
			Page:        utils.ParseIntParam(ctx.Query("page"), 0),
			Size:        utils.ParseIntParam(ctx.Query("size"), 0),
		})
		ctx.JSON(http.StatusOK, gin.H{
			"code":  http.StatusOK,
			"pages": pageCount,
			"total": total,
			"data":  submissions,
		})
	}
}

// BatchFind
// @Summary 提交记录批量查询
// @Description
// @Tags 提交
// @Accept json
// @Produce json
// @Param PgsToken header string true "PgsToken"
// @Param 查找请求 query request.SubmissionBatchFindRequest false "SubmissionBatchFindRequest"
// @Router /api/submissions/batch/ [get]
func (c *SubmissionControllerImpl) BatchFind(ctx *gin.Context) {
	submissions, err := c.SubmissionService.BatchFind(request.SubmissionBatchFindRequest{
		Size:             utils.ParseIntParam(ctx.Query("size"), 1),
		SizePerChallenge: utils.ParseIntParam(ctx.Query("size_per_challenge"), 0),
		UserId:           int64(utils.ParseIntParam(ctx.Query("user_id"), 0)),
		ChallengeId:      utils.MapStringsToInts(ctx.QueryArray("challenge_id")),
		Status:           utils.ParseIntParam(ctx.Query("status"), 0),
		SortBy:           ctx.QueryArray("sort_by"),
		IsDetailed:       ctx.Query("is_detailed") == "true",
		TeamId:           int64(utils.ParseIntParam(ctx.Query("team_id"), 0)),
		GameId:           int64(utils.ParseIntParam(ctx.Query("game_id"), 0)),
	})
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": submissions,
	})
}

// Create
// @Summary 提交
// @Description
// @Tags 提交
// @Accept json
// @Produce json
// @Param PgsToken header string true "PgsToken"
// @Param 创建请求 body request.SubmissionCreateRequest true "SubmissionCreateRequest"
// @Router /api/submissions/ [post]
func (c *SubmissionControllerImpl) Create(ctx *gin.Context) {
	submissionCreateRequest := request.SubmissionCreateRequest{}
	err := ctx.ShouldBindJSON(&submissionCreateRequest)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusBadRequest,
			"msg":  utils.GetValidMsg(err, &submissionCreateRequest),
		})
		return
	}
	submissionCreateRequest.UserId = ctx.GetInt64("UserId")
	status, err := c.SubmissionService.Create(submissionCreateRequest)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":   http.StatusOK,
		"status": status,
	})
}

func (c *SubmissionControllerImpl) Delete(ctx *gin.Context) {
	deleteSubmissionRequest := request.SubmissionDeleteRequest{}
	err := ctx.ShouldBindJSON(&deleteSubmissionRequest)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusBadRequest,
			"msg":  utils.GetValidMsg(err, &deleteSubmissionRequest),
		})
		return
	}
	err = c.SubmissionService.Delete(deleteSubmissionRequest.SubmissionId)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
	})
}
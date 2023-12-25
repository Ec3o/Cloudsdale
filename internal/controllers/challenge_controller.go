package controllers

import "github.com/gin-gonic/gin"

type ChallengeController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Find(ctx *gin.Context)
	FindById(ctx *gin.Context)
}
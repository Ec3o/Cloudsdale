package controllers

import "github.com/gin-gonic/gin"

type InstanceController interface {
	Create(ctx *gin.Context)
	Remove(ctx *gin.Context)
	Renew(ctx *gin.Context)
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
}
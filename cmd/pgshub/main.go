package main

import (
	"github.com/elabosak233/pgshub/controller"
	_ "github.com/elabosak233/pgshub/docs"
	"github.com/elabosak233/pgshub/router"
	"github.com/elabosak233/pgshub/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"os"
	"strconv"
)

// @title PgsHub Backend
// @version 1.0
// @description PgsHub Backend
func main() {
	Welcome()
	utils.InitLogger()
	utils.LoadConfig()
	db := DatabaseConnection()

	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	cor := cors.DefaultConfig()
	cor.AllowOrigins = []string{"http://localhost:3000"}
	cor.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	r.Use(cors.New(cor))

	appRepository := InitRepositories(db)
	appService := InitServices(appRepository)
	router.NewRouters(
		r,
		controller.NewUserController(appService),
		controller.NewGroupController(appService),
		controller.NewChallengeController(appService),
		controller.NewUserGroupController(appService),
	)

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler()))

	s := &http.Server{
		Addr:    viper.GetString("Server.Host") + ":" + viper.GetString("Server.Port"),
		Handler: r,
	}
	utils.Logger.Infof("PgsHub Core 服务已启动，访问地址 %s:%d", viper.GetString("Server.Host"), viper.GetInt("Server.Port"))
	_ = s.ListenAndServe()
}

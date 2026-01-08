package main

import(
	"github.com/gin-gonic/gin"
	"net/http"
)

func main(){
	r:=gin.Default();

	r.GET("/",func(c *gin.Context){
		c.JSON(http.StatusOK,gin.H{
			"message": "Welcome to Prism AI Gateway!",
			"version": "0.1.0",
			"status":  "running",
		})
	})

	r.GET("/health",func(c *gin.Context){
		c.JSON(http.StatusOK,gin.H{
			"status": "healthy",
		})
	})

	// 启动HTTP服务器，监听8080端口
	r.Run(":8080")
}
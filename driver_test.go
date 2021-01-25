/**
2 * @Author: Nico
3 * @Date: 2021/1/25 9:42
4 */
package ginx

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zeta-io/zeta"
	"testing"
)

func TestNew1(t *testing.T) {
	z := zeta.New(New(gin.New()))
	z.Any("/:p1/test", func(c *gin.Context, s struct{
		Q1 string `query:"q1"`
		P1 string `path:"p1"`
		Body map[string]interface{} `body:""`
		//F1 *multipart.FileHeader `file:"f1"`
		H1 string `header:"h1"`
		C1 string `cookie:"c1"`
		B1 string `body:"b1"`
	}) interface{}{
		return s
	})
	z.Run(":8080")
}

func TestNew2(t *testing.T) {
	z := zeta.New(New(gin.New()))
	z.Use(func(c *gin.Context){
		fmt.Println("use1")
		c.Next()
	},func(c *gin.Context){
		fmt.Println("use2")
		c.Next()
	})
	z.Get("/test", func(c *gin.Context){
		fmt.Println("api /test pre")
		c.Next()
	},func(c *gin.Context){
		fmt.Println("api /test")
		c.Next()
	},func(c *gin.Context){
		fmt.Println("api /test after")
	})
	g := z.Group("/group")
	g.Use(func(c *gin.Context){
		fmt.Println("group use1")
		c.Next()
	},func(c *gin.Context){
		fmt.Println("group use2")
		c.Next()
	})
	g.Get("/test", func(c *gin.Context){
		fmt.Println("group api /test pre")
	},func(c *gin.Context){
		fmt.Println("group api /test")
	},func(c *gin.Context){
		fmt.Println("group api /test after")
	})
	z.Run(":8080")
}
# Introduce
Gin driver for [zeta](https://github.com/zeta-io/zeta).
# Feature
Features:
 - Parameters of the assembly
 - Parameter validation
# Usage
Sample:
```go
package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zeta-io/ginx"
	"github.com/zeta-io/zeta"
)

func list(context context.Context, c1 *context.Context, args struct{
	Name string `json:"name" param:"query,name" validator:"required"`
}) (string, error){
	fmt.Println(args.Name)
	return "hello zeta", nil
}

func main() {
	router := zeta.Router("/api/:version/users")
	router.Get("", list)

	e := zeta.New(router, ginx.New(gin.New())).Run(":8080")
	if e != nil{
		panic(e)
	}
}
```
Request parameter tag:
 - param:"query,${name}": Bind query parameters.
 - param:"form,${name}": Bind form parameters.
 - param:"body": Bind body parameters.
 - param:"path,${name}": Bind path parameters.

# Validator
It's used [https://github.com/go-playground/validator](https://github.com/go-playground/validator) by default.


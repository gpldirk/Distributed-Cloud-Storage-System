package main

import (
	"github.com/cloud/service/apigw/route"
)

func main() {
	r := route.Router()
	r.Run(":8080")
}

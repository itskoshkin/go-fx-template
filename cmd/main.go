package main

import "go-fx-template/internal/core"

// @title Go REST Template API
// @description API for managing items
// @version 1.0
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter token as "Bearer <access_token>"
func main() {
	core.Load().Run()
}

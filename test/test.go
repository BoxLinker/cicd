package main

import (
	"github.com/BoxLinker/cicd/models"
	"fmt"
)

func main(){
	t := models.SCMType("github")
	fmt.Println(string(t))
}


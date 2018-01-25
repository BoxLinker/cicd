package main

import "fmt"

type User struct {
	Name string
}

func main(){
	var arr *User

	arr.Name = "123"
	fmt.Printf("arr: %v", arr)
}

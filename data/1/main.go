package main

import (
	"fmt"
)

func main() {
	var (
		asd = Roles{}
		asd2 = Roles{}
		asd1 = Permissions{}
	)
	asd2.CreateRole()
	asd.ReadRoleFile()
	asd1.ReadPermissionsFile()
	fmt.Println(asd)
	fmt.Println(asd1)
}

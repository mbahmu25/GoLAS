package main

import (
	"GoLas/readLas"
	"fmt"
)

func main() {
	fileData := readLas.Read_file("output_fix.las")
	fmt.Println(fileData[102420])
}

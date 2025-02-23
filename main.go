package main

import (
	"GoLas/readLas"
	"fmt"
	"os"
)

func main() {
	filePath := os.Args
	fmt.Println(filePath)
	fileData := readLas.Read_file("output_fix.las")
	fmt.Println(fileData[102420])
}

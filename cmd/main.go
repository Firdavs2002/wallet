package main

import (
	"fmt"

	"github.com/Firdavs2002/wallet/pkg/wallet"
)

func main() {
	svc := &wallet.Service{}
	svc.RegisterAccount("+992000000001")
	svc.RegisterAccount("+992000000002")
	svc.RegisterAccount("+992000000003")
	svc.ExportToFile("../data/export.txt")
	svc.ImportFromFile("../data/import.txt")
	pays := []int64{1, 2, 3}
	p := pays // s = append(s, pays...)
	fmt.Println(p)
}

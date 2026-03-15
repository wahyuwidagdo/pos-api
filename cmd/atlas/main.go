package main

import (
	"fmt"
	"io"
	"os"

	"pos-api/internal/models"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	sb := gormschema.New("postgres")
	str, err := sb.Load(
		&models.User{},
		&models.Product{},
		&models.Category{},
		&models.Transaction{},
		&models.TransactionDetail{},
		&models.InventoryLog{},
		&models.CashFlow{},
		&models.PaymentMethod{},
		&models.StoreSetting{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, str)
}

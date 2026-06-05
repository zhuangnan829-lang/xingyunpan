package main

import (
	"fmt"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/pkg/logger"
)

func main() {
	if err := config.LoadDefault(); err != nil {
		panic(err)
	}
	if err := logger.Init(&logger.Config{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}); err != nil {
		panic(err)
	}
	if err := config.InitDatabase(); err != nil {
		panic(err)
	}
	defer config.CloseDatabase()

	if err := config.AutoMigrate(&model.FileCustomPropertyValue{}); err != nil {
		panic(err)
	}

	fmt.Println("file custom property table migrated")
}

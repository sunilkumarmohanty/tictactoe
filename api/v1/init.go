package v1

import (
	"log"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		log.Fatalf("cannot initialize logger. Error : %v", err)
	}
}

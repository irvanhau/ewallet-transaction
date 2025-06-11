package cmd

import (
	"ewallet-transaction/external"
	"ewallet-transaction/helpers"
	"ewallet-transaction/internal/api"
	"ewallet-transaction/internal/interfaces"
	"ewallet-transaction/internal/repository"
	"ewallet-transaction/internal/services"
	"log"

	"github.com/gin-gonic/gin"
)

func ServeHTTP() {
	d := dependencyInject()

	r := gin.Default()

	r.GET("/health", d.HealthCheckAPI.HealthCheckHandlerHTTP)

	transactionV1 := r.Group("/transaction/v1")
	transactionV1.POST("/create", d.MiddlewareValidateToken, d.TransactionAPI.Create)
	transactionV1.POST("/refund", d.MiddlewareValidateToken, d.TransactionAPI.RefundTransaction)
	transactionV1.PUT("/update-status/:reference", d.MiddlewareValidateToken, d.TransactionAPI.UpdateStatusTransaction)
	transactionV1.GET("/:reference", d.MiddlewareValidateToken, d.TransactionAPI.GetTransactionDetail)
	transactionV1.GET("/", d.MiddlewareValidateToken, d.TransactionAPI.GetTransaction)

	err := r.Run(":" + helpers.GetEnv("PORT", "8080"))
	if err != nil {
		log.Fatal(err)
	}
}

type Dependency struct {
	HealthCheckAPI interfaces.IHealthCheckAPI
	External       interfaces.IExternal
	TransactionAPI interfaces.ITransactionHandler
}

func dependencyInject() Dependency {
	healthCheckSvc := &services.HealthCheck{}
	healthCheckAPI := &api.HealthCheck{
		HealthCheckServices: healthCheckSvc,
	}
	external := &external.External{}

	transactionRepo := &repository.TransactionRepo{
		DB: helpers.DB,
	}
	transactionSvc := &services.TransactionService{
		TransactionRepo: transactionRepo,
		External:        external,
	}
	transactionAPI := &api.TransactionAPI{
		TransactionService: transactionSvc,
	}

	return Dependency{
		HealthCheckAPI: healthCheckAPI,
		External:       external,
		TransactionAPI: transactionAPI,
	}
}

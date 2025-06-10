package interfaces

import (
	"context"
	"ewallet-transaction/internal/models"

	"github.com/gin-gonic/gin"
)

type ITransactionRepository interface {
	CreateTransaction(ctx context.Context, trx *models.Transaction) error
}

type ITransactionService interface {
	CreateTransaction(ctx context.Context, req *models.Transaction) (models.CreateTransactionResponse, error)
}

type ITransactionHandler interface {
	Create(c *gin.Context)
}

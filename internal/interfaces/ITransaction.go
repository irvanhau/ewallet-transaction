package interfaces

import (
	"context"
	"ewallet-transaction/internal/models"

	"github.com/gin-gonic/gin"
)

type ITransactionRepository interface {
	CreateTransaction(context.Context, *models.Transaction) error
	UpdateStatusTransaction(context.Context, string, string, string) error
	GetTransactionByReference(context.Context, string, bool) (models.Transaction, error)
	GetTransaction(ctx context.Context, userID uint) ([]models.Transaction, error)
}

type ITransactionService interface {
	CreateTransaction(context.Context, *models.Transaction) (models.CreateTransactionResponse, error)
	UpdateStatusTransaction(ctx context.Context, tokenData models.TokenData, req *models.UpdateStatusTransaction) error
	GetTransaction(ctx context.Context, userID uint) ([]models.Transaction, error)
	GetTransactionDetail(ctx context.Context, reference string) (models.Transaction, error)
}

type ITransactionHandler interface {
	Create(*gin.Context)
	UpdateStatusTransaction(c *gin.Context)
	GetTransaction(c *gin.Context)
	GetTransactionDetail(c *gin.Context)
}

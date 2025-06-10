package constants

const (
	SuccessMessage      = "success"
	ErrFailedBadRequest = "data tidak sesuai"
	ErrServerError      = "terjadi kesalahan pada server"
)

const (
	TransactionStatusPending  = "PENDING"
	TransactionStatusSuccess  = "SUCCESS"
	TransactionStatusFailed   = "FAILED"
	TransactionStatusReversed = "REVERSED"
)

const (
	TransactionTypeTopup    = "TOPUP"
	TransactionTypePurchase = "PURCHASE"
	TransactionTypeRefund   = "REFUND"
)

var MapTransactionType = map[string]bool{
	TransactionTypeTopup:    true,
	TransactionTypeRefund:   true,
	TransactionTypePurchase: true,
}

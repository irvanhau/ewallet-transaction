package services

import (
	"context"
	"encoding/json"
	"ewallet-transaction/constants"
	"ewallet-transaction/external"
	"ewallet-transaction/helpers"
	"ewallet-transaction/internal/interfaces"
	"ewallet-transaction/internal/models"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type TransactionService struct {
	TransactionRepo interfaces.ITransactionRepository
	External        interfaces.IExternal
}

func (s *TransactionService) CreateTransaction(ctx context.Context, req *models.Transaction) (models.CreateTransactionResponse, error) {
	var (
		resp models.CreateTransactionResponse
	)

	req.TransactionStatus = constants.TransactionStatusPending
	req.Reference = helpers.GenerateReference()

	jsonAdditionalInfo := map[string]interface{}{}
	if req.AdditionalInfo != "" {
		err := json.Unmarshal([]byte(req.AdditionalInfo), &jsonAdditionalInfo)
		if err != nil {
			return resp, errors.Wrap(err, "additional info type is invalid")
		}
	}

	err := s.TransactionRepo.CreateTransaction(ctx, req)
	if err != nil {
		return resp, errors.Wrap(err, "failed to insert transaction")
	}

	resp.Reference = req.Reference
	resp.TransactionStatus = req.TransactionStatus
	return resp, nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, userID uint) ([]models.Transaction, error) {
	return s.TransactionRepo.GetTransaction(ctx, userID)
}

func (s *TransactionService) GetTransactionDetail(ctx context.Context, reference string) (models.Transaction, error) {
	return s.TransactionRepo.GetTransactionByReference(ctx, reference, true)
}

func (s *TransactionService) UpdateStatusTransaction(ctx context.Context, tokenData models.TokenData, req *models.UpdateStatusTransaction) error {
	// Get transaction by reference
	trx, err := s.TransactionRepo.GetTransactionByReference(ctx, req.Reference, false)
	if err != nil {
		return errors.Wrap(err, "failed to get transaction")
	}

	// Validate transaction status flow
	statusValid := false
	mapStatusFlow := constants.MapTransactionStatusFlow[trx.TransactionStatus]
	for i := range mapStatusFlow {
		if mapStatusFlow[i] == req.TransactionStatus {
			statusValid = true
		}
	}
	if !statusValid {
		return fmt.Errorf("transaction status flow invalid. request status = %s", req.TransactionStatus)
	}

	// Request update balance to ewallet-wallet
	reqUpdateBalance := external.UpdateBalance{
		Amount:    trx.Amount,
		Reference: req.Reference,
	}

	if req.TransactionStatus == constants.TransactionStatusReversed {
		reqUpdateBalance.Reference = "REVERSED-" + req.Reference

		now := time.Now()
		expiredReversalTime := trx.CreatedAt.Add(constants.MaximumReversalDuration)
		if expiredReversalTime.Before(now) {
			return errors.New("reversal duration is already expired")
		}
	}
	var (
		errUpdateBalance error
	)
	switch trx.TransactionType {
	case constants.TransactionTypeTopup:
		if req.TransactionStatus == constants.TransactionStatusSuccess {
			_, errUpdateBalance = s.External.CreditBalance(ctx, tokenData.Token, reqUpdateBalance)
		} else if req.TransactionStatus == constants.TransactionStatusReversed {
			_, errUpdateBalance = s.External.DebitBalance(ctx, tokenData.Token, reqUpdateBalance)
		}
	case constants.TransactionTypePurchase:
		if req.TransactionStatus == constants.TransactionStatusSuccess {
			_, errUpdateBalance = s.External.DebitBalance(ctx, tokenData.Token, reqUpdateBalance)
		} else if req.TransactionStatus == constants.TransactionStatusReversed {
			_, errUpdateBalance = s.External.CreditBalance(ctx, tokenData.Token, reqUpdateBalance)
		}
	}

	if errUpdateBalance != nil {
		return errors.Wrap(errUpdateBalance, "failed to update balance")
	}

	// Update additional info
	var (
		currentAdditionalInfo = map[string]interface{}{}
		newAdditionalInfo     = map[string]interface{}{}
	)

	if trx.AdditionalInfo != "" {
		err = json.Unmarshal([]byte(trx.AdditionalInfo), &currentAdditionalInfo)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal current additional info")
		}
	}

	if req.AdditionalInfo != "" {
		err = json.Unmarshal([]byte(req.AdditionalInfo), &newAdditionalInfo)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal new additional info")
		}
	}

	for key, val := range newAdditionalInfo {
		currentAdditionalInfo[key] = val
	}

	byteAdditionalInfo, err := json.Marshal(currentAdditionalInfo)
	if err != nil {
		return errors.Wrap(err, "failed to marshal merged additional info")
	}

	// Update Status in DB
	err = s.TransactionRepo.UpdateStatusTransaction(ctx, req.Reference, req.TransactionStatus, string(byteAdditionalInfo))
	if err != nil {
		return errors.Wrap(err, "failed to update status transaction")
	}

	return nil
}

func (s *TransactionService) RefundTransaction(ctx context.Context, tokenData models.TokenData, req *models.RefundTransaction) (models.CreateTransactionResponse, error) {
	var (
		resp models.CreateTransactionResponse
	)
	// Get transaction by reference
	trx, err := s.TransactionRepo.GetTransactionByReference(ctx, req.Reference, false)
	if err != nil {
		return resp, errors.Wrap(err, "failed to get transaction")
	}

	if trx.TransactionStatus != constants.TransactionStatusSuccess && trx.TransactionStatus != constants.TransactionTypePurchase {
		return resp, errors.New("current transaction status is not success or transaction type is not purchase")
	}

	refundReference := "REFUND-" + req.Reference
	reqCreditBalance := external.UpdateBalance{
		Reference: refundReference,
		Amount:    trx.Amount,
	}
	_, err = s.External.CreditBalance(ctx, tokenData.Token, reqCreditBalance)
	if err != nil {
		return resp, errors.Wrap(err, "failed to credit balance")
	}

	transaction := models.Transaction{
		UserID:            tokenData.UserID,
		Amount:            trx.Amount,
		TransactionType:   constants.TransactionTypeRefund,
		TransactionStatus: constants.TransactionStatusSuccess,
		Reference:         refundReference,
		Description:       req.Description,
		AdditionalInfo:    req.AdditionalInfo,
	}

	err = s.TransactionRepo.CreateTransaction(ctx, &transaction)
	if err != nil {
		return resp, errors.Wrap(err, "failed to create new transaction refund")
	}

	resp.Reference = refundReference
	resp.TransactionStatus = transaction.TransactionStatus

	return resp, nil
}

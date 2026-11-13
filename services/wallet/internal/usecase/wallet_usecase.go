package usecase

import (
	"context"
	"time"
	"wallet-service/internal/domain"

	"github.com/google/uuid"
)

type walletUsecase struct {
	walletRepo      domain.WalletRepository
	transactionRepo domain.TransactionRepository
	transactor      domain.Transactor
}

// GetBalance implements [domain.WalletUsecase].
func (u walletUsecase) GetBalance(ctx context.Context, userID string) (*domain.WalletResponse, error) {
	//1. cari dompet user
	wallet, err := u.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err == domain.ErrWalletNotFound {
			//2. lazy initialization : buat dompet kosong
			newWallet := &domain.Wallet{
				ID:        uuid.New().String(),
				UserID:    userID,
				Balance:   0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if errCreate := u.walletRepo.Create(ctx, newWallet); errCreate != nil {
				return nil, errCreate
			}
			wallet = newWallet
		} else {
			return nil, err
		}
	}
	//3. kembalikan response
	return &domain.WalletResponse{
		ID:      wallet.ID,
		UserID:  wallet.UserID,
		Balance: wallet.Balance,
	}, nil
}

// TopUp implements [domain.WalletUsecase].
func (u walletUsecase) TopUp(ctx context.Context, userID string, req *domain.TopUpRequest) (*domain.TransactionResponse, error) {
	//1. validasi
	if req.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	//2. dapatkan dompet (atau buat baru jika lazy init memicu)
	walletRes, err := u.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	//3. persiapkan entitas mutasi
	transactionID := uuid.New().String()
	transaction := &domain.Transaction{
		ID:                transactionID,
		RecipientWalletID: walletRes.ID,
		Type:              domain.TxTypeTopUp,
		Amount:            req.Amount,
		Status:            domain.TxStatusSuccess,
		ReferenceID:       uuid.New().String(),
		CreatedAt:         time.Now(),
	}

	//4. eksekusi saldo + ledger dalam satu transaksi (ACID)
	err = u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		//4a. update saldo
		if err := u.walletRepo.UpdateBalance(txCtx, walletRes.ID, req.Amount); err != nil {
			return err
		}
		//4b. catat ledger
		return u.transactionRepo.Create(txCtx, transaction)
	})
	if err != nil {
		return nil, err
	}

	//5. kembalikan response
	return &domain.TransactionResponse{
		ID:          transaction.ID,
		WalletID:    transaction.RecipientWalletID,
		Type:        string(transaction.Type),
		Amount:      transaction.Amount,
		Status:      string(transaction.Status),
		ReferenceID: transaction.ReferenceID,
		CreatedAt:   transaction.CreatedAt,
	}, nil
}

func NewWalletUsecase(wRepo domain.WalletRepository, tRepo domain.TransactionRepository, transactor domain.Transactor) domain.WalletUsecase {
	return walletUsecase{
		walletRepo:      wRepo,
		transactionRepo: tRepo,
		transactor:      transactor,
	}
}

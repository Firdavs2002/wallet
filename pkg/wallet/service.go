package wallet

import (
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Firdavs2002/wallet/pkg/types"
	"github.com/google/uuid"
)

//ErrPhoneRegistered - телефон уже регитрирован
var ErrPhoneRegistered = errors.New("phone already registred")

//ErrAmountMustBePositive - счёт не может быть пустым
var ErrAmountMustBePositive = errors.New("amount must be greater than zero")

//ErrAccountNotFound - пользователь не найден
var ErrAccountNotFound = errors.New("account not found")

//ErrNotEnoughtBalance - на счете недостаточно средств
var ErrNotEnoughtBalance = errors.New("account not enough balance")

//ErrPaymentNotFound - платеж не найден
var ErrPaymentNotFound = errors.New("payment not found")

// ErrFavoriteNotFound - Избранное не найдено
var ErrFavoriteNotFound = errors.New("favorite not found")

// ErrFileNotFound - файл не найден
var ErrFileNotFound = errors.New("file not fount")

// Service представляет информацию о пользователе.
type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

// RegisterAccount - метод для регистрация нового прользователя.
func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}
	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)
	return account, nil
}

//Pay метод для регистрации платижа
func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var account *types.Account

	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughtBalance
	}

	account.Balance -= amount

	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}

	s.payments = append(s.payments, payment)
	return payment, nil
}

// FindAccountByID ищем пользователя по ID
func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	var account *types.Account

	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	return account, nil
}

// FindPaymentByID ищем платёж по ID
func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	var payment *types.Payment

	for _, pay := range s.payments {
		if pay.ID == paymentID {
			payment = pay
		}
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	return payment, nil
}

// FindFavoriteByID ищем платёж по ID в Избранное
func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}
	return nil, ErrFavoriteNotFound
}

//Deposit method
func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount < 0 {
		return ErrAmountMustBePositive
	}

	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return err
	}

	account.Balance += amount
	return nil
}

// Reject метод для отмены покупок
func (s *Service) Reject(paymentID string) error {
	pay, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}

	acc, err := s.FindAccountByID(pay.AccountID)
	if err != nil {
		return err
	}

	pay.Status = types.PaymentStatusFail
	acc.Balance += pay.Amount

	return nil
}

// Repeat позволāет по идентификатору повторитþ платёж
func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	pay, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	payment, err := s.Pay(pay.AccountID, pay.Amount, pay.Category)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// FavoritePayment добавления новых Избранных
func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)

	if err != nil {
		return nil, err
	}

	favoriteID := uuid.New().String()
	newFavorite := &types.Favorite{
		ID:        favoriteID,
		AccountID: payment.AccountID,
		Name:      name,
		Amount:    payment.Amount,
		Category:  payment.Category,
	}

	s.favorites = append(s.favorites, newFavorite)
	return newFavorite, nil
}

//PayFromFavorite для совершения платежа в Избранное
func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}

	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// ExportToFile экспартирует данные аккаунтов в файл
func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Print(err)
		return ErrFileNotFound
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	str := ""

	for _, acc := range s.accounts {
		ID := strconv.Itoa(int(acc.ID)) + ";"
		phone := string(acc.Phone) + ";"
		balance := strconv.Itoa(int(acc.Balance))

		str += ID
		str += phone
		str += balance + "|"
	}
	_, err = file.Write([]byte(str))
	if err != nil {
		log.Print(err)
		return ErrFileNotFound
	}

	return nil
}

// ImportFromFile импортирует данные с аккаунта
func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
		return ErrFileNotFound
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()

	s.ExportToFile(path) // создаём файл чтоб с него импортироватӣ (плохая идея :-) )

	content := make([]byte, 0)
	buf := make([]byte, 4)
	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
			return ErrFileNotFound
		}
		content = append(content, buf[:read]...)
	}
	data := string(content)

	accounts := strings.Split(string(data), "|")
	accounts = accounts[:len(accounts)-1] // это фигня служит для того чтоб цикл пощел на n - 1, т.к split добавил в слайс ещё один элемент (причём пустой)
	for _, account := range accounts {
		vals := strings.Split(account, ";")

		ID, err := strconv.Atoi(vals[0])
		if err != nil {
			return err
		}

		balance, err := strconv.Atoi(vals[2])
		if err != nil {
			return err
		}

		newAccount := &types.Account{
			ID:      int64(ID),
			Phone:   types.Phone(vals[1]),
			Balance: types.Money(balance),
		}

		s.accounts = append(s.accounts, newAccount)
	}

	return nil
}

package fio

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

// TransactionsResponse represents response transactions.
type TransactionsResponse struct {
	Info         StatementInfo
	Transactions []Transaction
}

// StatementInfo represents account statement info.
type StatementInfo struct {
	AccountID      int64
	BankID         string
	Currency       string
	IBAN           string
	BIC            string
	OpeningBalance decimal.Decimal
	ClosingBalance decimal.Decimal
	DateStart      time.Time
	DateEnd        time.Time
	YearList       int64
	IDFrom         int64
	IDTo           int64
	IDLastDownload int64
	IDList         int64
}

// Transaction represents transaction.
type Transaction struct {
	ID                 int64
	Date               time.Time
	Amount             decimal.Decimal
	Currency           string
	Account            string
	AccountName        string
	BankName           string
	ConstantSymbol     string
	VariableSymbol     string
	UserIdentification string
	RecipientMessage   string
	Type               string
	Comment            string
	BIC                string
	OrderID            string
}

// ByPeriodOptions represents options passed to ByPeriod.
type ByPeriodOptions struct {
	DateFrom time.Time
	DateTo   time.Time
}

// TransactionsService is a service for working with transactions.
type TransactionsService struct {
	client *Client
}

// ByPeriod returns transactions in date period.
func (s *TransactionsService) ByPeriod(ctx context.Context, opts ByPeriodOptions) (*TransactionsResponse, error) {
	urlStr := s.client.buildURL("ib_api/rest/periods", fmtDate(opts.DateFrom), fmtDate(opts.DateTo), "transactions.xml")
	req, err := s.client.get(urlStr)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return parseTransactionsResponse(resp.Body)
}

// ExportOptions represents options passed to Export.
type ExportOptions struct {
	DateFrom time.Time
	DateTo   time.Time
	Format   ExportFormat
}

// Export writes transactions in date period to provided writer.
func (s *TransactionsService) Export(ctx context.Context, opts ExportOptions, w io.Writer) error {
	exportFmt := fmt.Sprintf("transactions.%v", opts.Format)
	urlStr := s.client.buildURL("ib_api/rest/periods", fmtDate(opts.DateFrom), fmtDate(opts.DateTo), exportFmt)
	req, err := s.client.get(urlStr)
	if err != nil {
		return err
	}

	resp, err := s.client.do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, cerr := io.Copy(w, resp.Body)
	return cerr
}

// GetStatementOptions represents options passed to GetStatement.
type GetStatementOptions struct {
	Year int
	ID   int
}

// GetStatement returns statement by its year/id.
func (s *TransactionsService) GetStatement(ctx context.Context, opts GetStatementOptions) (*TransactionsResponse, error) {
	urlStr := s.client.buildURL("ib_api/rest/by-id", strconv.Itoa(opts.Year), strconv.Itoa(opts.ID), "transactions.xml")
	req, err := s.client.get(urlStr)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return parseTransactionsResponse(resp.Body)
}

type ExportStatementOptions struct {
	Year   int
	ID     int
	Format ExportFormat
}

// ExportStatement writes statement by its year/id to provided writer.
func (s *TransactionsService) ExportStatement(ctx context.Context, opts ExportStatementOptions, w io.Writer) error {
	exportFmt := fmt.Sprintf("transactions.%v", opts.Format)
	urlStr := s.client.buildURL("ib_api/rest/by-id", strconv.Itoa(opts.Year), strconv.Itoa(opts.ID), exportFmt)
	req, err := s.client.get(urlStr)
	if err != nil {
		return err
	}

	resp, err := s.client.do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, cerr := io.Copy(w, resp.Body)
	return cerr
}

// SinceLastDownload returns transactions since last download.
func (s *TransactionsService) SinceLastDownload(ctx context.Context) (*TransactionsResponse, error) {
	urlStr := s.client.buildURL("ib_api/rest/last", "transactions.xml")
	req, err := s.client.get(urlStr)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return parseTransactionsResponse(resp.Body)
}

// SetLastDownloadIDOptions represents options passed to SetLastDownloadID.
type SetLastDownloadIDOptions struct {
	ID int
}

// SetLastDownloadID sets the last downloaded id of statement.
func (s *TransactionsService) SetLastDownloadID(ctx context.Context, opts SetLastDownloadIDOptions) error {
	urlStr := s.client.buildURL("ib_api/rest/set-last-id", strconv.Itoa(opts.ID))
	req, err := s.client.get(urlStr)
	if err != nil {
		return err
	}

	resp, err := s.client.do(ctx, req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

// SetLastDownloadDateOptions represents options passed to SetLastDownloadDate.
type SetLastDownloadDateOptions struct {
	Date time.Time
}

// SetLastDownloadDate sets the last download date of statement.
func (s *TransactionsService) SetLastDownloadDate(ctx context.Context, opts SetLastDownloadDateOptions) error {
	urlStr := s.client.buildURL("ib_api/rest/set-last-date", fmtDate(opts.Date))
	req, err := s.client.get(urlStr)
	if err != nil {
		return err
	}

	resp, err := s.client.do(ctx, req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func fmtDate(t time.Time) string {
	return t.Format(dateFormat)
}

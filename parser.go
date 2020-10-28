package fio

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

// https://www.fio.cz/docs/cz/API_Bankovnictvi.pdf
const (
	fieldTransactionID      = "22" // ID pohybu
	fieldDate               = "0"  // Datum
	fieldAmount             = "1"  // Objem
	fieldCurrency           = "14" // Měna
	fieldAccount            = "2"  // Protiúčet
	fieldAccountName        = "10" // Název protiúčtu
	fieldBankCode           = "3"  // Kód banky
	fieldBankName           = "12" // Název banky
	fieldConstantSymbol     = "4"  // KS
	fieldVariableSymbol     = "5"  // VS
	fieldSpecificSymbol     = "6"  // SS
	fieldUserIdentification = "7"  // Uživatelská identifikace
	fieldRecipientMessage   = "16" // Zpráva pro příjemce
	fieldType               = "8"  // Typ pohybu
	fieldSpecification      = "18" // Upřesnění
	fieldComment            = "25" // Komentář
	fieldBIC                = "26" // BIC
	fieldOrderID            = "17" // ID pokynu
	fieldAuthor             = "9"  // Provedl
	fieldPayerReference     = "27" // Reference plátce

	xmlTimeFormat = "2006-01-02-07:00"
)

var (
	xmlGMTLocation *time.Location
)

func init() {
	loc, err := time.LoadLocation("GMT")
	if err != nil {
		panic(err)
	}
	xmlGMTLocation = loc
}

type xmlErrorResponse struct {
	XMLName xml.Name       `xml:"response"`
	Result  xmlErrorResult `xml:"result"`
}

type xmlErrorResult struct {
	ErrorCode string `xml:"errorCode"`
	Status    string `xml:"status"`
	Message   string `xml:"message"`
	Detail    string `xml:"detail"`
}

type xmlTransactionsResponse struct {
	XMLName      xml.Name          `xml:"AccountStatement"`
	Info         xmlStatementInfo  `xml:"Info"`
	Transactions []xmlTtransaction `xml:"TransactionList>Transaction"`
}

type xmlStatementInfo struct {
	AccountID      int64      `xml:"accountId"`
	BankID         string     `xml:"bankId"`
	Currency       string     `xml:"currency"`
	IBAN           string     `xml:"iban"`
	BIC            string     `xml:"bic"`
	OpeningBalance xmlDecimal `xml:"openingBalance"`
	ClosingBalance xmlDecimal `xml:"closingBalance"`
	DateStart      xmlTime    `xml:"dateStart"`
	DateEnd        xmlTime    `xml:"dateEnd"`
	YearList       int64      `xml:"yearList"`
	IDList         int64      `xml:"idList"`
	IDFrom         int64      `xml:"idFrom"`
	IDTo           int64      `xml:"idTo"`
	IDLastDownload int64      `xml:"idLastDownload"`
}

type xmlTtransaction struct {
	Columns []xmlTransactionColumn `xml:",any"`
}

type xmlTransactionColumn struct {
	ID    string `xml:"id,attr"`
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type xmlTime struct {
	time.Time
}

func (t *xmlTime) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var v string
	if err := dec.DecodeElement(&v, &start); err != nil {
		return err
	}

	tt, err := parseGMTTime(v)
	if err != nil {
		return err
	}

	*t = xmlTime{tt}
	return nil
}

type xmlDecimal struct {
	decimal.Decimal
}

func (d *xmlDecimal) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var v string
	if err := dec.DecodeElement(&v, &start); err != nil {
		return err
	}

	dd, err := parseAmount(v)
	if err != nil {
		return err
	}

	*d = xmlDecimal{dd}
	return nil
}

func parseTransactionsResponse(r io.Reader) (*TransactionsResponse, error) {
	var xmlResp xmlTransactionsResponse
	enc := xml.NewDecoder(r)
	if err := enc.Decode(&xmlResp); err != nil {
		return nil, err
	}

	resp := new(TransactionsResponse)
	resp.Info = StatementInfo{
		AccountID:      xmlResp.Info.AccountID,
		BankID:         xmlResp.Info.BankID,
		Currency:       xmlResp.Info.Currency,
		IBAN:           xmlResp.Info.IBAN,
		BIC:            xmlResp.Info.BIC,
		OpeningBalance: xmlResp.Info.OpeningBalance.Decimal,
		ClosingBalance: xmlResp.Info.ClosingBalance.Decimal,
		DateStart:      xmlResp.Info.DateStart.Time,
		DateEnd:        xmlResp.Info.DateEnd.Time,
		YearList:       xmlResp.Info.YearList,
		IDList:         xmlResp.Info.IDList,
		IDFrom:         xmlResp.Info.IDFrom,
		IDTo:           xmlResp.Info.IDTo,
		IDLastDownload: xmlResp.Info.IDLastDownload,
	}

	for _, xmlTx := range xmlResp.Transactions {
		tx, err := parseTransaction(xmlTx)
		if err != nil {
			return nil, err
		}
		resp.Transactions = append(resp.Transactions, *tx)
	}
	return resp, nil
}

func parseTransaction(t xmlTtransaction) (*Transaction, error) {
	tx := new(Transaction)
	for _, col := range t.Columns {
		switch col.ID {
		case fieldTransactionID:
			v, err := parseInteger(col.Value)
			if err != nil {
				return nil, err
			}
			tx.ID = v
		case fieldDate:
			v, err := parseGMTTime(col.Value)
			if err != nil {
				return nil, err
			}
			tx.Date = v
		case fieldAmount:
			v, err := parseAmount(col.Value)
			if err != nil {
				return nil, err
			}
			tx.Amount = v
		case fieldCurrency:
			tx.Currency = col.Value
		case fieldAccount:
			tx.Account = col.Value
		case fieldBankCode:
			tx.BankCode = col.Value
		case fieldAccountName:
			tx.AccountName = col.Value
		case fieldBankName:
			tx.BankName = col.Value
		case fieldConstantSymbol:
			tx.ConstantSymbol = col.Value
		case fieldVariableSymbol:
			tx.VariableSymbol = col.Value
		case fieldSpecificSymbol:
			tx.SpecificSymbol = col.Value
		case fieldUserIdentification:
			tx.UserIdentification = col.Value
		case fieldRecipientMessage:
			tx.RecipientMessage = col.Value
		case fieldType:
			tx.Type = col.Value
		case fieldSpecification:
			tx.Specification = col.Value
		case fieldComment:
			tx.Comment = col.Value
		case fieldBIC:
			tx.BIC = col.Value
		case fieldOrderID:
			tx.OrderID = col.Value
		case fieldPayerReference:
			tx.PayerReference = col.Value
		case fieldAuthor:
		default:
			return nil, fmt.Errorf(`unable to parse column: "%v"`, col.Name)
		}
	}
	return tx, nil
}

func parseAmount(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}

func parseInteger(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseGMTTime(s string) (time.Time, error) {
	return time.ParseInLocation(xmlTimeFormat, s, xmlGMTLocation)
}

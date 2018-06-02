package fio

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

const transactionsResponse = `
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<AccountStatement>
  <Info>
    <accountId>2501201133</accountId>
    <bankId>8330</bankId>
    <currency>EUR</currency>
    <iban>SK2383300000002501201133</iban>
    <bic>FIOZSKBAXXX</bic>
    <openingBalance>0.00</openingBalance>
    <closingBalance>45.97</closingBalance>
    <dateStart>2017-01-01+01:00</dateStart>
    <dateEnd>2017-05-01+02:00</dateEnd>
    <idFrom>13926601410</idFrom>
    <idTo>13926601410</idTo>
  </Info>
  <TransactionList>
    <Transaction>
      <column_22 name="ID pohybu" id="22">13926601410</column_22>
      <column_0 name="Datum" id="0">2017-04-11+02:00</column_0>
      <column_1 name="Objem" id="1">45.97</column_1>
      <column_14 name="Měna" id="14">EUR</column_14>
      <column_2 name="Protiúčet" id="2">SK2183100000001100248431</column_2>
      <column_10 name="Název protiúčtu" id="10">john doe</column_10>
      <column_12 name="Název banky" id="12">ZUNO BANK AG, pobočka zahraničnej banky</column_12>
      <column_4 name="KS" id="4">0558</column_4>
      <column_7 name="Uživatelská identifikace" id="7">john doe</column_7>
      <column_16 name="Zpráva pro příjemce" id="16">/DO2017-04-10/SPPrevod zo zuno, john doe</column_16>
      <column_8 name="Typ" id="8">Bezhotovostní příjem</column_8>
      <column_25 name="Komentář" id="25">john doe</column_25>
      <column_26 name="BIC" id="26">RIDBSKBXXXX</column_26>
      <column_17 name="ID pokynu" id="17">15689512949</column_17>
    </Transaction>
  </TransactionList>
</AccountStatement>
`

func TestByPeriod(t *testing.T) {
	setup()
	defer teardown()

	dateFrom := time.Now()
	dateTo := time.Now()
	urlStr := fmt.Sprintf("/ib_api/rest/periods/%v/%v/%v/transactions.xml", testingToken, fmtDate(dateFrom), fmtDate(dateTo))

	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, transactionsResponse)
	})

	opts := ByPeriodOptions{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}
	resp, err := client.Transactions.ByPeriod(context.Background(), opts)

	if assert.NoError(t, err, "Transactions.ByPeriod returned error") {
		openingBalance, _ := decimal.NewFromString("0.00")
		closingBalance, _ := decimal.NewFromString("45.97")

		want := &TransactionsResponse{
			Info: StatementInfo{
				AccountID:      2501201133,
				BankID:         "8330",
				Currency:       "EUR",
				IBAN:           "SK2383300000002501201133",
				BIC:            "FIOZSKBAXXX",
				OpeningBalance: openingBalance,
				ClosingBalance: closingBalance,
				DateStart:      time.Date(2017, time.January, 1, 0, 0, 0, 0, time.Local),
				DateEnd:        time.Date(2017, time.May, 1, 0, 0, 0, 0, time.Local),
				IDFrom:         13926601410,
				IDTo:           13926601410,
				IDLastDownload: 0,
				IDList:         0,
			},
			Transactions: []Transaction{
				{
					ID:                 13926601410,
					Date:               time.Date(2017, time.April, 11, 0, 0, 0, 0, time.Local),
					Amount:             closingBalance,
					Currency:           "EUR",
					Account:            "SK2183100000001100248431",
					AccountName:        "john doe",
					BankName:           "ZUNO BANK AG, pobočka zahraničnej banky",
					RecipientMessage:   "/DO2017-04-10/SPPrevod zo zuno, john doe",
					ConstantSymbol:     "0558",
					BIC:                "RIDBSKBXXXX",
					OrderID:            "15689512949",
					Comment:            "john doe",
					UserIdentification: "john doe",
					Type:               "Bezhotovostní příjem",
				},
			},
		}

		assert.Equal(t, resp, want)
	}
}

func TestExport(t *testing.T) {
	setup()
	defer teardown()

	testingFormat := JSONFormat
	dateFrom := time.Now()
	dateTo := time.Now()
	urlStr := fmt.Sprintf("/ib_api/rest/periods/%v/%v/%v/transactions.%v", testingToken, fmtDate(dateFrom), fmtDate(dateTo), testingFormat)

	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, transactionsResponse)
	})

	buf := new(bytes.Buffer)
	opts := ExportOptions{
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Format:   testingFormat,
	}
	err := client.Transactions.Export(context.Background(), opts, buf)

	if assert.NoError(t, err, "Transactions.Export returned error") {
		assert.Equal(t, len(transactionsResponse), buf.Len())
	}
}

func TestGetStatement(t *testing.T) {
	setup()
	defer teardown()

	year := 2017
	id := 1
	urlStr := fmt.Sprintf("/ib_api/rest/by-id/%v/%v/%v/transactions.xml", testingToken, year, id)

	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, transactionsResponse)
	})

	opts := GetStatementOptions{
		Year: year,
		ID:   id,
	}
	resp, err := client.Transactions.GetStatement(context.Background(), opts)

	if assert.NoError(t, err, "Transactions.GetStatement returned error") {
		openingBalance, _ := decimal.NewFromString("0.00")
		closingBalance, _ := decimal.NewFromString("45.97")

		want := &TransactionsResponse{
			Info: StatementInfo{
				AccountID:      2501201133,
				BankID:         "8330",
				Currency:       "EUR",
				IBAN:           "SK2383300000002501201133",
				BIC:            "FIOZSKBAXXX",
				OpeningBalance: openingBalance,
				ClosingBalance: closingBalance,
				DateStart:      time.Date(2017, time.January, 1, 0, 0, 0, 0, time.Local),
				DateEnd:        time.Date(2017, time.May, 1, 0, 0, 0, 0, time.Local),
				IDFrom:         13926601410,
				IDTo:           13926601410,
				IDLastDownload: 0,
				IDList:         0,
			},
			Transactions: []Transaction{
				{
					ID:                 13926601410,
					Date:               time.Date(2017, time.April, 11, 0, 0, 0, 0, time.Local),
					Amount:             closingBalance,
					Currency:           "EUR",
					Account:            "SK2183100000001100248431",
					AccountName:        "john doe",
					BankName:           "ZUNO BANK AG, pobočka zahraničnej banky",
					RecipientMessage:   "/DO2017-04-10/SPPrevod zo zuno, john doe",
					ConstantSymbol:     "0558",
					BIC:                "RIDBSKBXXXX",
					OrderID:            "15689512949",
					Comment:            "john doe",
					UserIdentification: "john doe",
					Type:               "Bezhotovostní příjem",
				},
			},
		}

		assert.Equal(t, resp, want)
	}
}

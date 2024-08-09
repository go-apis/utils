package tests

import (
	"bytes"
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-apis/utils/xsheet"
)

//go:embed test.xlsx
var testData []byte

type MyItem struct {
	Something   string            `sheet:"-"`
	Payout      string            `sheet:"payout;alias:payout_code,code,payout code"`
	Amount      string            `sheet:"amount"`
	Currency    string            `sheet:"currency"`
	FirstName   string            `sheet:"first_name;alias:first,first name"`
	LastName    string            `sheet:"last_name;alias:last,last name"`
	Line1       string            `sheet:"line_1;alias:address_line_1,address line 1,line 1"`
	Line2       string            `sheet:"line_2;alias:address_line_2,address line 2,line 2"`
	City        string            `sheet:"city;alias:address_city,address city"`
	State       string            `sheet:"state;alias:address_state,address state"`
	ZipCode     string            `sheet:"zip_code;alias:address_zip_code,address_zipcode,address zip code,address zipcode,zipcode,zip code"`
	Country     string            `sheet:"country;alias:address_country,address country"`
	Email       string            `sheet:"email"`
	PhoneNumber string            `sheet:"phone_number;alias:phone number,phone"`
	Description string            `sheet:"description"`
	RowNumber   int               `sheet:";rownumber"`
	Metadata    map[string]string `sheet:";extra"`
}

func Test_It(t *testing.T) {
	ctx := context.TODO()
	fileName := "test.xlsx"
	fileType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	reader := bytes.NewReader(testData)
	output := []*MyItem{
		{
			RowNumber:   1,
			Payout:      "Test1",
			Amount:      "85",
			Currency:    "USD",
			FirstName:   "Demo",
			LastName:    "Account",
			Line1:       "11 None Street",
			City:        "Newland",
			State:       "NC",
			ZipCode:     "28657",
			Country:     "United States",
			Email:       "my@email.com",
			PhoneNumber: "0299999999",
			Description: "Description which will show up places",
			Metadata: map[string]string{
				"My": "a",
				"Id": "2",
			},
		},
		{
			RowNumber:   2,
			Payout:      "Test2",
			Amount:      "30",
			Currency:    "USD",
			FirstName:   "Test",
			LastName:    "User",
			Line1:       "22 None Street",
			City:        "Portland",
			State:       "OR",
			ZipCode:     "97212",
			Country:     "United States",
			Email:       "other@email.com",
			PhoneNumber: "0288888888",
			Description: "Demo",
			Metadata: map[string]string{
				"My": "b",
				"Id": "3",
			},
		},
	}

	p, err := xsheet.NewParser[MyItem]()
	require.NoError(t, err)

	items, err := p.Parse(ctx, fileName, fileType, reader)
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.ElementsMatch(t, items, output, "items are incorrect")
}

func Test_Props(t *testing.T) {
	props, err := xsheet.NewProps[MyItem]()
	require.NoError(t, err)
	require.NotNil(t, props)
}

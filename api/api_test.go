package api

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetData(t *testing.T) {
	expectedBytes, err := ioutil.ReadFile("./tests-fixtures/expected_output.json")
	if err != nil {
		t.Fatal("unable to read expected_output file: ", err)
	}
	expected := []Data{}
	err = json.Unmarshal(expectedBytes, &expected)
	if err != nil {
		t.Fatal("Unable to unmarshal expected data: ", err)
	}
	service := New()

	actual, err := service.GetData()
	if err != nil {
		t.Fatal("Unexpected error in process: ", err)
	}
	if len(expected) != len(actual) {
		t.Fatal("Length missmatch between expected and actual data")
	}
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i].Name, actual[i].Name)
		//assert.Equal(t, expected[i].Price, actual[i].Price)
		//assert.Equal(t, expected[i].CashDiscount, actual[i].CashDiscount)
		//assert.Equal(t, expected[i].Category, actual[i].Category)
		//assert.Equal(t, expected[i].CategoryImportance, actual[i].CategoryImportance)
		//assert.Equal(t, expected[i].Rate, actual[i].Rate)
		//assert.Equal(t, expected[i].Discount, actual[i].Discount)
	}
}

package api

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	"github.com/canastto/client"
)

const (
	firstEndpoint = "https://run.mocky.io/v3/77f7e692-73f3-4676-a4ce-8576dd99ca0c"
	secondEnpoint = "https://run.mocky.io/v3/26029c20-0eb4-43b1-b8ba-871384052fc7"

	// relevance rates
	rateRelevance               = 0.3
	cashDiscountRelevance       = 0.5
	categoryImportanceRelevance = 0.2
)

var (
	c client.Client
)

// Data is the result struct of the problem
type Data struct {
	Name               string  `json:"name"`
	Price              int     `json:"price"`
	Discount           int     `json:"discount"`
	CashDiscount       int     `json:"cash_discount"`
	Rate               int     `json:"rate"`
	Category           string  `json:"category"`
	CategoryImportance int     `json:"category_importance"`
	Relevance          float64 `json:"relevance"`
}

// API interface to retrieve data and enable dependency injection
type API interface {
	GetData() ([]Data, error)
}

type api struct{}

type firstEndpointData []struct {
	Category   string `json:"categoria"`
	Importance int    `json:"importancia"`
	Products   []struct {
		Name      string `json:"nombre"`
		Price     string `json:"precio"`
		HighPrice string `json:"precio_alto"`
		Rate      int    `json:"calificacion"`
	} `json:"productos"`
}

type secondEndpointData struct {
	Categories []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Importance int    `json:"importance"`
	} `json:"categories"`
	Products []struct {
		ProductData struct {
			Name       string `json:"name"`
			Price      int    `json:"price"`
			Rate       int    `json:"rate"`
			Discount   int    `json:"discount"`
			Categories [1]struct {
				ID int `json:"category_id"`
			} `json:"categories"`
			Stock int `json:"stock"`
		} `json:"product_data"`
	} `json:"products"`
}

// New creates a default API interface
func New() API {
	if c == (client.Client)(nil) {
		c = client.New()
	}
	return &api{}
}

func cleanPrice(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(strings.ReplaceAll(strings.ReplaceAll(s, "$", ""), ".", ""))
}

func fixFirstEndpointResponse(result *firstEndpointData) ([]Data, error) {
	var data = []Data{}
	for _, res := range *result {
		for _, product := range res.Products {
			d := Data{
				Name:               product.Name,
				Rate:               product.Rate,
				Category:           res.Category,
				CategoryImportance: res.Importance,
			}
			// clean product price and high price
			price, err := cleanPrice(product.Price)
			if err != nil {
				return nil, err
			}
			highPrice, err := cleanPrice(product.HighPrice)
			if err != nil {
				return nil, err
			}

			d.Price = price
			if highPrice != 0 {
				d.CashDiscount = highPrice - price
				d.Discount = int(float64(d.CashDiscount) / float64(highPrice) * 100)
			}

			data = append(data, d)
		}
	}
	return data, nil
}

func fixSecondEndpointResponse(result *secondEndpointData) []Data {
	var data = []Data{}
	// helper is a hashmap that maps category id to idx in the result categories
	var helper = make(map[int]int)
	for i := 0; i < len(result.Categories); i++ {
		helper[result.Categories[i].ID] = i
	}
	for _, product := range result.Products {
		categoryIdx := helper[product.ProductData.Categories[0].ID]

		d := Data{
			Name:               product.ProductData.Name,
			Price:              product.ProductData.Price,
			Discount:           product.ProductData.Discount,
			CashDiscount:       int(float64(product.ProductData.Price) * float64(product.ProductData.Discount) / float64(100)),
			Category:           result.Categories[categoryIdx].Name,
			CategoryImportance: result.Categories[categoryIdx].Importance,
			Rate:               product.ProductData.Rate,
		}
		data = append(data, d)
	}

	return data
}

func getFirstData() ([]Data, error) {
	resp, err := c.Get(firstEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := &firstEndpointData{}
	err = json.Unmarshal(b, result)
	if err != nil {
		return nil, err
	}

	return fixFirstEndpointResponse(result)
}

func getSecondData() ([]Data, error) {
	resp, err := c.Get(secondEnpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := &secondEndpointData{}
	err = json.Unmarshal(b, result)
	if err != nil {
		return nil, err
	}

	return fixSecondEndpointResponse(result), nil
}

func computeRelevance(data []Data) {

	var (
		maxRate               int
		minRate               int
		maxCashDiscount       int
		minCashDiscount       int
		maxCategoryImportance int
		minCategoryImportance int
	)
	minRate = data[0].Rate
	minCashDiscount = data[0].CashDiscount
	minCategoryImportance = data[0].CategoryImportance

	for _, d := range data {
		if d.Rate > maxRate {
			maxRate = d.Rate
		}
		if d.Rate < minRate {
			minRate = d.Rate
		}
		if d.CashDiscount > maxCashDiscount {
			maxCashDiscount = d.CashDiscount
		}
		if d.CashDiscount < minCashDiscount {
			minCashDiscount = d.CashDiscount
		}
		if d.CategoryImportance > maxCategoryImportance {
			maxCategoryImportance = d.CategoryImportance
		}
		if d.CategoryImportance < minCategoryImportance {
			minCategoryImportance = d.CategoryImportance
		}
	}

	for i, d := range data {
		relevance := 0.0
		relevance += float64(d.Rate) / float64(maxRate) * rateRelevance
		relevance += float64(d.CashDiscount) / float64(maxCashDiscount) * cashDiscountRelevance
		relevance += float64(d.CategoryImportance) / float64(maxCategoryImportance) * categoryImportanceRelevance
		// working with lists in go is like working with pointers, which enables us to
		// write relevance directly into the memory address by accessing the list element
		data[i].Relevance = relevance
	}
}

// GetData queries both endpoints and retrieves the data
func (a *api) GetData() ([]Data, error) {
	// for future releases the functions to get data from the first
	// and second enpoint could be executed concurrently
	firstEndpointData, err := getFirstData()
	if err != nil {
		return nil, err
	}
	secondEndpointData, err := getSecondData()
	if err != nil {
		return nil, err
	}

	// merge endpoints' data response and sort by relevance
	data := append(firstEndpointData, secondEndpointData...)
	computeRelevance(data)

	sort.SliceStable(data, func(i, j int) bool {
		return data[i].Relevance > data[j].Relevance
	})

	return data, nil
}

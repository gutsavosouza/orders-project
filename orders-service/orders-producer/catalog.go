package main

var catalog = map[string]float64{
	"SKU-1001": 19.99,
	"SKU-1002": 29.99,
	"SKU-1003": 9.27,
	"SKU-1004": 14.99,
	"SKU-1005": 24.99,
	"SKU-1006": 39.99,
	"SKU-1007": 49.49,
	"SKU-1008": 59.99,
	"SKU-1009": 4.99,
	"SKU-1010": 12.99,
	"SKU-3954": 99.99,
	"SKU-3955": 149.99,
	"SKU-2044": 49.50,
	"SKU-2045": 79.45,
	"SKU-2046": 29.95,
	"SKU-2047": 19.95,
	"SKU-2048": 9.75,
	"SKU-2049": 14.50,
	"SKU-2050": 24.49,
}

func getRandomSKU() string {
	if len(catalog) == 0 {
		return ""
	}

	for k, _ := range catalog {
		return k
	}
	return ""
}

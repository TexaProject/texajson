package texajson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Page struct {
	AIName   string  `json:"AIName"`
	IntName  string  `json:"IntName"`
	ArtiMts  float64 `json:"ArtiMts"`
	HumanMts float64 `json:"HumanMts"`
}

// ToString returns the string equivalent JSON format of Page
func (p Page) ToString() string {
	return ToJson(p)
}

//GetPages returns a converted Page Array persistent to the mts.json
func GetPages() []Page {
	raw, err := ioutil.ReadFile("./www/data/mts.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var c []Page
	json.Unmarshal(raw, &c)
	return c
}

// ConvtoPage converts a set of data vars into a Page struct variable
func ConvtoPage(AIName string, IntName string, ArtiMts float64, HumanMts float64) Page {
	var newPage Page
	newPage.AIName = AIName
	newPage.IntName = IntName
	newPage.ArtiMts = ArtiMts
	newPage.HumanMts = HumanMts
	return newPage
}

// AddtoPageArray Appends a new page 'p' to the specified target PageArray 'pa'
func AddtoPageArray(p Page, pa []Page) []Page {
	for x := range pa {
		if p == pa[x] {
			panic("JSON ERROR: Can't append a Duplicate Page into PageArray")
		}
	}
	return (append(pa, p))
}

// ToJson marshals PageArray data into JSON format
func ToJson(p interface{}) string {
	bytes, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	ioutil.WriteFile("./www/data/mts.json", bytes, 0644)
	return string(bytes)
}

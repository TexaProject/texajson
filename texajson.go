package texajson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// SlabPage exports schema for reportcard/mts.json
type SlabPage struct {
	SlabName     string `json:"SlabName"`
	NQDropped    int    `json:"NQDropped"`
	AvgSlabSize  int    `json:"AvgSlabSize"`
	NSlabExposed int    `json:"NSlabExposed"`
}

// ToString returns the string equivalent JSON format of SlabPage
func (p SlabPage) ToString() string {
	fmt.Println("###SlabToString()")
	return ToJson(p)
}

//GetSlabPages returns a converted SlabPage Array persistent to the mts.json
func GetSlabPages() []SlabPage {
	fmt.Println("###GetSlabPages()")
	raw, err := ioutil.ReadFile("./www/data/reportcard/slab.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var c []SlabPage
	json.Unmarshal(raw, &c)
	return c
}

func dupCount(list []string) map[string]int {
	fmt.Println("###dupCount()")
	duplicate_frequency := make(map[string]int)

	for _, item := range list {
		// check if the item/element exist in the duplicate_frequency map

		_, exist := duplicate_frequency[item]

		if exist {
			duplicate_frequency[item] += 1 // increase counter by 1 if already in the map
		} else {
			duplicate_frequency[item] = 1 // else start counting from 1
		}
	}
	return duplicate_frequency
}

// ConvtoSlabPage configures the parameters of SlabPage using the ArtiQSA
func ConvtoSlabPage(ArtiQSA []uint64, SlabNameArray []string, slabSeqArray []string) []SlabPage {
	fmt.Println("###ConvtoSlabPage()")
	sp := make([]SlabPage, len(SlabNameArray))

	// Initialization for all Slabs mentioned in the SlabNameArray
	for k := 0; k < len(SlabNameArray); k++ {
		sp[k].SlabName = SlabNameArray[k]
		sp[k].NQDropped = 0
		sp[k].AvgSlabSize = 0
		sp[k].NSlabExposed = 0
	}
	fmt.Println("###sp")
	fmt.Println(sp)

	for i := 0; i < len(ArtiQSA); i++ {
		if ArtiQSA[i] == 0 {
			for k := 0; k < len(SlabNameArray); k++ {
				if sp[k].SlabName == slabSeqArray[i] {
					sp[k].NQDropped++
				}
			}
		}
	}
	fmt.Println("###sp-postNQDropped")
	fmt.Println(sp)

	fmt.Println("###dupMap")
	dupMap := dupCount(slabSeqArray)
	for k, v := range dupMap {
		for x := 0; x < len(SlabNameArray); x++ {
			if k == SlabNameArray[x] {
				if v >= 1 {
					sp[x].AvgSlabSize = (sp[x].AvgSlabSize*sp[x].NSlabExposed + v) / (sp[x].NSlabExposed + 1)
					sp[x].NSlabExposed++
				}
			}
		}
	}

	// for k := 0; k < len(SlabNameArray); k++ {
	// 	sizeCounter := 0
	// 	for j := 0; j < len(slabSeqArray); j++ {
	// 		if SlabNameArray[k] == slabSeqArray[j] {
	// 			sizeCounter++
	// 		}
	// 	}

	// }

	// for i := 0; i < len(ArtiQSA); i++ {
	// 	for k := 0; k < len(SlabNameArray); k++ {
	// 		if sp[k].SlabName == slabSeqArray[i] {

	// 		}
	// 	}
	// }

	// for i := 0; i < len(ArtiQSA); i++ {
	// 	for k := 0; k < len(SlabNameArray); k++ {
	// 		if sp[k].SlabName == slabSeqArray[i] {

	// 		}
	// 	}
	// }
	return (sp)
}

// AddtoSlabPageArray Appends a new Slabpage 'p' to the specified target SlabPageArray 'pa'
func AddtoSlabPageArray(p SlabPage, pa []SlabPage) []SlabPage {
	fmt.Println("###AddtoSlabPageArray()")
	for x := 0; x < len(pa); x++ {
		if p.SlabName == pa[x].SlabName {
			// panic("JSON ERROR: Can't append a Duplicate SlabPage into SlabPageArray")
			pa[x].NSlabExposed += p.NSlabExposed
			pa[x].NQDropped += p.NQDropped
			pa[x].AvgSlabSize = (pa[x].AvgSlabSize + p.AvgSlabSize) / pa[x].NSlabExposed
			return pa
		}

	}
	return (append(pa, p))
}

// SlabToJson marshals SlabPageArray data into JSON format
func SlabToJson(p interface{}) string {
	fmt.Println("###SlabToJson()")
	bytes, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	ioutil.WriteFile("./www/data/reportcard/slab.json", bytes, 0644)
	return string(bytes)
}

// Page exports schema for mts.json
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
		if (p.AIName == pa[x].AIName) && (p.IntName == pa[x].IntName) {
			pa[x].ArtiMts = p.ArtiMts
			pa[x].HumanMts = p.HumanMts
			// panic("JSON ERROR: Can't append a Duplicate Page into PageArray")
			return pa
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

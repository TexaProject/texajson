package texajson

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sync"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/ipfs-cluster/api"

	ipldcrud "github.com/0zAND1z/ipldcrud"
	"github.com/go-redis/redis"
	goCid "github.com/ipfs/go-cid"
	"github.com/ipfs/ipfs-cluster/api/rest/client"
)

var SlabTempSize []int
var SlabTempNQD []int
var RedisClient *redis.Client
var sh *shell.Shell
var instance *IpfsCluster
var once sync.Once

var (
	Cat  = "Cat"
	Mts  = "Mts"
	Slab = "Slab"
)

// IpfsCluster ... Client for IPFS-CLUSTER
type IpfsCluster struct {
	cluster client.Client
}

// GetInstance ... returns an IPFS cluster client
func GetInstance() *IpfsCluster {
	return instance
}

// InitCluster ... Initialize ipfs-cluster client
func InitCluster(host, port string) {
	once.Do(func() {
		clusterClient, err := client.NewDefaultClient(&client.Config{
			Host:     host,
			Port:     port,
			LogLevel: "info",
		})
		if err != nil {
			log.Fatal("InitCluster(): Unable to initiate ipfs-cluster", err)
		}
		instance = &IpfsCluster{cluster: clusterClient}
	})
}

// PinCid .. Pins a cid into the ipfs-cluster network
func (c IpfsCluster) PinCid(cid string) (retVal *api.Pin, err error) {
	cidToPin, err := goCid.Parse(cid)
	if err != nil {
		return
	}
	retVal, err = c.cluster.Pin(context.Background(), cidToPin, api.DefaultAddParams().PinOptions)
	if err != nil {
		return
	}
	return
}

// CatValArray exports the sub-JSON document for CatPage
type CatValArray struct {
	CatName     string        `json:"CatName"`
	Spf         float64       `json:"Spf"`
	Interaction []Interaction `json:"Interaction"`
}

// CatPage exports schema for data/cat.json
type CatPage struct {
	AIName string        `json:"AIName"`
	CatVal []CatValArray `json:"CatVal"`
}

// ToString returns the string equivalent JSON format of CatPage
func (p CatPage) ToString() string {
	return ToJson(p)
}

// Interaction represents communication b/w Ai and
// human and its quantum score along with Justification.
type Interaction struct {
	HumanTransaction string `json:"HumanTransaction"`
	AiTransaction    string `json:"AiTransaction"`
	QuantumScore     uint64 `json:"QuantumScore"`
	Justification    string `json:"Justification"`
}

func (i Interaction) ToString() string {
	return ToJson(i)
}

func init() {
	//make redis connection
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	result, err := RedisClient.Ping().Result()
	if err != nil {
		panic("Err Connecting to Redis")
	} else {
		fmt.Println("Connected to Redis", result)
	}
	// Setup IPFS client
	sh := ipldcrud.InitShell("https://ipfs.infura.io:5001")
	if sh == nil {
		log.Fatalln("init(): Unable to initiate IPFS shell.")
	}
}

//GetCatPages returns a converted CatPage Array persistent to the mts.json
func GetCatPages() []CatPage {
	raw, err := RedisClient.Get(Cat).Result()
	if err != nil && err.Error() != "redis: nil" {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	c := []CatPage{}
	json.Unmarshal([]byte(raw), &c)
	return c
}

// ConvtoCatPage converts a set of data vars into a CatPage struct variable
func ConvtoCatPage(AIName string, slabPageArray []SlabPage, SlabNameArray []string) CatPage {
	var newCatPage CatPage
	newCatPage.AIName = AIName
	cv := make([]CatValArray, len(SlabNameArray))

	fmt.Println("####SlabTempSize")
	fmt.Println(SlabTempSize)

	for index := 0; index < len(slabPageArray); index++ {
		for n := 0; n < len(SlabNameArray); n++ {
			if SlabNameArray[n] == slabPageArray[index].SlabName {
				cv[n].CatName = slabPageArray[index].SlabName

				// Adding Interaction array
				cv[n].Interaction = slabPageArray[index].Interactions

				ef := (float64(slabPageArray[index].NQDropped) / float64(slabPageArray[index].AvgSlabSize))
				rf := (float64(SlabTempSize[n]-SlabTempNQD[n]) / float64(SlabTempSize[n]))
				// cv[index].Spf = ((float64(SlabTempSize[n]-slabPageArray[index].NQDropped) / float64(SlabTempSize[n])) / (float64(slabPageArray[index].NQDropped) / float64(slabPageArray[index].AvgSlabSize)))
				spfTemp := rf / ef
				if math.IsInf(spfTemp, 0) {
					cv[n].Spf = 999
				} else {
					cv[n].Spf = spfTemp
				}
			}
		}
	}

	fmt.Println("####cv")
	fmt.Println(cv)

	newCatPage.CatVal = cv
	return newCatPage
}

// AddtoCatPageArray Appends a new CatPage 'p' to the specified target CatPageArray 'pa'
func AddtoCatPageArray(p CatPage, pa []CatPage) []CatPage {
	for index := 0; index < len(pa); index++ {
		if pa[index].AIName == p.AIName {
			for a := 0; a < len(p.CatVal); a++ {
				for m := 0; m < len(pa[index].CatVal); m++ {
					if p.CatVal[a].CatName == pa[index].CatVal[m].CatName {
						pa[index].CatVal[m].Spf = p.CatVal[a].Spf
					}
				}
				pa[index].CatVal = append(pa[index].CatVal, p.CatVal[a])
				// pa[index].CatVal = append(pa[index].CatVal, p.CatVal[a])
			}
			return pa
		}
	}
	return (append(pa, p))
}

// CatToJson marshals CatPageArray data into JSON format
func CatToJson(p interface{}) string {
	bytes, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	strBytes := string(bytes)

	err = RedisClient.Set(Cat, strBytes, 0).Err()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	SlabTempSize = nil
	SlabTempNQD = nil

	return strBytes
}

// SlabPage exports schema for reportcard/mts.json
type SlabPage struct {
	SlabName     string        `json:"SlabName"`
	NQDropped    int           `json:"NQDropped"`
	AvgSlabSize  int           `json:"AvgSlabSize"`
	NSlabExposed int           `json:"NSlabExposed"`
	Interactions []Interaction `json:"Interaction"`
}

// ToString returns the string equivalent JSON format of SlabPage
func (p SlabPage) ToString() string {
	fmt.Println("###SlabToString()")
	return ToJson(p)
}

//GetSlabPages returns a converted SlabPage Array persistent to the mts.json
func GetSlabPages() []SlabPage {
	fmt.Println("###GetSlabPages()")
	raw, err := RedisClient.Get(Slab).Result()
	if err != nil && err.Error() != "redis: nil" {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	c := []SlabPage{}
	json.Unmarshal([]byte(raw), &c)
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
func ConvtoSlabPage(ArtiQSA []uint64, SlabNameArray []string, slabSeqArray []string, justifications []string, transactions []string) []SlabPage {
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

	SlabTempNQD = make([]int, len(SlabNameArray))
	transactionCounter := 0
	for i := 0; i < len(ArtiQSA); i++ {
		if ArtiQSA[i] == 0 {
			for k := 0; k < len(SlabNameArray); k++ {
				if sp[k].SlabName == slabSeqArray[i] {
					SlabTempNQD[k]++
					sp[k].NQDropped++
				}
			}
		}

		for j := 0; j < len(SlabNameArray); j++ {
			var interaction Interaction
			if sp[j].SlabName == slabSeqArray[i] {
				if i == 0 {
					interaction = Interaction{HumanTransaction: "", AiTransaction: transactions[transactionCounter], QuantumScore: ArtiQSA[i], Justification: justifications[i]}
					transactionCounter++
					fmt.Println("transaction index", transactionCounter)
				} else {
					hTransaction := transactions[transactionCounter]
					transactionCounter++
					aiTransaction := transactions[transactionCounter]
					transactionCounter++
					interaction = Interaction{HumanTransaction: hTransaction, AiTransaction: aiTransaction, QuantumScore: ArtiQSA[i], Justification: justifications[i]}
					fmt.Println(transactionCounter)
				}
				if sp[j].Interactions == nil {
					sp[j].Interactions = make([]Interaction, 0)
				}
				sp[j].Interactions = append(sp[j].Interactions, interaction)
			}
		}
	}
	fmt.Println("###sp-postNQDropped")
	fmt.Println(sp)

	fmt.Println("###dupMap")
	dupMap := dupCount(slabSeqArray)
	SlabTempSize = make([]int, len(SlabNameArray))
	for k, v := range dupMap {
		for x := 0; x < len(SlabNameArray); x++ {
			if k == SlabNameArray[x] {
				if v >= 1 {
					SlabTempSize[x] = v
					sp[x].AvgSlabSize = (sp[x].AvgSlabSize*sp[x].NSlabExposed + v) / (sp[x].NSlabExposed + 1)
					sp[x].NSlabExposed++
				}
			}
		}
	}
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
			pa[x].Interactions = p.Interactions
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
	strBytes := string(bytes)
	err = RedisClient.Set(Slab, strBytes, 0).Err()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return strBytes
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
	raw, err := RedisClient.Get(Mts).Result()
	if err != nil && err.Error() != "redis: nil" {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	c := []Page{}
	json.Unmarshal([]byte(raw), &c)
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
	strBytes := string(bytes)
	err = RedisClient.Set(Mts, strBytes, 0).Err()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return strBytes
}

// Result is the master table recording the results of all test sessions for a given AI
type Result struct {
	AIName         string          `json:"AIName"`
	Interrogations []Interrogation `json:"Interrogations"`
}

// Interrogation is used to record the data from a session
type Interrogation struct {
	IntName  string        `json:"IntName"`
	ArtiMts  float64       `json:"ArtiMts"`
	HumanMts float64       `json:"HumanMts"`
	CatVal   []CatValArray `json:"CatVal"`
}

// NewResultObject is used to create a new Result object for a new AI
func NewResultObject(aiName string) Result {
	return Result{
		AIName:         aiName,
		Interrogations: []Interrogation{},
	}
}

// NewInterrogationObject is created a new object and returns it
func NewInterrogationObject(IntName string, ArtiMts, HumanMts float64, CatVal []CatValArray) Interrogation {
	return Interrogation{
		IntName:  IntName,
		ArtiMts:  ArtiMts,
		HumanMts: HumanMts,
		CatVal:   CatVal,
	}
}

// WriteDataToIPFS is used to write a data to IPFS using ipldcrud and return the CID
func WriteDataToIPFS(ipfsURL string, data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Println("WriteDataToIPFS(): Issue in marshaling data!")
	}
	sh := ipldcrud.InitShell(ipfsURL) // Can be replaced with any hosted IPFS API URL also. Example: "https://ipfs.infura.io:5001" or "http://localhost:5001"
	resultCid := ipldcrud.Set(sh, bytes)
	fmt.Println("WriteDataToIPFS(): Results of this testing session are globally accessible at https://explore.ipld.io/#/explore/" + resultCid)
	fmt.Println("WriteDataToIPFS(): You can also access them locally through ipld-explorer at http://localhost:3000/#/explore/" + resultCid)
	return resultCid
}

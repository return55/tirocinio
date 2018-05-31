package webDriver

import (
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tirocinio/structures"
	"os"
	"regexp"
	"strconv"
	"strings"
	"encoding/gob"
)

const (
	seleniumPath    = "webDriver/selenium-server-standalone-3.11.0.jar"
	geckoDriverPath = "webDriver/geckodriver-v0.20.1-linux64/geckodriver"
	port            = 8080
)

func StartSelenium() (*selenium.Service, selenium.WebDriver) {
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),            // Output debug information to STDERR.
	}

	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)

	if err != nil {
		panic(err)
	}

	selenium.SetDebug(false)

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}

	return service, wd
}

func GetInitialDocument(service *selenium.Service, wd selenium.WebDriver, startURL string) structures.Document {
	if err := wd.Get(startURL); err != nil {
		panic(err)
	}
	textBox, err := wd.FindElement(selenium.ByID, "gs_hdr_tsi")
	if err != nil {
		panic(err)
	}
	if err := textBox.SendKeys(`TCP performance`); err != nil {
		panic(err)
	}
	searchButton, err := wd.FindElement(selenium.ByID, "gs_hdr_tsb")
	if err != nil {
		panic(err)
	}
	if err := searchButton.Click(); err != nil {
		panic(err)
	}
	//---------------------------------cerco le info sul documento-------------------------------------------
	var initialDoc structures.Document
	var text string

	urls, err := wd.FindElements(selenium.ByXPATH, "//div/h3/a")
	if err != nil {
		panic(err)
	}
	initialDoc.Url, _ = urls[0].GetAttribute("href")

	authors, err := wd.FindElements(selenium.ByXPATH, "//div[@class='gs_a']")
	if err != nil {
		panic(err)
	}
	//estrai dalla stringa authors i nomi degli autori (non imprta se ci sono tutti)
	//esempio di authors[0] : "H Balakrishnan, VN Padmanabhan… - … ACM transactions on …, 1997 - ieeexplore.ieee.org"
	text, _ = authors[0].Text()
	leftSide := strings.Split(text, " -")[0]
	//!!!il carattere ... che segue gli autori corrisponde a â in utf-8!!!!
	leftSide = strings.Replace(leftSide, "\u2026", "", -1)
	initialDoc.Authors = strings.Split(leftSide, ", ")

	//altro = citato da + related + versioni
	other, err := wd.FindElements(selenium.ByXPATH, "//div[@class='gs_fl']/a")
	if err != nil {
		panic("sto cercando i link di: citato da + related + versioni")
	}

	text, _ = other[2].Text()
	words := strings.Split(text, " ")
	if numCitedBy, err := strconv.ParseUint(words[2], 10, 16); err != nil {
		panic(err)
	} else {
		initialDoc.NumCitedBy = uint16(numCitedBy)
	}

	linkCitedBy, _ := other[2].GetAttribute("href")
	initialDoc.LinkCitedBy = structures.URLScholar + linkCitedBy

	return initialDoc
}

//--------------------------------------------------------------------------------------------------
func GetCiteDocument(service *selenium.Service, wd selenium.WebDriver, initialDoc structures.Document) []structures.Document {
	if err := wd.Get(initialDoc.LinkCitedBy); err != nil {
		panic(err)
	} /*--------------------non mi serve--------------
	textBox, err := wd.FindElement(selenium.ByID, "gs_hdr_tsi")
	if err != nil {
		panic(err)
	}
	if err := textBox.SendKeys(`TCP performance`); err != nil{
		panic(err)
	}
	searchButton, err:= wd.FindElement(selenium.ByID, "gs_hdr_tsb")
	if err != nil{
		panic(err)
	}
	if err:=searchButton.Click(); err!=nil {
		panic(err)
	}*/
	//---------------------------------cerco le info sul documento-------------------------------------------
	citeInitialDoc := make([]structures.Document, 10 /*initialDoc.NumCitedBy*/)
	var text string

	urls, err := wd.FindElements(selenium.ByXPATH, "//div/h3/a")
	if err != nil {
		panic(err)
	}

	authors, err := wd.FindElements(selenium.ByXPATH, "//div[@class='gs_a']")
	if err != nil {
		panic(err)
	}

	other, err := wd.FindElements(selenium.ByXPATH, "//div[@class='gs_fl']/a")
	if err != nil {
		panic("sto cercando i link di: citato da + related + versioni")
	}
	fmt.Println(initialDoc.NumCitedBy)
	var docIndex, otherIndex uint16
	for docIndex, otherIndex = 0, 0; docIndex < 10; /*initialDoc.NumCitedBy*/ docIndex, otherIndex = docIndex+1, otherIndex+1 {
		fmt.Println(docIndex, "------", otherIndex)
		citeInitialDoc[docIndex].Url, _ = urls[docIndex].GetAttribute("href")

		text, _ = authors[docIndex].Text()
		leftSide := strings.Split(text, " -")[0]
		//!!!il carattere ... che segue gli autori corrisponde a â in utf-8!!!!  NON FUNZIONA
		leftSide = strings.Replace(leftSide, "\u2026", "", -1)
		citeInitialDoc[docIndex].Authors = strings.Split(leftSide, ", ")

		//scorro other, mi fermo quando trovo un match con citato da cosi' so che sono sull'elemento giusto
		for ; ; otherIndex++ {
			text, _ = other[otherIndex].Text()
			if t, _ := regexp.MatchString("Citato da.*", text); t {
				words := strings.Split(text, " ")
				if numCitedBy, err := strconv.ParseUint(words[2], 10, 16); err != nil {
					panic(err)
				} else {
					citeInitialDoc[docIndex].NumCitedBy = uint16(numCitedBy)
				}

				linkCitedBy, _ := other[otherIndex].GetAttribute("href")
				citeInitialDoc[docIndex].LinkCitedBy = structures.URLScholar + linkCitedBy
				break
			}
		}
	}
	return citeInitialDoc
}

//modifica perche' riceva un unico array di document
func PrintDocuments(allDoc[] structures.Document) {
	if len(allDoc)==0 {
		fmt.Println("Non ci sono documenti da stampare")
		return
	}
	fmt.Println("Documento iniziale:\nUrl: ", allDoc[0].Url, "\nAutori:")
	for _, autore := range allDoc[0].Authors {
		fmt.Println("\t", autore)
	}
	fmt.Println("Numero di documenti che lo citano: ", allDoc[0].NumCitedBy)
	fmt.Println("Link ai documenti che lo citano: ", allDoc[0].LinkCitedBy)

	fmt.Println("\nDocumento che citano:")
	for docIndex:=1; docIndex<len(allDoc);  docIndex++ {
		fmt.Println("Url: ", allDoc[docIndex].Url, "\nAutori:")
		for _, autore := range allDoc[docIndex].Authors {
			fmt.Println("\t", autore)
		}
		fmt.Println("Numero di documenti che lo citano: ", allDoc[docIndex].NumCitedBy)
		fmt.Println("Link ai documenti che lo citano: ", allDoc[docIndex].LinkCitedBy)
	}
}

//salvo i documenti su un file
func SaveDocuments(allDoc[] structures.Document){
	file, err := os.Create(structures.SaveFilePath)
	if err!=nil{
		panic(err)
	}
	defer file.Close()
	
	enc := gob.NewEncoder(file)
	enc.Encode(allDoc)
}

//carico i documenti da file
func LoadDocuments(allDoc[] structures.Document){
	allDoc = nil
	file, err := os.Open(structures.SaveFilePath)
	if err!=nil{
		panic(err)
	}
	defer file.Close()
	
	dec := gob.NewDecoder(file)
	dec.Decode(allDoc)
}

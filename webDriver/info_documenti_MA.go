package webDriver

import (
	"fmt"

	"github.com/return55/tirocinio/structures"

	"github.com/tebeka/selenium"
)

//Informazioni che Raccolgo dai Documenti:
// Url         string
// Authors     []string
// NumCitedBy  uint16
// LinkCitedBy string
// Abstract    string
// Date 		time.Time  //Date=time.Date(anno,mese,giorno,0,0,0,0,time.UTC) //year, month, day := d.Date()
// FieldsOfStudy	[]string

//Data un pagina (impostata dal WebDriver) prendo un certo numero di documenti dalla pagina
//partendo dal primo in alto.
//Se il numero (numDocs) e' maggiore del numero di documenti nella pagina (tipicamente 8),
//mi limito a restituire i documenti presenti nella pagina e la loro quantita'.
func GetDocumentsFromPage_MA(wd selenium.WebDriver, numDocs uint64) ([]structures.Document, uint16) {
	//scorro i link ai documenti presenti nella pagina
	links, err := wd.FindElements(selenium.ByXPATH, "//article")
	if err != nil {
		panic(err)
	}
	fmt.Println("Lunghezza: ", len(links))
	for _, link := range links {
		url, err := link.TagName()
		if err != nil {
			fmt.Println("---------------------------------")
		}
		fmt.Println(url)
	}
	return nil, 0
	/*
			url, err := link.GetAttribute("href")
			if err != nil {
				panic(err)
			}
			//vado alla pagina del documento
			if err := wd.Get(structures.URLAcademic + url); err != nil {
				panic(err)
			}

			//raccolgo le informazioni
			sources, err := wd.FindElements(selenium.ByXPATH, "//div/h3/a")
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

			//imposto il valore del numero di documenti nella pagina in base a quanti
			//url di documenti ho letto
			docInThePage := uint64(len(urls))
			//numero di documenti che effettivamente leggo: numDocs = min(docInThePage, numDocs)
			if docInThePage < numDocs {
				numDocs = docInThePage
			}
		}

		documents := make([]structures.Document, numDocs)
		var text string
		var docIndex, otherIndex uint64
		//stampa-*---------------------------------------------
		fmt.Println("Lunghezza: ", len(urls))
		for docIndex, otherIndex = 0, 0; docIndex < numDocs; docIndex, otherIndex = docIndex+1, otherIndex+1 {
			//fmt.Println(docIndex, "------", otherIndex)
			documents[docIndex].Url, _ = urls[docIndex].GetAttribute("href")

			text, _ = authors[docIndex].Text()
			leftSide := strings.Split(text, " -")[0]
			//!!!il carattere '…' che segue gli autori corrisponde a "\u2026" in utf-8!!!!
			leftSide = strings.Replace(leftSide, "\u2026", "", -1)
			documents[docIndex].Authors = strings.Split(leftSide, ", ")

			//scorro other, mi fermo quando trovo un match con: "Citato da", cosi' so che sono sull'elemento giusto
			for ; ; otherIndex++ {
				text, err = other[otherIndex].Text()
				fmt.Println(text)
				if err != nil {
					url, _ := wd.CurrentURL()
					fmt.Println("Url: ", url)
					panic(err)
				}
				if t, _ := regexp.MatchString("Cited by.*", text); t {
					words := strings.Split(text, " ")
					if numCitedBy, err := strconv.ParseUint(words[2], 10, 16); err != nil {
						panic(err)
					} else {
						documents[docIndex].NumCitedBy = uint16(numCitedBy)
					}
					linkCitedBy, _ := other[otherIndex].GetAttribute("href")
					documents[docIndex].LinkCitedBy = structures.URLScholar + linkCitedBy
					break
				}
			}
		}
		return documents, uint16(docIndex)
	*/
}

//Restituisce il documento da cui inizia la ricerca
func GetInitialDocument_MA(wd selenium.WebDriver) *structures.Document {
	if err := wd.Get(structures.URLAcademic); err != nil {
		panic(err)
	}
	//stampo url
	url, _ := wd.CurrentURL()
	fmt.Println("Url: ", url)

	textBox, err := wd.FindElement(selenium.ByXPATH,
		"//ma-queryformulation[@class='searchWrap']/div/"+
			"div[@class='search-input']/input[@class='searchControl']")
	if err != nil {
		panic(err)
	}
	if err := textBox.SendKeys(`GDPR`); err != nil {
		panic(err)
	}
	searchButton, err := wd.FindElement(selenium.ByXPATH,
		"//ma-queryformulation[@class='searchWrap']/div/div[@class='search-btn']")
	if err != nil {
		panic(err)
	}

	if err := searchButton.Click(); err != nil {
		panic(err)
	}
	//stampa----------------------------------------
	url, _ = wd.CurrentURL()
	fmt.Println("Url: ", url)

	//prendo il primo documento della pagina
	docs, numDocs := GetDocumentsFromPage_MA(wd, 1)
	if numDocs > 1 {
		panic(fmt.Sprintf("GetInitialDocument - GetDocumentsFromPage\nHa resituito piu' di un documento"))
	}
	fmt.Println("\n\n", numDocs, docs)
	return nil
	//return docs[0]
}

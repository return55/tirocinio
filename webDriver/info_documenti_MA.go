package webDriver

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"math/rand"
	"math"
	"time"

	"github.com/return55/tirocinio/structures"

	"github.com/tebeka/selenium"
)

//Informazioni che Raccolgo dai Documenti:
// *Title           string
// *Url            sources  //URL dei vari sorgenti disponibili
// *Authors        []author //Nomi, cognomi e affiliazioni dei vari autori
// *NumCitations   uint16
// *LinkCitations  string //Link alla pagina di Academy con i documenti che lo citano
// *NumReferences  uint16
// *LinkReferences string //Link alla pagina di Academy con i documenti che cita
// Abstract       string
// *Date           per ora string//time.Time //Data pubblicazione
// *FieldsOfStudy  []string

var (
	//creo il logger per la ricerca in Microsoft Academic
	fileMA, _ = os.OpenFile("robaMA.LOG", os.O_WRONLY, 0600)
	logger    = log.New(fileMA, "", 0)
)

//Condizione per il caricamento della pagina con i risultati (aspetto 1 article)
func conditionResultPage(wd selenium.WebDriver) (bool, error) {
	elem, err := wd.FindElements(selenium.ByXPATH, "//article[@class='paper paper-mode-2 card']")

	if err != nil {
		panic(err)
	}
	if len(elem) == 0 {
		return false, err
	}
	return true, err
}

//Condizione per il caricamento della pagina del singolo documento (aspetto i fields of study)
func conditionDocumentPage(wd selenium.WebDriver) (bool, error) {
	elem, err := wd.FindElements(selenium.ByXPATH, 
		"//section[@class='pure-u-1 pure-u-md-1-4 entity-right detail-right']"+
		"/ma-ulist/div/div[@class='ulist-body']/ul[@class='ulist-content']")

	if err != nil {
		panic(err)
	}
	if len(elem) == 0 {
		fmt.Println("Lunghezza 0")
		return false, err
	}
	fmt.Println("************* ",elem)
	return true, err
}

//Data un pagina (impostata dal WebDriver) prendo un certo numero di documenti dalla pagina
//partendo dal primo in alto.
//Se il numero (numDocs) e' maggiore del numero di documenti nella pagina (tipicamente 8),
//mi limito a restituire i documenti presenti nella pagina e la loro quantita'.
func GetDocumentsFromPage_MA(wd selenium.WebDriver, numDocs int) ([]structures.MADocument, uint64) {
	//aspetto che gli elementi article siano caricati
	wd.Wait(conditionResultPage)

	//scorro i link ai documenti presenti nella pagina
	/*links, err := wd.FindElements(selenium.ByXPATH, "//article")
	if err != nil {
		panic(err)
	}
	 fmt.Println("Lunghezza: ", len(links))
	for _, link := range links {
	 	url, err := link.GetAttribute("class")
	 	if err != nil {
	 		fmt.Println("---------------------------------")
	 	}
	 	fmt.Println(url)
	 }
	 return nil, 0*/

	//prendo i titoli dei documenti  (titles.GetAttribute("title"))
	titles, err := wd.FindElements(selenium.ByXPATH,
		"//article/section[@class='paper-title']/h2/a[@class='blue-title']")
	if err != nil {
		panic(err)
	}

	//creo array di documenti pari al minimo(numDocs, numResults)
	var min int
	if numDocs <= len(titles) {
		min = numDocs
	} else {
		min = len(titles)
	}
	documents := make([]structures.MADocument, min)
	
	//-----------------------------------------------------------------------
	authorsAndAffiliations, err := wd.FindElements(selenium.ByXPATH,
			"//section[@class='paper-authors']/ma-ulist/div/div[@class='ulist-body']/ul")

		if err != nil {
			if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
				currentUrl, err := wd.CurrentURL()
				if err != nil {
					panic(err)
				}
				logger.Printf("Alla pagina %s non ci sono info sugli Autori", currentUrl)
				//gli autori rimangono array vuoti, non devo fare niente				
				//ci andra' un return
			} else {
				panic(err)
			}
		}
		fmt.Println("Numero doc: ", len(authorsAndAffiliations)); 
		//se il numero degli autori e' diverso da quello dei titoli, c'e' un problema
		if len(titles) != len(authorsAndAffiliations) {
			currentUrl, err := wd.CurrentURL()
			if err != nil {
				panic(err)
			}
			logger.Printf("Numero titoli(%i) != numero gruppi autori(%i) per %s\n", len(titles), len(authorsAndAffiliations), currentUrl) 
		}
		//scorro autori e affiliazioni
		for pos:=0; pos<min; pos++ {
			//prendo gli autori
			authors, err := authorsAndAffiliations[pos].FindElements(selenium.ByXPATH,
				"li/span/a")
					fmt.Println("Numero autori: ", len(authors))
			if err != nil {
				panic(err)
			}
		
			textAuthors := make([]string, len(authors))
			for i := 0; i < len(authors); i++ {
				textAuthors[i], err = authors[i].Text()
				fmt.Println("Autrice: ", textAuthors[i])
				if err != nil {
					panic(err)
				}
			}
			//prendo le affiliazioni 
			affiliation, err := authorsAndAffiliations[pos].FindElements(selenium.ByXPATH,
				"li/span/span[@class='affiliation']/ul/li/a[@class='button-link']")
								fmt.Println("Numero affiliazioni: ", len(authors))
			textAffiliation := make([]string, len(authors))
			_ = textAffiliation[0]
			if err != nil {
				if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
					//Non ci sono affiliazioni
					currentUrl, err := wd.CurrentURL()
					if err != nil {
						panic(err)
					}
					logger.Printf("Alla pagina %s non ci sono Affiliazioni doc  numero %i", currentUrl, pos)
					//Imposto le affiliazioni a valori nulli ("")
					/*for i:=0; i<len(textAffiliation); i++ {
						textAffiliation[i]=""
					} */
				} else {
					panic(err)
				}
			} else {
				for i := 0; i < len(affiliation); i++ {
					textAffiliation[i], err = affiliation[i].Text()
					fmt.Println("Affiliazione: ", textAffiliation[i])
					if err != nil {
						panic(err)
					}
				}				
			}
			//Aggiungo autori e affiliazioni al documento
			for i := 0; i < len(authors); i++ {
				documents[pos].Authors = append(documents[pos].Authors,
					structures.Author{textAuthors[i], textAffiliation[i]})
			}
		
		}
		
		//-------------------------------------------------------------------------------------------------------*/
	//controllo
	/*fmt.Println("Numero documenti: ", len(titles))
	for _, t := range titles {
		tit, _ := t.GetAttribute("title")
		tit = strings.Replace(tit, "%!(EXTRA string=", "", 1)
		tit = strings.TrimSuffix(tit, ")")
		fmt.Println("Titolo: ", tit)
	}
	return nil, 0*/

	//scorro i documenti della pagina
	for count := 0; count < min; count++ {
		//assegno il titolo
		tit, _ := titles[count].GetAttribute("title")
		//Devo rimuovere il prefisso: %!(EXTRA string= e il suffisso: )
		//Non ho idea da dove arrivino
		tit = strings.Replace(tit, "%!(EXTRA string=", "", 1)
		tit = strings.TrimSuffix(tit, ")")
		documents[count].Title = tit
		//per prendere tutte le informazioni devo andare alla pagina del documento:
		titles[count].Click()
		//aspetto di caricare la pagina (i fields of study come riferimento)
		wd.Wait(conditionDocumentPage)
		//Espando tutti gli "show more": fields of study, sources
		showMore, err := wd.FindElements(selenium.ByXPATH,
			"//section[@class='pure-u-1 pure-u-md-1-4 entity-right detail-right']//div[@class='ulist-show-more']/a")
		if err != nil {
			if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
				currentUrl, err := wd.CurrentURL()
				if err != nil {
					panic(err)
				}
				logger.Printf("Alla pagina %s non ci sono Show More", currentUrl)
			} else {
				panic(err)
			}
		}
		fmt.Println("Numero show more: ", len(showMore))
		for _, showToClick := range showMore {
			err = showToClick.Click()
			if err != nil {
				if t, _ := regexp.MatchString(".*element not interactable.*", err.Error()); t {
					currentUrl, err := wd.CurrentURL()
					if err != nil {
						panic(err)
					}
					logger.Printf("Gli Show More non sono interagibili", currentUrl)
				} else {
					panic(err)
				}
			}
		}
		/*currentUrl, err := wd.CurrentURL()
		fmt.Println("URL: ", currentUrl)*/
		//prendo i fields of study e sources
		fieldsAndSources, err := wd.FindElements(selenium.ByXPATH,
			"//section[@class='pure-u-1 pure-u-md-1-4 entity-right detail-right']"+
				"/ma-ulist/div/div[@class='ulist-body']/ul[@class='ulist-content']")
		if err != nil {
			panic(err)
		}
		fmt.Println(fieldsAndSources)
		//fields of study
		fieldsOfStudy, err := fieldsAndSources[0].FindElements(selenium.ByXPATH,
			"li/a/span")
		if err != nil {
			if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
				currentUrl, err := wd.CurrentURL()
				if err != nil {
					panic(err)
				}
				logger.Printf("Alla pagina %s non ci sono Fields Of Study", currentUrl)
			} else {
				panic(err)
			}
		}
		for _, field := range fieldsOfStudy {
			textField, _ := field.Text()
			documents[count].FieldsOfStudy = append(documents[count].FieldsOfStudy, textField)
		}
		//sources
		sources, err := fieldsAndSources[1].FindElements(selenium.ByXPATH,
			"li/a")
		if err != nil {
			if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
				currentUrl, err := wd.CurrentURL()
				if err != nil {
					panic(err)
				}
				logger.Printf("Alla pagina %s non ci sono Sources", currentUrl)
			} else {
				panic(err)
			}
		}
		for _, source := range sources {
			URLSource, _ := source.GetAttribute("href")
			//controllo se e' un PDF
			if t, _ := regexp.MatchString(".*\\.pdf", URLSource); t {
				documents[count].Url.PDF = append(documents[count].Url.PDF, URLSource)
			} else {
				documents[count].Url.WWW = append(documents[count].Url.WWW, URLSource)
			}
		}		
		//Prendo la data(posizione 0)
		date, err := wd.FindElement(selenium.ByXPATH,
			"//section[@class='paper-year']/span")
		if err != nil {
			panic(err)
		}
		documents[count].Date, _ = date.Text()
		fmt.Println("Data: ", documents[count].Date)
		//Prendo le citations (0), references (1) (opz. related (2))
		referencesAndCitations, err := wd.FindElements(selenium.ByXPATH,
			"//div[@class='pure-u-md-4-24 pure-u-1 digit']")
		if err != nil {
			panic(err)
		}
		fmt.Println(len(referencesAndCitations))
		//References
		numRef, err := referencesAndCitations[0].FindElement(selenium.ByXPATH,
			"h1")
		if err != nil {
			panic(err)
		}
		textNumRef, _ := numRef.Text()
		//elimino la virgola (se presente)
		textNumRef = strings.Replace(textNumRef, ",", "", -1)
		documents[count].NumReferences, err = strconv.ParseInt(textNumRef, 10, 0)
		if err != nil {
			//Non ci sono references
			documents[count].NumReferences = 0
			documents[count].LinkReferences = ""
		} else {
			URLRef, err := referencesAndCitations[0].FindElement(selenium.ByXPATH,
				"a")
			if err != nil {
				panic(err)
			}
			textURLRef, _ := URLRef.GetAttribute("href")
			documents[count].LinkReferences = structures.URLAcademic + textURLRef
		}
		//Citations
		numCit, err := referencesAndCitations[1].FindElement(selenium.ByXPATH,
			"h1")
		if err != nil {
			panic(err)
		}
		textNumCit, _ := numCit.Text()
		
		//elimino la virgola (se presente)
		textNumCit = strings.Replace(textNumCit, ",", "", -1)
		fmt.Println("Numero citazioni: ", textNumCit)
		documents[count].NumCitations, err = strconv.ParseInt(textNumCit, 10, 0)
		fmt.Println("numero citazioni: ", documents[count].NumCitations)
		if err != nil {
			fmt.Println("Entro nell'errore delle citazioni")
			//Non ci sono citations
			documents[count].NumCitations = 0
			documents[count].LinkCitations = ""
		} else {
			URLCit, err := referencesAndCitations[1].FindElement(selenium.ByXPATH,
				"a")
			if err != nil {
				panic(err)
			}
			textURLCit, _ := URLCit.GetAttribute("href")
			documents[count].LinkCitations = structures.URLAcademic + textURLCit
		}
		//Abstract (0)
		abstractSec, err := wd.FindElement(selenium.ByXPATH,
			"//section[@class='paper-abstract']/p")
		if err != nil {
			if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
				currentUrl, err := wd.CurrentURL()
				if err != nil {
					panic(err)
				}
				logger.Printf("Alla pagina %s non ci sono Abstract", currentUrl)
				documents[count].Abstract = ""
			} else {
				panic(err)
			}
		} else {
			documents[count].Abstract, _ = abstractSec.Text()
		}
	
		logger.Println("---------------------------------------------------")
		//Torno alla pagina dei risultati(FORSE NON E' NECESSARIO)
		wd.Back()
	}

	return documents, uint64(min)

}

//Restituisce il documento da cui inizia la ricerca
func GetInitialDocument_MA(wd selenium.WebDriver) structures.MADocument {
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
	if err := textBox.SendKeys(`bio`); err != nil {
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

	return docs[0]
}


//DA METTERE A POSTO-------------------------------------------------------
//Dato un link alla pagina di partenza, comincio a raccogliere i documenti (8 per pagina)
//finche' non arrivo a numDoc.
func GetCiteDocuments_MA(wd selenium.WebDriver, linkCitedBy string, numDoc uint64) ([]structures.MADocument, uint64) {
	if err := wd.Get(linkCitedBy); err != nil {
		panic(err)
	}
	var allDoc []structures.MADocument
	//Mi serve per dire quanti documenti ho preso
	initialNumDoc := numDoc
	fmt.Println("***** numDoc= " + strconv.FormatUint(numDoc, 10))

	//genero la sequenza di numeri casuali
	r:=rand.New(rand.NewSource(12))
	
	for numDoc > 0 {

		newDoc, numNewDoc := GetDocumentsFromPage_MA(wd, int(numDoc))
		allDoc = append(allDoc, newDoc...)
		//tolgo il numero di documenti appena letti
		numDoc = numDoc - numNewDoc
		fmt.Println("***** docRead= ", numNewDoc)
		
		/* Scorro una pagina alla volta in sequenza 
		//vado alla prosssima pagina, se possibile:
		linkAvanti, err := wd.FindElement(selenium.ByXPATH, "//b[text()='Next']/..")
		//se non trovo il link per andare avanti, mi fermo
		if err != nil {
			if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
				return allDoc, docRead
			} else {
				panic(err)
			}
		}

		url, err := linkAvanti.GetAttribute("href")
		if err != nil {
			panic(err)
		}
		///////////////////////////////////*/

		/* Scorro in sequenza ma aspetto un tempo che cresce in modo esponenziale */
		waitTimeSec := time.Duration((math.Round(r.ExpFloat64())))
		time.Sleep(waitTimeSec * time.Second)
		
		//vado alla prosssima pagina, se possibile:
		linkAvanti, err := wd.FindElement(selenium.ByXPATH, "//div[@class='entityResultPager']/ul/li/a[@aria-label='Next']")
		//se non trovo il link per andare avanti, mi fermo
		if err != nil {
			if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
				return allDoc, initialNumDoc - numDoc
			} else {
				panic(err)
			}
		}

		err = linkAvanti.Click()
		if err != nil {
			panic(err)
		}
		//////////////////////////////////////////////
		/*if err := wd.Get(structures.URLScholar + url); err != nil {
			panic(err)
		}*/
	}
	return allDoc, initialNumDoc - numDoc
}

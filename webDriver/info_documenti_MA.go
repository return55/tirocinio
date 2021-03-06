package webDriver

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
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
// *Abstract       string
// Date           per ora string//time.Time //Data pubblicazione
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
	logger.Println("Condizione pagina dei risultati")
	if len(elem) < structures.NumArticlePerPageMA {
		return false, err
	}

	return true, err
}

//Sempre per la pagina dei risultati
func conditionTitleText(wd selenium.WebDriver) (bool, error) {
	titles, err := wd.FindElements(selenium.ByXPATH,
		"//article/section[@class='paper-title']/h2/a[@class='blue-title']")
	if err != nil {
		panic(err)
	}
	for _, t := range titles {
		tit, err := t.Text()
		//tit, _ := t.GetAttribute("title")
		if tit == "" || err != nil {
			return false, err
		}
	}
	return true, err
}

//Condizione per il caricamento della pagina del singolo documento (aspetto i aspetto i fields of study e sources)
func conditionDocumentPage(wd selenium.WebDriver) (bool, error) {
	elem, err := wd.FindElements(selenium.ByXPATH,
		"//section[@class='pure-u-1 pure-u-md-1-4 entity-right detail-right']"+
			"/ma-ulist/div/div[@class='ulist-body']/ul[@class='ulist-content']")

	if err != nil {
		panic(err)
	}
	if len(elem) < 2 {
		logger.Println("Condizione caricamento pagina documento: fields e sources")
		return false, err
	}
	logger.Println("************* ", elem)
	currentUrl, err := wd.CurrentURL()
	if err != nil {
		panic(err)
	}
	logger.Println(currentUrl)
	return true, err
}

//Condizione per il caricamento della pagina del singolo documento (aspetto gli showmore)
func conditionDocumentPage2(wd selenium.WebDriver) (bool, error) {
	elem, err := wd.FindElements(selenium.ByXPATH,
		"//section[@class='pure-u-1 pure-u-md-1-4 entity-right detail-right']//div[@class='ulist-show-more']/a")

	if err != nil {
		panic(err)
	}
	if len(elem) == 0 {
		logger.Println("Condizione caricamento pagina documento: show more")
		return false, err
	}
	logger.Println("************* ", elem)
	currentUrl, err := wd.CurrentURL()
	if err != nil {
		panic(err)
	}
	logger.Println(currentUrl)
	return true, err
}

//Mette a posto i titoli e si salva i link ai vari documenti
//Usa la dimensione di URLDocuments per scorrere titles e document
func setTitlesAndGetURLs(titles []selenium.WebElement, documents []structures.MADocument, URLDocuments []string) {
	for i := 0; i < len(URLDocuments); i++ {
		tit, _ := titles[i].GetAttribute("title")
		tit = strings.Replace(tit, "%!(EXTRA string=", "", 1)
		tit = strings.TrimSuffix(tit, ")")
		documents[i].Title = tit

		logger.Println("Titolo ", i, ": ", documents[i].Title)

		URLDoc, _ := titles[i].GetAttribute("href")
		URLDocuments[i] = structures.URLAcademic + URLDoc
		documents[i].URL = structures.URLAcademic + URLDoc

		logger.Println("URL ", i, ": ", URLDocuments[i])
	}
}

//Prendendo gli autori dalla pagina principale, ne lascio indietro alcuni perche'
//non sono subito visibili.
//Stesso discorso per le affiliazioni.
func getAuthorsInResultPage(wd selenium.WebDriver, documents []structures.MADocument) {
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
			return
		} else {
			panic(err)
		}
	}
	logger.Println("Numero tot autori: ", len(authorsAndAffiliations))
	//scorro autori e affiliazioni
	for pos := 0; pos < len(documents); pos++ {
		//prendo gli autori
		authors, err := authorsAndAffiliations[pos].FindElements(selenium.ByXPATH,
			"li/span/a")
		logger.Println("Numero autori: ", len(authors))
		if err != nil {
			panic(err)
		}
		textAuthors := make([]string, len(authors))
		for i := 0; i < len(authors); i++ {
			textAuthors[i], err = authors[i].Text()
			logger.Println("Autore: ", textAuthors[i])
			if err != nil {
				panic(err)
			}
		}
		//prendo le affiliazioni
		//NOTA:
		//un autore puo' avere piu' affiliazioni per lo stesso articolo, io
		//prendo solo la prima.
		affiliation, err := authorsAndAffiliations[pos].FindElements(selenium.ByXPATH,
			"li/span/span[@class='affiliation']/ul")
		logger.Println("Numero affiliazioni: ", len(authors))
		textAffiliation := make([]string, len(authors))
		if err != nil {
			if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
				//Non ci sono affiliazioni
				currentUrl, err := wd.CurrentURL()
				if err != nil {
					panic(err)
				}
				logger.Printf("Alla pagina %s non ci sono Affiliazioni doc  numero %i", currentUrl, pos)
				//Non e' necessario
				//Imposto le affiliazioni a valori nulli ("")
				/*for i:=0; i<len(textAffiliation); i++ {
					textAffiliation[i]=""
				} */
			} else {
				panic(err)
			}
		} else {
			//Prendo il primo li di ogni ul
			for i := 0; i < len(affiliation); i++ {
				firstAff, err := affiliation[i].FindElement(selenium.ByXPATH,
					"li/a[@class='button-link']")
				if err == nil {
					textAffiliation[i], err = firstAff.GetAttribute("title")
				} else {
					firstAff, err = affiliation[i].FindElement(selenium.ByXPATH,
						"li/span")
					if err == nil {
						textAffiliation[i], err = firstAff.GetAttribute("title")
					} else {
						textAffiliation[i] = ""
					}
				}
				logger.Println("Affiliazione: ", textAffiliation[i])

			}
		}
		//Aggiungo autori e affiliazioni al documento
		for i := 0; i < len(authors); i++ {
			documents[pos].Authors = append(documents[pos].Authors,
				structures.Author{textAuthors[i], textAffiliation[i]})
		}
	}
}

//Esapndo gli show more di fileds of study e sources, altrimenti ne perderei alcuni
func expandShowMore(wd selenium.WebDriver) {
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
	logger.Println("Numero show more: ", len(showMore))
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
}

//Imposto i fields of study e sources(www e pdf) per un singolo doc
func setFieldsOfStudyAndSources(wd selenium.WebDriver, document *structures.MADocument) {
	fields, err := wd.FindElements(selenium.ByXPATH,
		"//div[@class='tag-cloud']")
	if err != nil {
		panic(err)
	}
	logger.Println(fields)
	//fields of study
	fieldsOfStudy, err := fields[0].FindElements(selenium.ByXPATH,
		"/ma-link-tag/a/div[@class='text']")
	if err != nil {
		if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
			currentUrl, err := wd.CurrentURL()
			if err != nil {
				panic(err)
			}
			logger.Printf("Alla pagina %s non ci sono Fields Of Study", currentUrl)
			//Se non ci sono campi, lo scrivo per neo4j
			document.FieldsOfStudy = append(document.FieldsOfStudy, "No Fields")
		} else {
			panic(err)
		}
	} else {
		for _, field := range fieldsOfStudy {
			textField, _ := field.Text()
			document.FieldsOfStudy = append(document.FieldsOfStudy, textField)
		}
	}

	//all sources: 0 PDF, 1 WWW
	sources, err := wd.FindElements(selenium.ByXPATH,
		"//ma-link-collection[@class='source-links au-target']/div[@class='ma-link-collection']")
	if err != nil {
		panic(err)
	}
	logger.Println(sources)
	//sources PDF
	PDFsources, err := sources[0].FindElements(selenium.ByXPATH, "/a")
	if err != nil {
		panic(err)
	}
	if len(PDFsources) >= 0 {
		for _, source := range PDFsources {
			URLSource, _ := source.GetAttribute("href")
			document.Url.PDF = append(document.Url.PDF, URLSource)
		}
	}
	//sources WEB
	WWWsources, err := sources[1].FindElements(selenium.ByXPATH, "/a")
	if err != nil {
		panic(err)
	}
	if len(WWWsources) >= 0 {
		for _, source := range WWWsources {
			URLSource, _ := source.GetAttribute("href")
			document.Url.WWW = append(document.Url.WWW, URLSource)
		}
	}
	/*//non posso gestire errore "index out of range", devo controllare la dimensione di fieldsAndSources
	if len(fieldsAndSources) > 1 {
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
		} else {
			for _, source := range sources {
				URLSource, _ := source.GetAttribute("href")
				//controllo se e' un PDF
				if t, _ := regexp.MatchString(".*\\.pdf", URLSource); t {
					document.Url.PDF = append(document.Url.PDF, URLSource)
				} else {
					document.Url.WWW = append(document.Url.WWW, URLSource)
				}
			}
		}
	}*/
	//Se non ho trovato PDF o WWW, me lo segno per neo4j
	if len(document.Url.PDF) == 0 {
		document.Url.PDF = append(document.Url.PDF, "No Source PDF")
	}
	if len(document.Url.WWW) == 0 {
		document.Url.WWW = append(document.Url.WWW, "No Source WWW")
	}
}

//Imposto la data per un singolo doc
func setDate(wd selenium.WebDriver, document *structures.MADocument) {
	date, err := wd.FindElement(selenium.ByXPATH,
		"//section[@class='paper-year']/span")
	if err != nil {
		panic(err)
	}
	document.Date, _ = date.Text()
	//Se non ho trovato la Data, lo dico a neo4j
	/*if document.Date == "" {
		document.Date = "No Date"
	}*/
	logger.Println("Data: ", document.Date)
}

//Imposto le citazioni(numero e link) e le refernces(numero e link)
func setCitationsAndReferences(wd selenium.WebDriver, document *structures.MADocument) []string {
	/*referencesAndCitations, err := wd.FindElements(selenium.ByXPATH,
		"//div[@class='pure-u-md-4-24 pure-u-1 digit']")
	if err != nil {
		panic(err)
	}
	logger.Println(len(referencesAndCitations))
	//References
	numRef, err := referencesAndCitations[0].FindElement(selenium.ByXPATH,
		"h1")
	if err != nil {
		panic(err)
	}
	textNumRef, _ := numRef.Text()
	//elimino la virgola (se presente)
	textNumRef = strings.Replace(textNumRef, ",", "", -1)
	document.NumReferences, err = strconv.ParseInt(textNumRef, 10, 0)
	if err != nil {
		//Non ci sono references
		document.NumReferences = 0
		document.LinkReferences = ""
	} else {
		URLRef, err := referencesAndCitations[0].FindElement(selenium.ByXPATH,
			"a")
		if err != nil {
			panic(err)
		}
		textURLRef, _ := URLRef.GetAttribute("href")
		document.LinkReferences = structures.URLAcademic + textURLRef
	}*/
	//Citations
	numCit, err := wd.FindElement(selenium.ByXPATH,
		"div[@class='stats']/ma-statistics-item/a[@title='Citations*']/div[@class='data']/div[@class='count']")
	if err != nil {
		panic(err)
	}
	textNumCit, _ := numCit.Text()
	fmt.Println("Num citazioni= ", textNumCit)

	logger.Println("Numero citazioni: ", textNumCit)
	document.NumCitations, err = strconv.ParseInt(textNumCit, 10, 0)
	logger.Println("numero citazioni: ", document.NumCitations, " ... ", textNumCit)

	//prendo i link ai documenti che citano
	var urlDocCitations []string
	//premo pulsante "Cyted by"
	buttonCytedBy, err := wd.FindElement(selenium.ByXPATH,
		"ma-call-to-action[@data-appinsights-route='Cited By']")
	if err != nil {
		panic(err)
	}
	err = buttonCytedBy.Click()
	if err != nil {
		panic(err)
	}
	//prendo i url dei doc
	docs, err := wd.FindElements(selenium.ByXPATH,
		"div[@class='results']/ma-card/div/compose/div[@class='paper']/a[@class='title au-target']/")
	if err != nil {
		panic(err)
	}
	fmt.Println("numero cyted by", len(docs))
	for _, doc := range docs {
		href, err := doc.GetAttribute("href")
		if err != nil {
			panic(err)
		}
		urlDocCitations = append(urlDocCitations, structures.URLAcademic+href)
	}

	return urlDocCitations
}

//Imposto l'abstract del doc
func setAbstract(wd selenium.WebDriver, document *structures.MADocument) {
	abstractSec, err := wd.FindElement(selenium.ByXPATH,
		"//section[@class='paper-abstract']/p")
	if err != nil {
		if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
			currentUrl, err := wd.CurrentURL()
			if err != nil {
				panic(err)
			}
			logger.Printf("Alla pagina %s non ci sono Abstract", currentUrl)
			document.Abstract = ""
		} else {
			panic(err)
		}
	} else {
		document.Abstract, _ = abstractSec.Text()
	}
}

//Data un pagina (impostata dal WebDriver) prendo un certo numero di documenti dalla pagina
//partendo dal primo in alto.
//Se il numero (numDocs) e' maggiore del numero di documenti nella pagina (tipicamente 8),
//mi limito a restituire i documenti presenti nella pagina e la loro quantita'.
func GetDocumentsFromPage_MA(wd selenium.WebDriver, numDocs int) ([]structures.MADocument, uint64) {

	currentUrl, err := wd.CurrentURL()
	if err != nil {
		panic(err)
	}
	logger.Println("Url attuale: ", currentUrl)

	//per non far arrabbiare MA
	//time.Sleep(4000 * time.Millisecond)

	_ = wd.Refresh()
	//aspetto che gli elementi article siano caricati
	wd.WaitWithTimeout(conditionResultPage, 3000*time.Millisecond)

	sorgente, err := wd.PageSource()
	if err != nil {
		panic(err)
	}
	logger.Println(sorgente)

	//controllo quanti documenti hanno num citazioni >= soglia

	//prendo i titoli dei documenti  (titles.GetAttribute("title"))
	titles, err := wd.FindElements(selenium.ByXPATH,
		"//article/section[@class='paper-title']/h2/a[@class='blue-title']")
	if err != nil {
		panic(err)
	}
	/*
		fileS, _ := os.OpenFile("sorgenteDopoTitoli.html", os.O_WRONLY, 0600)
		logger = log.New(fileS, "", 0)
		sorgente, err = wd.PageSource()
		if err != nil {
			panic(err)
		}
		logger.Println(sorgente)

		currentUrl, err = wd.CurrentURL()
		if err != nil {
			panic(err)
		}
		logger.Println("Url attuale dopo titoli: ", currentUrl)*/

	//creo array di documenti pari al minimo(numDocs, numResults)
	var min int
	if numDocs <= len(titles) {
		min = numDocs
	} else {
		min = len(titles)
	}
	documents := make([]structures.MADocument, min)
	URLDocuments := make([]string, min)

	//Assegno i titoli ai doc e mi savo i relativi link in una var a parte perche'
	//una volta cambiata la pagina perdo il riferimento all'elemento con il link
	//assegno il titolo
	setTitlesAndGetURLs(titles, documents, URLDocuments)

	//Prendendo gli autori dalla pagina principale, ne lascio indietro alcuni perche'
	//non sono subito visibili.
	//Stesso discorso per le affiliazioni.
	getAuthorsInResultPage(wd, documents)

	//scorro i documenti della pagina
	for count := 0; count < min; count++ {
		//per non far arrabbiare MA
		//time.Sleep(2000 * time.Millisecond)

		//per prendere tutte le informazioni devo andare alla pagina del documento:
		if err := wd.Get(URLDocuments[count]); err != nil {
			panic(err)
		}

		//per non far arrabbiare MA
		//time.Sleep(1050 * time.Millisecond)

		//aspetto di caricare la pagina (i fields of study come riferimento)
		wd.WaitWithTimeout(conditionDocumentPage, 60*time.Second)

		currentUrl, err := wd.CurrentURL()
		if err != nil {
			panic(err)
		}
		logger.Println("URL: ", currentUrl)

		//Espando gli "show more" di fields of study e sources
		expandShowMore(wd)

		//prendo i fields of study e sources
		setFieldsOfStudyAndSources(wd, &documents[count])

		//Prendo la data(posizione 0)    NON FUNZIONA
		setDate(wd, &documents[count])

		//Prendo le citations (0), references (1) (opz. related (2))
		setCitationsAndReferences(wd, &documents[count])

		//Abstract (0)
		setAbstract(wd, &documents[count])

		logger.Println("---------------------------------------------------")
		//Torno alla pagina dei risultati(E' NECESSARIO -- NON NE SONO SICURO (FAI UNA PROVA))
		//wd.Back()
	}

	return documents, uint64(min)

}

//Uso sempre una soglia come criterio per raccogliere le informazioni ma mi
//limito a raccogliere: titolo, LinkCitations, numCitations, fields of study.
//Devo anche raccogliere i fields of study (keyword) -> devo visitare la pagina di ogni articolo
//Puo' avere 2 comportamenti, threshold puo' essere:
//- il numero minimo di citazioni di un documento
//- il massimo numero di citazioni tra i documenti che citano (il primo della prima pagina dei risultati),
//	prendo quelli che hanno almeno un numero di citazioni pari a una percentuale di threshold.
//	Se threshold = -1 -> sono al primo giro
//
//Ritorno: documenti, quanti ne ho presi e la nuova soglia (che cambia solo al primo giro)
func GetDocumentsFromPageBasic_MA(wd selenium.WebDriver, maxCit, threshold, perc int) ([]structures.MADocument, int, int) {
	currentUrl, err := wd.CurrentURL()
	if err != nil {
		panic(err)
	}
	logger.Println("Url attuale **** : ", currentUrl)

	//per non far arrabbiare MA
	//time.Sleep(3000 * time.Millisecond)

	_ = wd.Refresh()
	//aspetto che gli elementi article siano caricati
	//wd.WaitWithTimeout(conditionResultPage, 60*time.Second)
	wd.WaitWithTimeout(conditionTitleText, 60*time.Second)

	sorgente, err := wd.PageSource() // |  va insieme all'else
	//_, err = wd.PageSource()
	if err != nil {
		logger.Println(err)
	} else {
		logger.Println(sorgente)
	}

	//prendo i titoli degli articoli, mi serviranno per scorrere
	//le pagine
	titles, err := wd.FindElements(selenium.ByXPATH,
		"//article/section[@class='paper-title']/h2/a[@class='blue-title']")
	if err != nil {
		panic(err)
	}
	//piccolo controllo
	for i, t := range titles {
		tit, _ := t.Text()
		//tit, _ := t.GetAttribute("title")
		logger.Println("Titolo ", i, " :", tit)
	}
	//prendo tutti gli articoli
	articles, err := wd.FindElements(selenium.ByXPATH,
		"//article[@class='paper paper-mode-2 card']")
	if err != nil {
		panic(err)
	}
	//prendo il numero di citazioni di ogni articolo e conto
	//quanti di questi superano la soglia (howMany).
	//Quando scendo sotto la soglia: ridimensiono titles e creo
	//la variabile che conterra i documenti e quella con gli URL.
	howMany := len(articles)
	numsCitations := make([]int64, len(articles))
	for i, article := range articles {
		//il numero di citazioni sta nel primo elemento della lista
		numCitations, err := article.FindElement(selenium.ByXPATH,
			"section[@class='paper-actions']/ul/li/a[@class='c-count']/span")
		if err != nil {
			if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
				//controllo se e' presente la scritta "Not cited"
				notCited, err := article.FindElement(selenium.ByXPATH,
					"section[@class='paper-actions']/ul/li/span")
				if err != nil {
					panic(err)
				}
				textNotCited, _ := notCited.Text()
				if textNotCited == "Not cited" {
					//non ha citazioni, quindi non soddisfa la soglia
					break
				}
			} else {
				panic(err)
			}
		}
		//se arrivo qui, ho un numero di citazioni e lo devo controllare
		textNumCitations, _ := numCitations.Text()
		//estraggo i numeri: formato del .Text() -> "Citations (n)"
		textNumCitations = strings.Fields(textNumCitations)[1]
		textNumCitations = strings.Replace(textNumCitations, "(", "", -1)
		textNumCitations = strings.Replace(textNumCitations, ")", "", -1)
		//elimino la virgola (se presente)
		textNumCitations = strings.Replace(textNumCitations, ",", "", -1)
		numsCitations[i], err = strconv.ParseInt(textNumCitations, 10, 0)

		logger.Println("Num citazioni: ", numsCitations[i])

		//se e' il primo risulatato della pagina e non ho ancora un valore valido di threshold
		if i == 0 && maxCit == -1 {
			maxCit = int(numsCitations[i])
			logger.Println("Soglia: ", maxCit)
		}
		/*Condozione banale sulla soglia
		if int(numsCitations[i]) < threshold {*/
		/*Condizione superiore a una percentuale di maxCit*/
		if numsCitations[i] < int64(threshold) || float32(numsCitations[i]) < float32(maxCit)*(float32(perc)/100) {
			howMany = i
			break
		}
	}
	//ridimensiono e creo array per documenti e url
	titles = titles[:howMany]
	numsCitations = numsCitations[:howMany]
	docs := make([]structures.MADocument, howMany)
	URLDocuments := make([]string, howMany)

	//Assegno i titoli e gli URL ai doc e mi savo i relativi link in una var a parte perche'
	//una volta cambiata la pagina perdo il riferimento all'elemento con il link
	//assegno il titolo
	setTitlesAndGetURLs(titles, docs, URLDocuments)

	//raccolgo i fields of study e creo i documenti
	for i := 0; i < howMany; i++ {
		docs[i].NumCitations = numsCitations[i]
		//link alle citazioni
		linkCitations, err := articles[i].FindElement(selenium.ByXPATH,
			"section[@class='paper-actions']/ul/li/a[@class='c-count']")
		if err != nil {
			panic(err)
		}
		textLinkCitations, err := linkCitations.GetAttribute("href")
		if err != nil {
			panic(err)
		}
		docs[i].LinkCitations = structures.URLAcademic + textLinkCitations
		//raccolgo i fields of study
		//per prendere tutte le informazioni devo andare alla pagina del documento:
		if err := wd.Get(URLDocuments[i]); err != nil {
			panic(err)
		}
		//aspetto di caricare la pagina (show more)
		wd.Wait(conditionDocumentPage2)
		//Espando gli "show more" di fields of study e sources
		expandShowMore(wd)
		//aspetto di caricare la pagina (i fields of study come riferimento)
		wd.Wait(conditionDocumentPage)
		//prendo i fields of study e sources
		setFieldsOfStudyAndSources(wd, &docs[i])
	}
	return docs, len(docs), maxCit
}

//Condizione per il caricamento della pagina iniziale: aspetto che si
//carichi la text box
func conditionMainPage(wd selenium.WebDriver) (bool, error) {
	elem, err := wd.FindElements(selenium.ByXPATH,
		"//ma-queryformulation[@class='searchWrap']/div//"+
			"input[@class='searchControl']")

	if err != nil {
		panic(err)
	}
	if len(elem) == 0 {
		return false, err
	}
	return true, err
}

//Restituisce il documento da cui inizia la ricerca
func GetInitialDocument_MA(wd selenium.WebDriver, phrase string) structures.MADocument {
	if err := wd.Get(structures.URLAcademic); err != nil {
		panic(err)
	}
	//stampo url
	url, _ := wd.CurrentURL()
	logger.Println("Url: ", url)

	fileS, _ := os.OpenFile("sorgenteInitialDoc.html", os.O_WRONLY, 0600)
	logger = log.New(fileS, "", 0)
	sorgente, err := wd.PageSource()
	if err != nil {
		panic(err)
	}
	logger.Println(sorgente)
	//Aspetto che si carichi la text box
	wd.Wait(conditionMainPage)

	fileS, _ = os.OpenFile("sorgenteInitialDocDopoWait.html", os.O_WRONLY, 0600)
	logger = log.New(fileS, "", 0)
	sorgente, err = wd.PageSource()
	if err != nil {
		panic(err)
	}
	logger.Println(sorgente)

	textBox, err := wd.FindElement(selenium.ByXPATH,
		"//ma-queryformulation[@class='searchWrap']/div//"+
			"input[@class='searchControl']")
	if err != nil {
		panic(err)
	}
	if err := textBox.SendKeys(phrase); err != nil {
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
	logger.Println("Url: ", url)

	//prendo il primo documento della pagina
	docs, numDocs := GetDocumentsFromPage_MA(wd, 1)
	if numDocs > 1 {
		panic(fmt.Sprint("GetInitialDocument - GetDocumentsFromPage\nHa resituito piu' di un documento"))
	}

	return docs[0]
}

func GetInitialDocumentByURL_MA(wd selenium.WebDriver, startURL string) structures.MADocument {
	if err := wd.Get(startURL); err != nil {
		panic(err)
	}
	//stampo url
	url, _ := wd.CurrentURL()
	logger.Println("Url: ", url)

	var initialDoc structures.MADocument

	//aspetto gli showmore
	//wd.WaitWithTimeout(conditionDocumentPage2, 60*time.Second)

	//prendo il titolo
	title, err := wd.FindElement(selenium.ByXPATH,
		"//div[@class='name-section']")
	if err != nil {
		panic(err)
	}
	initialDoc.Title, _ = title.Text()

	//Espando gli "show more" di fields of study e sources
	//expandShowMore(wd)

	//aspetto di caricare la pagina (i fields of study come riferimento)
	//wd.WaitWithTimeout(conditionDocumentPage, 60*time.Second)

	//prendo i fields of study e sources
	setFieldsOfStudyAndSources(wd, &initialDoc)

	//Prendo le citations (0), references (1) (opz. related (2))
	setCitationsAndReferences(wd, &initialDoc)

	//Add its URL
	initialDoc.URL = startURL

	return initialDoc
}

//Aspetto che la sezione in basso con i numeri delle pagine dei
//risultati si carichino (dove e' presente il link alla prossima pagina)
func conditionNextLink(wd selenium.WebDriver) (bool, error) {
	elem, err := wd.FindElements(selenium.ByXPATH, "//div[@class='entityResultPager']")

	if err != nil {
		panic(err)
	}
	if len(elem) == 0 {
		logger.Println("Aspetto che si carichino i link alle pagine successive dei rusultati")
		return false, err
	}
	return true, err
}

//Dato un link alla pagina di partenza, comincio a raccogliere i documenti (8 per pagina)
//finche' non arrivo a numDoc.
//Anche qui ho bisogno del numero delle pagine in cui sono distribuiti i doc che citano.
func GetCiteDocuments_MA(wd selenium.WebDriver, LinkCitations string, numDoc uint64, numPages int) ([]structures.MADocument, uint64) {
	if err := wd.Get(LinkCitations); err != nil {
		panic(err)
	}
	var allDoc []structures.MADocument
	//Mi serve per dire quanti documenti ho preso
	initialNumDoc := numDoc
	logger.Println("***** numDoc= " + strconv.FormatUint(numDoc, 10))

	//genero la sequenza di numeri casuali
	r := rand.New(rand.NewSource(12))

	for pageNumber := 1; pageNumber <= numPages; pageNumber++ {

		if pageNumber != 1 {
			//vado alla pagina successiva
			linkNextPage, err := wd.FindElement(selenium.ByXPATH,
				"//div[@class='entityResultPager']/ul/li/a[contains(text(),strconv.Itoa(pageNumber))]")
			//se non trovo il link per andare avanti, mi fermo
			if err != nil {
				if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
					logger.Println("\n\nSono uscito perche' non ho trovato il link alla prossima pagina\n")
					return allDoc, numDoc
				} else {
					panic(err)
				}
			}
			err = linkNextPage.Click()
			if err != nil {
				panic(err)
			}
		}
		//Salvo il link alla pagina dei documenti che citano perche' vado nelle
		//pagine dei singoli documenti che compaiono e non riesco a tornare indietro.
		currentUrl, err := wd.CurrentURL()
		if err != nil {
			panic(err)
		}

		newDoc, numNewDoc := GetDocumentsFromPage_MA(wd, int(numDoc))
		allDoc = append(allDoc, newDoc...)
		//tolgo il numero di documenti appena letti
		numDoc = numDoc - numNewDoc
		logger.Println("***** docRead= ", numNewDoc)
		logger.Println("***** numDoc= ", numDoc)

		if numDoc <= 0 {
			if numDoc == 0 {
				//tutto ok!
				return allDoc, initialNumDoc
			} else {
				//Qualcosa non va
				logger.Println("\nGetCiteDocuments_MA(): Ho raccolto piu' documenti di quelli che mi servivano!!")
				return allDoc, initialNumDoc
			}
		}

		//Torno alla pagina con i rusultati
		if err := wd.Get(currentUrl); err != nil {
			panic(err)
		}
		//Aspetto che si carichi entityResultPager dove sono presenti i link alle
		//varie pagine dei risultati.
		wd.WaitWithTimeout(conditionNextLink, 5000*time.Millisecond)

		/* Scorro in sequenza ma aspetto un tempo che cresce in modo esponenziale */
		waitTimeSec := time.Duration((math.Round(r.ExpFloat64())))
		time.Sleep(waitTimeSec * time.Second)
	}
	return allDoc, initialNumDoc - numDoc
}

//Aspetto che si carichi la sezione dei "Sort by"
func conditionSortBy(wd selenium.WebDriver) (bool, error) {
	elem, err := wd.FindElements(selenium.ByXPATH,
		"//div[@class='result-stats']/div/section/select/option")

	if err != nil {
		panic(err)
	}
	if len(elem) == 0 {
		logger.Println("Aspetto che si carichi sort by")
		return false, err
	}
	logger.Println("Numero option: ", len(elem))
	return true, err
}

//Raccolgie i documenti in base a una soglia sul numero di citazioni.
//Serve per creare l'albero
func GetCiteDocumentsByThreshold_MA(wd selenium.WebDriver, LinkCitations string, numPages, threshold, perc int) ([]structures.MADocument, int) {
	if err := wd.Get(LinkCitations); err != nil {
		panic(err)
	}
	var allDoc []structures.MADocument
	numDoc := 0

	//genero la sequenza di numeri casuali
	//r := rand.New(rand.NewSource(12))

	wd.WaitWithTimeout(conditionSortBy, 60*time.Second)
	//ordino i risultati per numero di citazioni decrescente, cosi' non appena
	//trovo un articolo sotto la soglia mi fermo.
	mostCitations, err := wd.FindElement(selenium.ByXPATH,
		"//div[@class='result-stats']/div/section/select/option[4]")
	if err != nil {
		logger.Println("Non sono riuscito ad ordinare i risultati!!!")
	} else {
		err = mostCitations.Click()
		if err != nil {
			panic(err)
		}
	}

	//This is the citations number of the first of result's articles
	//-1 indicates the search hasn't started yet
	maxCit := -1

	for pageNumber := 1; pageNumber <= numPages; pageNumber++ {
		//aspetto che si carichi la pagina, specialmente nel caso abbia
		//appena ordinato i risultati.
		/*waitTimeSec := time.Duration((math.Round(r.ExpFloat64())))
		logger.Println("Aspetto ", waitTimeSec, " secondi.")
		time.Sleep(waitTimeSec * time.Second)*/

		if pageNumber != 1 {
			//vado alla pagina successiva
			linkNextPage, err := wd.FindElement(selenium.ByXPATH,
				"//div[@class='entityResultPager']/ul/li/a[contains(text(),'"+strconv.Itoa(pageNumber)+"')]")
			//se non trovo il link per andare avanti, mi fermo
			if err != nil {
				if t, _ := regexp.MatchString(".*no such element.*", err.Error()); t {
					logger.Println("\n\nSono uscito perche' non ho trovato il link alla prossima pagina\n")
					return allDoc, numDoc
				} else {
					panic(err)
				}
			}
			err = linkNextPage.Click()
			if err != nil {
				panic(err)
			}
		}

		//Salvo il link alla pagina dei documenti che citano perche' vado nelle
		//pagine dei singoli documenti che compaiono e non riesco a tornare indietro.
		currentUrl, err := wd.CurrentURL()
		if err != nil {
			panic(err)
		}

		//se non considero la percentuale:
		//newDoc, numNewDoc, _ := GetDocumentsFromPageBasic_MA(wd, threshold)
		//e modificare la condizione nella funzione GetDocumentsFromPageBasic_MA
		newDoc, numNewDoc, maxCit := GetDocumentsFromPageBasic_MA(wd, maxCit, threshold, perc)
		allDoc = append(allDoc, newDoc...)
		//tolgo il numero di documenti appena letti
		numDoc = numDoc + numNewDoc
		logger.Println("***** docRead= ", numNewDoc)
		logger.Println("***** numDoc= ", numDoc)
		logger.Println("***** soglia_max= ", maxCit)

		//se ho preso meno di 8 doc, significa che sono sceso sotto la soglia
		//e mi fermo
		if numNewDoc < structures.NumArticlePerPageMA {
			return allDoc, numDoc
		}
		//Torno alla pagina con i rusultati
		if err := wd.Get(currentUrl); err != nil {
			panic(err)
		}
		//Aspetto che si carichi entityResultPager dove sono presenti i link alle
		//varie pagine dei risultati.
		wd.WaitWithTimeout(conditionNextLink, 60*time.Second)

		//Torno alla pagina con i rusultati
		//if err := wd.Get(currentUrl); err != nil {
		//	panic(err)
		//}
	}
	return allDoc, numDoc
}

func GetInfo(wd selenium.WebDriver, startURL string) (structures.MADocument, []string) {
	fmt.Println(startURL)
	if err := wd.Get(startURL); err != nil {
		panic(err)
	}
	//stampo url
	url, _ := wd.CurrentURL()
	logger.Println("Url: ", url)

	var initialDoc structures.MADocument

	//aspetto gli showmore
	//wd.WaitWithTimeout(conditionDocumentPage2, 60*time.Second)

	//prendo il titolo
	title, err := wd.FindElement(selenium.ByXPATH,
		"//div[@class='name-section']")
	if err != nil {
		panic(err)
	}
	initialDoc.Title, _ = title.Text()

	//Espando gli "show more" di fields of study e sources
	//expandShowMore(wd)

	//aspetto di caricare la pagina (i fields of study come riferimento)
	//wd.WaitWithTimeout(conditionDocumentPage, 60*time.Second)

	//prendo i fields of study e sources
	setFieldsOfStudyAndSources(wd, &initialDoc)

	//Prendo le citations (0), references (1) (opz. related (2))
	urlDocCitations := setCitationsAndReferences(wd, &initialDoc)

	//Add its URL
	initialDoc.URL = startURL

	return initialDoc, urlDocCitations
}

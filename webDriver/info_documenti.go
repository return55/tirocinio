package webDriver

import (
	"encoding/gob"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"math"
	"math/rand"
	"time"

	"github.com/return55/tirocinio/structures"

	"github.com/tebeka/selenium"
)

const (
	seleniumPath     = "webDriver/selenium-server-standalone-3.141.59.jar"
	geckoDriverPath  = "webDriver/geckodriver-v0.23.0-linux64/geckodriver"
	chromeDriverPath = "webDriver/chromedriver2.42"
	chromeBinary     = "/opt/google/chrome/chrome"
	defaultPort      = 8080
)

//Restituisco service solo per potrelo chiudere in main.go, non lo uso mai
//port sara' diverso da -1 solo nel caso stia aprendo delle connesioni per i thread
func StartSelenium(port int) (*selenium.Service, selenium.WebDriver) {
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),            // Output debug information to STDERR.
	}

	if port == -1 {
		port = defaultPort
	}

	var service *selenium.Service
	var err error
	service, err = selenium.NewSeleniumService(seleniumPath, port, opts...)
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

	if err = wd.SetImplicitWaitTimeout(300 * time.Second); err != nil {
		panic(err)
	}
	if err = wd.SetPageLoadTimeout(300 * time.Second); err != nil {
		panic(err)
	}

	return service, wd
}

//Data un pagina (impostata dal WebDriver) prendo un certo numero di documenti dalla pagina
//partendo dal primo in alto.
//Se il numero (numDocs) e' maggiore del numero di documenti nella pagina (tipicamente 10),
//mi limito a restituire i documenti presenti nella pagina e la loro quantita'.
func GetDocumentsFromPage(wd selenium.WebDriver, numDocs uint64, maxCit, threshold, perc int) ([]structures.Document, uint16) {
	//raccolgo le informazioni
	urls, err := wd.FindElements(selenium.ByXPATH, "//div[@class='gs_ri']/h3/a")
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

	documents := make([]structures.Document, numDocs)
	var text string
	var docIndex, otherIndex uint64
	//stampa-*---------------------------------------------
	fmt.Println("Lunghezza: ", len(urls))
	for docIndex, otherIndex = 0, 0; docIndex < numDocs; docIndex, otherIndex = docIndex+1, otherIndex+1 {
		//fmt.Println(docIndex, "------", otherIndex)
		documents[docIndex].Url, _ = urls[docIndex].GetAttribute("href")
		documents[docIndex].Title, _ = urls[docIndex].Text()

		text, _ = authors[docIndex].Text()
		leftSide := strings.Split(text, " -")[0]
		//!!!il carattere '…' che segue gli autori corrisponde a "\u2026" in utf-8!!!!
		leftSide = strings.Replace(leftSide, "\u2026", "", -1)
		documents[docIndex].Authors = strings.Split(leftSide, ", ")

		//questo pezzo può contenere la data
		//inizializzo a -1 nel caso non sia presente
		documents[docIndex].Date = -1
		center := strings.Split(text, " -")[1]
		//!!!il carattere '…' che segue gli autori corrisponde a "\u2026" in utf-8!!!!
		center = strings.Replace(center, "\u2026", "", -1)
		for _, val := range strings.Split(center, ",") {
			if date, err := strconv.ParseInt(val, 10, 64); err == nil {
				documents[docIndex].Date = date
				break
			}
		}

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
				if NumCitations, err := strconv.ParseUint(words[2], 10, 16); err != nil {
					panic(err)
				} else {
					documents[docIndex].NumCitations = uint16(NumCitations)
				}
				LinkCitations, _ := other[otherIndex].GetAttribute("href")
				documents[docIndex].LinkCitations = structures.URLScholar + LinkCitations
				break
			}
		}
		//se e' il primo risulatato della pagina e non ho ancora un valore valido di threshold
		if docIndex == 0 && maxCit == -1 {
			maxCit = int(documents[docIndex].NumCitations)
			logger.Println("Soglia: ", maxCit)
			/*Condozione banale sulla soglia
			if int(numsCitations[i]) < threshold {*/
			/*Condizione superiore a una percentuale di maxCit*/
		} else if documents[docIndex].NumCitations < uint16(threshold) || float32(documents[docIndex].NumCitations) < float32(maxCit)*(float32(perc)/100) {
			return documents[:docIndex-1], uint16(docIndex)
			break
		}
	}
	return documents, uint16(docIndex)
}

//Restituisce il documento da cui inizia la ricerca
func GetInitialDocument(wd selenium.WebDriver) structures.Document {
	if err := wd.Get(structures.URLScholar); err != nil {
		panic(err)
	}
	//stampo url
	url, _ := wd.CurrentURL()
	fmt.Println("Url: ", url)
	//vado alla pagina in inglese
	linkEnglishPage, err := wd.FindElement(selenium.ByXPATH, "//div[@id='gs_hp_eng']/a")
	if err != nil {
		panic(err)
	}
	if err := linkEnglishPage.Click(); err != nil {
		panic(err)
	}
	//stampo url
	url, _ = wd.CurrentURL()
	fmt.Println("Url: ", url)

	//cerco nella pagina il pulsante che appare solo se ho fatto il login
	/*roba, err := wd.FindElement(selenium.ByXPATH, "//a[@id='gs_hdr_act_s']")
	if err != nil {
		panic(err)
	}
	nome, err := roba.GetAttribute("alt")
	if err != nil {
		panic(err)
	}
	fmt.Println("Nome: ", nome)
	/////////////
	testo, err := roba.Text()
	if err != nil {
		panic(err)
	}
	fmt.Println("Testo: ", testo)*/

	textBox, err := wd.FindElement(selenium.ByID, "gs_hdr_tsi")
	if err != nil {
		panic(err)
	}
	if err := textBox.SendKeys(`GDPR`); err != nil {
		panic(err)
	}
	searchButton, err := wd.FindElement(selenium.ByID, "gs_hdr_tsb")
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
	docs, numDocs := GetDocumentsFromPage(wd, 1, -1, 0, 0)
	if numDocs > 1 {
		panic(fmt.Sprintf("GetInitialDocument - GetDocumentsFromPage\nHa resituito piu' di un documento"))
	}
	fmt.Println("\n\n", numDocs)
	return docs[0]
}

func GetFirstDocumentOfPage(wd selenium.WebDriver, url string) structures.Document {
	if err := wd.Get(url); err != nil {
		panic(err)
	}
	//prendo il primo documento della pagina
	docs, numDocs := GetDocumentsFromPage(wd, 1, -1, 0, 0)
	if numDocs > 1 {
		panic(fmt.Sprintf("GetFirstDocumentOfPage - GetDocumentsFromPage\nHa resituito piu' di un documento\n"))
	}
	return docs[0]
}

//Dato un link alla pagina di partenza, comincio a raccogliere i documenti (10 per pagina)
//finche' non arrivo a numDoc.
func GetCiteDocuments(wd selenium.WebDriver, LinkCitations string, numDoc uint64, threshold, perc int) ([]structures.Document, uint64) {
	if err := wd.Get(LinkCitations); err != nil {
		panic(err)
	}
	var allDoc []structures.Document
	var docRead uint64 = 0
	fmt.Println("***** docRead= " + strconv.FormatUint(docRead, 10))
	fmt.Println("***** numDoc= " + strconv.FormatUint(numDoc, 10))

	//genero la sequenza di numeri casuali
	r := rand.New(rand.NewSource(12))
	maxCit := -1
	for numDoc > docRead {

		newDoc, numNewDoc := GetDocumentsFromPage(wd, numDoc-docRead, maxCit, threshold, perc)
		allDoc = append(allDoc, newDoc...)
		//incremento il numero dei documenti letti
		docRead = docRead + uint64(numNewDoc)
		fmt.Println("***** docRead= " + strconv.FormatUint(docRead, 10))

		//Scorro una pagina alla volta in sequenza
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

		/* Scorro in sequenza ma aspetto un tempo che cresce in modo esponenziale */
		waitTimeSec := time.Duration((math.Round(r.ExpFloat64())))
		fmt.Println("Tempo di attesa2: ", waitTimeSec)
		time.Sleep(waitTimeSec * 10 * time.Second)

		if err := wd.Get(structures.URLScholar + url); err != nil {
			panic(err)
		}

	}
	return allDoc, docRead
}

//modifica perche' riceva un unico array di document
func PrintDocuments(allDoc []structures.Document) {
	if len(allDoc) == 0 {
		fmt.Println("Non ci sono documenti da stampare")
		return
	}
	fmt.Println("Documento iniziale:\nUrl: ", allDoc[0].Url, "\nAutori:")
	for _, autore := range allDoc[0].Authors {
		fmt.Println("\t", autore)
	}
	fmt.Println("Numero di documenti che lo citano: ", allDoc[0].NumCitations)
	fmt.Println("Link ai documenti che lo citano: ", allDoc[0].LinkCitations)

	fmt.Println("\nDocumento che citano:")
	for docIndex := 1; docIndex < len(allDoc); docIndex++ {
		fmt.Println("Url: ", allDoc[docIndex].Url, "\nAutori:")
		for _, autore := range allDoc[docIndex].Authors {
			fmt.Println("\t", autore)
		}
		fmt.Println("Numero di documenti che lo citano: ", allDoc[docIndex].NumCitations)
		fmt.Println("Link ai documenti che lo citano: ", allDoc[docIndex].LinkCitations)
	}
}

//salvo i documenti su un file
func SaveDocuments(allDoc []structures.MADocument) {
	file, err := os.Create(structures.SaveFilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	enc.Encode(allDoc)
}

//carico i documenti da file
func LoadDocuments(dim int) []structures.Document {
	allDoc := make([]structures.Document, dim)
	file, err := os.Open(structures.SaveFilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	dec.Decode(allDoc)
	fmt.Println(allDoc[0])
	return allDoc
}

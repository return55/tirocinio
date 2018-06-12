package main

import (
	"fmt"
	"os"
	"strconv"
	
	"github.com/return55/tirocinio/docDatabase"
	"github.com/return55/tirocinio/structures"
	"github.com/return55/tirocinio/webDriver"
	
	"github.com/tebeka/selenium"
)

//Partendo dal classico documento iniziale vado alla pagina di scholar con
//i documenti che lo citano e prendo il primo in alto.
//Ripeto il processo per il nuovo documento e vado avanti cosi' per n volte,
//dove n e' il primo parametro passato da riga di comando.
func GetEverFirst(wd selenium.WebDriver) bool {
	if len(os.Args) > 1 {
		repeatFor, err := strconv.ParseUint(os.Args[1], 10, 64)
		if err != nil{
			return false
		}
		fmt.Println("Numero iterazioni: ", repeatFor)
		allDoc := make([]structures.Document, repeatFor+1)
		initialDoc := webDriver.GetInitialDocument(wd)
		allDoc[0] = initialDoc
		
		var firstDoc structures.Document
		var i uint64
		for i=0; i<repeatFor; i++ {
			firstDoc = webDriver.GetFirstDocumentOfPage(wd, allDoc[i].LinkCitedBy)
			allDoc[i+1] = firstDoc
		}
		
		docDatabase.DBGetEverFirst(allDoc)
		return true
	}
	return false
}

//Partendo dal classico documento iniziale vado alla pagina di scholar con
//i documenti che lo citano e prendo i primi n in classifica, dove n e' il primo
//parametro passato da riga di comando  e ne creo il database.
func GetFirstsNDoc(wd selenium.WebDriver) bool {
	if len(os.Args) > 1 {
		numDocs, err := strconv.ParseUint(os.Args[1], 10, 64)
		if err != nil{
			return false
		}
		fmt.Println("Numero documenti: ", numDocs)
		
		initialDoc := webDriver.GetInitialDocument(wd)

		citeInitialDoc, _ := webDriver.GetCiteDocument(wd, initialDoc, numDocs)
		
		allDoc := append(citeInitialDoc, initialDoc)
		allDoc[0], allDoc[len(allDoc)-1] = allDoc[len(allDoc)-1], allDoc[0]
		
		docDatabase.DBGetFirstsNDoc(allDoc)
		return true
	}
	return false
}

func main() {
	service, wd := webDriver.StartSelenium()

	defer service.Stop()
	defer wd.Quit()


	//Classico: docIniziale + primi n che lo citano
	if GetFirstsNDoc(wd) {
		fmt.Println("Tutto ok")
	}else{
		fmt.Println("Non hai passato un valore al main")
	}
/*	
	//Sempre Il Primo: docIniziale + n volte sempre il primo della classifica
	if GetEverFirst(wd) {
		fmt.Println("Tutto ok")
	}else{
		fmt.Println("Hai passato un valore non corretto al main")
	}
*/
/*	//Metodi utili
	webDriver.SaveDocuments(nil)
	webDriver.LoadDocuments(allDoc)
	webDriver.PrintDocuments(allDoc)
*/	
}








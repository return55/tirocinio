package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/return55/tirocinio/docDatabase"
	"github.com/return55/tirocinio/structures"
	"github.com/return55/tirocinio/webDriver"

	"github.com/tebeka/selenium"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

//Partendo dal classico documento iniziale vado alla pagina di scholar con
//i documenti che lo citano e prendo il primo in alto.
//Ripeto il processo per il nuovo documento e vado avanti cosi' per n volte,
//dove n e' il primo parametro passato da riga di comando.
func GetEverFirst_MA(wd selenium.WebDriver) bool {
	if len(os.Args) > 1 {
		repeatFor, err := strconv.ParseUint(os.Args[2], 10, 64)
		if err != nil {
			return false
		}
		fmt.Println("Numero iterazioni: ", repeatFor)
		allDoc := make([]structures.MADocument, repeatFor+1)
		initialDoc := webDriver.GetInitialDocument_MA(wd)
		allDoc[0] = initialDoc

		var firstDoc structures.MADocument
		var i uint64
		for i = 0; i < repeatFor; i++ {
			//firstDoc = webDriver.GetFirstDocumentOfPage_MA(wd, allDoc[i].LinkCitations)
			allDoc[i+1] = firstDoc
		}

		//fase neo4j
		conn := docDatabase.StartNeo4j()
		defer conn.Close()
		//pulisco il db
		docDatabase.CleanAll(conn)
		//aggiungo il documento iniziale
		docDatabase.AddDocument_MA(conn, allDoc[0], "")
		for docIndex := 1; docIndex < len(allDoc); docIndex++ {
			docDatabase.AddDocument_MA(conn, allDoc[docIndex], allDoc[0].Title)
		}
		return true
	}
	return false
}

//Costruisce l'albero delle citazioni a partire dal documento iniziale,
//posso decidere il numero dei livelli da input, per ora la soglia la decido io.
//NOTA:
//Non conoscero' i documenti che citano le foglie del mio albero
func creaAlberoCitazioni_MA(wd selenium.WebDriver) bool {
	if len(os.Args) > 1 {
		numLevels, err := strconv.ParseUint(os.Args[1], 10, 64)
		if err != nil {
			return false
		}
		fmt.Println("Numero livelli: ", numLevels)

		var allDoc []structures.MADocument
		initialDoc := webDriver.GetInitialDocument_MA(wd)
		allDoc = append(allDoc, initialDoc)

		if initialDoc.LinkCitations == "" {
			fmt.Println("\n\nIL DOCUMENTO DI PARTENZA NON E' CITATO DA NESSUNO\n")
			return false
		}
		//il numero delle pagine in cui sono distribuiti i risultati
		numPages := int((initialDoc.NumCitations / structures.NumArticlePerPageMA) + 1)
		citeInitialDoc, _ := webDriver.GetCiteDocumentsByThreshold_MA(wd, initialDoc.LinkCitations, numPages, 200)
		allDoc = append(allDoc, citeInitialDoc...)

		//fase neo4j
		conn := docDatabase.StartNeo4j()
		defer conn.Close()
		//pulisco il db
		docDatabase.CleanAll(conn)
		//aggiungo il documento iniziale
		docDatabase.AddDocument_MA(conn, allDoc[0], "")
		for docIndex := 1; docIndex < len(allDoc); docIndex++ {
			docDatabase.AddDocumentBasic_MA(conn, allDoc[docIndex], allDoc[0].Title)
		}

		//sono i documenti ancora da esplorare ovvero i figli appena creati
		parentDocs := allDoc[1:]
		var childDocs []structures.MADocument
		for ; numLevels > 0; numLevels-- {
			for _, doc := range parentDocs {
				childDocs = append(childDocs, getFirstsNDoc_MA(wd, doc, conn, 200)...)
			}
			parentDocs = childDocs
			childDocs = nil
		}
		return true
	}
	return false
}

//Solo di supporto: le passo il doc di partenza e vado alla pagina di scholar con
//i documenti che lo citano e prendo quelli con numero citazioni > soglia, infine
//li aggiungo al database.
func getFirstsNDoc_MA(wd selenium.WebDriver, initialDoc structures.MADocument, conn bolt.Conn, threshold int) []structures.MADocument {

	numPages := int((initialDoc.NumCitations / structures.NumArticlePerPageMA) + 1)
	citeInitialDoc, numFigli := webDriver.GetCiteDocumentsByThreshold_MA(wd, initialDoc.LinkCitations, numPages, threshold)

	for docIndex := 1; docIndex < len(citeInitialDoc); docIndex++ {
		docDatabase.AddDocumentBasic_MA(conn, citeInitialDoc[docIndex], initialDoc.Title)
	}
	fmt.Println("Titolo: ", initialDoc.Title, " -- num figli: ", numFigli)
	return citeInitialDoc
}

func main() {
	/*if len(os.Args) < 3 {
		fmt.Println("I parametri da passare al main possono essere: (everFirst | firstN | thread) num1 num2")
		return
	}*/

	service, wd := webDriver.StartSelenium(-1)

	defer service.Stop()
	defer wd.Quit()

	if t := creaAlberoCitazioni_MA(wd); t {
		fmt.Println("TUTTO OK")
	}

	/*
		switch os.Args[1] {
		//Classico: docIniziale + primi n che lo citano
		case "everFirst":
			if GetEverFirst(wd) {
				fmt.Println("Tutto ok")
			} else {
				fmt.Println("Parametri da passare: 'everFirst' numDocCheCitano")
			}
		//Sempre Il Primo: docIniziale + n volte sempre il primo della classifica
		case "firstN":
			if GetFirstsNDoc(wd) {
				fmt.Println("Tutto ok")
			} else {
				fmt.Println("Parametri da passare: 'firstN' numDoc")
			}
		case "thread":
			if Concurrency(wd) {
				fmt.Println("Tutto ok")
			} else {
				fmt.Println("Parametri da passare: 'thread' numThreads docPerLink lenLinkList")
			}
		default:
			fmt.Println("I parametri da passare al main possono essere: (everFirst | firstN | thread) num1 num2 ...")
		}*/

	/*	//Metodi utili
		webDriver.SaveDocuments(nil)
		webDriver.LoadDocuments(allDoc)
		webDriver.PrintDocuments(allDoc)
	*/
}

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/return55/tirocinio/docDatabase"
	"github.com/return55/tirocinio/structures"
	"github.com/return55/tirocinio/webDriver"

	"github.com/tebeka/selenium"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

//Costruisce l'albero delle citazioni a partire dal documento iniziale,
//posso decidere il numero dei livelli da input, per ora la soglia la decido io.
//NOTA:
//Non conoscero' i documenti che citano le foglie del mio albero
//"go run main_MA.go NUM_LIVELLI SOGLIA"
func creaAlberoCitazioni_MA(wd selenium.WebDriver) bool {
	if len(os.Args) > 2 {
		numLevels, err := strconv.ParseUint(os.Args[1], 10, 64)
		if err != nil {
			fmt.Println("Inserisci un numero quando chiami il main !!!!!")
			return false
		}
		threshold, err := strconv.ParseUint(os.Args[2], 10, 64)
		if err != nil {
			fmt.Println("Inserisci un numero quando chiami il main !!!!!")
			return false
		}
		fmt.Println("Numero livelli: ", numLevels)
		fmt.Println("Soglia: ", threshold)

		var allDoc []structures.MADocument
		initialDoc := webDriver.GetInitialDocument_MA(wd)
		allDoc = append(allDoc, initialDoc)
		fmt.Println("Doc: ", initialDoc)
		if initialDoc.LinkCitations == "" {
			fmt.Println("\n\nIL DOCUMENTO DI PARTENZA NON E' CITATO DA NESSUNO\n")
			return false
		}
		//il numero delle pagine in cui sono distribuiti i risultati
		numPages := int((initialDoc.NumCitations / structures.NumArticlePerPageMA) + 1)
		citeInitialDoc, _ := webDriver.GetCiteDocumentsByThreshold_MA(wd, initialDoc.LinkCitations, numPages, int(threshold))
		allDoc = append(allDoc, citeInitialDoc...)

		//fase neo4j
		conn := docDatabase.StartNeo4j()
		defer conn.Close()
		//pulisco il db
		docDatabase.CleanAll(conn)

		//aggiungo il documento iniziale
		docDatabase.AddDocumentBasic_MA(conn, allDoc[0], "")
		for docIndex := 1; docIndex < len(allDoc); docIndex++ {
			docDatabase.AddDocumentBasic_MA(conn, allDoc[docIndex], allDoc[0].Title)
		}

		//mi serve per tenere traccia dei livelli
		fileMA, _ := os.OpenFile("Quali_livelli_ho_fatto", os.O_WRONLY, 0600)
		logger := log.New(fileMA, "", 0)

		//sono i documenti ancora da esplorare ovvero i figli appena creati
		parentDocs := allDoc[1:]
		var childDocs []structures.MADocument
		for livelli := numLevels; numLevels > 0 && parentDocs != nil; numLevels-- {
			for _, doc := range parentDocs {
				//prima di esplorare un doc controllo se l'ho gia' esplorato
				if !docDatabase.AlreadyExplored(conn, doc.Title) {
					childDocs = append(childDocs, getFirstsNDoc_MA(wd, doc, conn, int(threshold))...)
				}
			}
			parentDocs = childDocs
			childDocs = nil
			logger.Println("Ho finito il livello: ", livelli-numLevels+1)
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

	for docIndex := 0; docIndex < len(citeInitialDoc); docIndex++ {
		docDatabase.AddDocumentBasic_MA(conn, citeInitialDoc[docIndex], initialDoc.Title)
	}
	fmt.Println("Titolo: ", initialDoc.Title, " -- num figli: ", numFigli)
	return citeInitialDoc
}

func main() {

	service, wd := webDriver.StartSelenium(-1)

	defer service.Stop()
	defer wd.Quit()

	if t := creaAlberoCitazioni_MA(wd); t {
		fmt.Println("TUTTO OK")
	} else {
		fmt.Println("Inserisci 2 numeri!!!!")
	}

}
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
		repeatFor, err := strconv.ParseUint(os.Args[2], 10, 64)
		if err != nil {
			return false
		}
		fmt.Println("Numero iterazioni: ", repeatFor)
		allDoc := make([]structures.Document, repeatFor+1)
		initialDoc := webDriver.GetInitialDocument(wd)
		allDoc[0] = initialDoc

		var firstDoc structures.Document
		var i uint64
		for i = 0; i < repeatFor; i++ {
			firstDoc = webDriver.GetFirstDocumentOfPage(wd, allDoc[i].LinkCitedBy)
			allDoc[i+1] = firstDoc
		}

		//fase neo4j
		conn := docDatabase.StartNeo4j()
		defer conn.Close()
		//pulisco il db
		docDatabase.CleanAll(conn)
		//aggiungo il documento iniziale
		docDatabase.AddDocument(conn, allDoc[0], "")
		for docIndex := 1; docIndex < len(allDoc); docIndex++ {
			docDatabase.AddDocument(conn, allDoc[docIndex], allDoc[0].Url)
		}
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
		if err != nil {
			return false
		}
		fmt.Println("Numero documenti: ", numDocs)

		initialDoc := webDriver.GetInitialDocument(wd)

		citeInitialDoc, _ := webDriver.GetCiteDocuments(wd, initialDoc.LinkCitedBy, numDocs)

		allDoc := append(citeInitialDoc, initialDoc)
		allDoc[0], allDoc[len(allDoc)-1] = allDoc[len(allDoc)-1], allDoc[0]

		//fase neo4j
		conn := docDatabase.StartNeo4j()
		defer conn.Close()
		//pulisco il db
		docDatabase.CleanAll(conn)
		//aggiungo il documento iniziale
		docDatabase.AddDocument(conn, allDoc[0], "")
		for docIndex := 1; docIndex < len(allDoc); docIndex++ {
			docDatabase.AddDocument(conn, allDoc[docIndex], allDoc[0].Url)
		}
		return true
	}
	return false
}

//Raccolgo documenti utilizzando dei threads:
//1)Faccio partire N threads che sono in attesa su un canale contenete i link delle pagine che citano (links)
//2)Partendo da un primo documento scelto da me, ne aggiungo il link ai documenti che lo citano al canale
//2.1)Aggiungo quel documento al db
//3)Parte il primo thread che
func Concurrency(wd selenium.WebDriver) bool {
	numThreads, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err != nil {
		return false
	}
	docPerPage, err := strconv.ParseUint(os.Args[3], 10, 64)
	if err != nil {
		return false
	}
	lenLinkList, err := strconv.ParseUint(os.Args[4], 10, 64)
	if err != nil {
		return false
	}
	links := make(chan string, lenLinkList)
	//canale su cui i thread mettono il numero di documenti letti
	//NOTA: basterebbe una lunghezza di 1 visto che la routine di Concurrency
	//non fa altro che estrarre elementi dal canale...in ogni caso la imposto a
	//4 per avere un po' di margine.
	chanNumNewDoc := make(chan uint64, 4)

	//faccio partire i threads di ricerca dei documenti
	var id uint64
	for id = 1; id <= numThreads; id++ {
		go threadGetDocument(id, docPerPage, links, chanNumNewDoc)
	}

	initialDoc := webDriver.GetInitialDocument(wd)
	//mi collego a neo4j e aggiungo il primo documento
	conn := docDatabase.StartNeo4j()
	defer conn.Close()
	//pulisco il db
	docDatabase.CleanAll(conn)
	//aggiungo il documento iniziale
	docDatabase.AddDocument(conn, initialDoc, "")

	//aggiungo il suo link ai doc che lo citano alla lista (links)
	fakeList := make([]string, 1)
	fakeList[0] = initialDoc.LinkCitedBy
	go threadAddLinks(fakeList, links, 0)

	//devo aggiungere un controllo per l'uscita dal programma
	//es.
	//tempo trascorso
	//numero documenti raccolti
	var totReadDoc uint64 = 0
	for totReadDoc < structures.MaxReadableDoc {
		totReadDoc += <-chanNumNewDoc
	}

	return true
}

//Thread che si occupa di estrarre docPerPage documenti dalla pagina indicata
//dal link che estrae dalla lista, invoca un altro thread che aggiunge i LinkCitedBy
//alla lista links, aggiunge i documenti al database.
func threadGetDocument(id uint64, docPerPage uint64, links chan string, chanNumNewDoc chan uint64) {
	//creo il web driver personale
	service, wd := webDriver.StartSelenium()
	defer service.Stop()
	defer wd.Quit()
	//creo la connesione con neo4j personale
	conn := docDatabase.StartNeo4j()
	defer conn.Close()

	for true {
		startLink := <-links
		newDocuments, numNewDoc := webDriver.GetCiteDocuments(wd, startLink, docPerPage)
		fmt.Println("Thread ", id, ": doc letti = ", numNewDoc)
		//creo la lista dei nuovi links ai citedBy
		newLinks := make([]string, numNewDoc)
		for index, doc := range newDocuments {
			newLinks[index] = doc.LinkCitedBy
		}
		//chiamo routine che si occupa di aggiungere i nuovi link alla coda
		go threadAddLinks(newLinks, links, id)
		//fase neo4j
		//ricavo l'URL del documento che ha: linkCitedBy = startLink
		rows, err := conn.QueryNeo("MATCH (doc:Document {linkCitedBy: {LinkCitedBy}})"+
			"RETURN doc.url", map[string]interface{}{"LinkCitedBy": startLink})
		if err != nil {
			panic(err)
		}
		URL, _, err := rows.NextNeo()
		if err != nil {
			panic(err)
		}
		fmt.Println("Thread ", id, ": link = ", URL[0].(string))
		//aggiungo i nuovi documenti al database
		for _, newDoc := range newDocuments {
			docDatabase.AddDocument(conn, newDoc, URL[0].(string))
		}
		//aggiungo il numero  dei documenti letti al canale
		chanNumNewDoc <- numNewDoc
	}
}

//Thread che si occupa di aggiungere i link che gli passo alla lista (canale)
func threadAddLinks(newLinks []string, links chan string, id uint64) {
	for _, link := range newLinks {
		links <- link
	}
	fmt.Println("AddLinks chiamato da ", id, " e' terminato.")
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("I parametri da passare al main possono essere: (everFirst | firstN | thread) num1 num2")
		return
	}

	service, wd := webDriver.StartSelenium()

	defer service.Stop()
	defer wd.Quit()

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
			fmt.Println("Parametri da passare: 'thread' numThreads docPerPage lenLinkList")
		}
	default:
		fmt.Println("I parametri da passare al main possono essere: (everFirst | firstN | thread) num1 num2")
	}

	/*	//Metodi utili
		webDriver.SaveDocuments(nil)
		webDriver.LoadDocuments(allDoc)
		webDriver.PrintDocuments(allDoc)
	*/
}

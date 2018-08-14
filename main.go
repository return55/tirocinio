package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/return55/tirocinio/docDatabase"
	"github.com/return55/tirocinio/structures"
	"github.com/return55/tirocinio/webDriver"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/tebeka/selenium"
)

var (
	//creo il logger per i thread
	fileThreadTimes, _ = os.OpenFile("thread_times.LOG", os.O_WRONLY, 0600)
	logger             = log.New(fileThreadTimes, "", 0)
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
		initialDoc := webDriver.GetInitialDocument_MA(wd)
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
		numDocs, err := strconv.ParseUint(os.Args[2], 10, 64)
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

//La uso per misurare il tempo impiegato da un thread per raccogliere i suoi
//documenti.
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	logger.Printf("%s ha impiegato %s", name, elapsed)
}

//Raccolgo documenti utilizzando dei threads:
//1)Faccio partire N threads che sono in attesa su un canale contenete i link delle pagine che citano (links)
//2)Partendo da un primo documento scelto da me, ne aggiungo il link ai documenti che lo citano al canale
//2.1)Aggiungo quel documento al db
//3)Parte il primo thread che aggiunge il documento iniziale al db e il suo linkCitedBy alla lista
//4)Questa funzione termina quando ho letto un certo numero di documenti
func Concurrency(wd selenium.WebDriver) bool {
	//stampo il tempo impiegato dalla funzione
	defer timeTrack(time.Now(), "Concurrency")
	logger.Println("ciao")
	numThreads, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err != nil {
		return false
	}
	docPerLink, err := strconv.ParseUint(os.Args[3], 10, 64)
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

	initialDoc := webDriver.GetInitialDocument(wd)
	//mi collego a neo4j e apro un pool di connessioni
	pool := docDatabase.StartPoolNeo4j(int(numThreads) + 1)
	for _, conn := range pool {
		defer conn.Close()
	}
	//connessione di Concurrency
	concurrencyConn := pool[0]

	//faccio partire i threads di ricerca dei documenti
	var id uint64
	for id = 1; id <= numThreads; id++ {
		go threadGetDocument(id, docPerLink, links, chanNumNewDoc, pool[id])
	}

	//pulisco il db
	docDatabase.CleanAll(concurrencyConn)
	//aggiungo il documento iniziale
	docDatabase.AddDocument(concurrencyConn, initialDoc, "")

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
		logger.Println("Main, doc letti = " + strconv.FormatUint(totReadDoc, 10))
	}

	return true
}

//Thread che si occupa di estrarre docPerLink documenti dalla pagina indicata
//dal link che estrae dalla lista, invoca un altro thread che aggiunge i LinkCitedBy
//alla lista links, aggiunge i documenti al database e invia a Cuncurrency il
//numero di documenti letti.
func threadGetDocument(id uint64, docPerLink uint64, links chan string, chanNumNewDoc chan uint64, conn bolt.Conn) {
	//misuro il tempo in cui il thread rimane in esecuzione
	defer timeTrack(time.Now(), "Thread "+strconv.FormatUint(id, 10)+" (fine)")
	//creo il web driver personale
	service, wd := webDriver.StartSelenium(structures.ThreadBasePort + int(id))
	defer service.Stop()
	defer wd.Quit()

	for iteration := 1; ; iteration++ {
		//per misurae il tempo trascorso per un singolo link
		startIterationTime := time.Now()

		startLink := <-links
		fmt.Println("--------------URL: ", startLink)
		newDocuments, numNewDoc := webDriver.GetCiteDocuments(wd, startLink, docPerLink)
		logger.Println("Thread ", id, ": doc letti = ", numNewDoc)
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
		//chiudo lo stream
		err = rows.Close()
		if err != nil {
			panic(err)
		}
		logger.Println("Thread ", id, ": link = ", URL[0].(string))
		//aggiungo i nuovi documenti al database
		//togli la i!!!!!!!!!!!!!!!!!!!!!!!!!
		for i, newDoc := range newDocuments {
			docDatabase.AddDocument(conn, newDoc, URL[0].(string))
			logger.Println("Thread id= ", id, " ha scritto doc numero ", i)
		}
		//aggiungo il numero  dei documenti letti al canale
		chanNumNewDoc <- numNewDoc
		//stampo il tempo trascorso dall'inizio dell'iterazione
		timeTrack(startIterationTime, "Thread "+strconv.FormatUint(id, 10)+" iterazione "+
			strconv.FormatInt(int64(iteration), 10))
	}
}

//Thread che si occupa di aggiungere i link che gli passo alla lista (canale)
func threadAddLinks(newLinks []string, links chan string, id uint64) {
	for _, link := range newLinks {
		links <- link
	}
	logger.Println("AddLinks chiamato da ", id, " e' terminato.")
}

func main2() {
	if len(os.Args) < 3 {
		fmt.Println("I parametri da passare al main possono essere: (everFirst | firstN | thread) num1 num2")
		return
	}

	service, wd := webDriver.StartSelenium(-1)

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
			fmt.Println("Parametri da passare: 'thread' numThreads docPerLink lenLinkList")
		}
	default:
		fmt.Println("I parametri da passare al main possono essere: (everFirst | firstN | thread) num1 num2 ...")
	}

	/*	//Metodi utili
		webDriver.SaveDocuments(nil)
		webDriver.LoadDocuments(allDoc)
		webDriver.PrintDocuments(allDoc)
	*/
}

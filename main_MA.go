package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/return55/tirocinio/docDatabase"
	"github.com/return55/tirocinio/structures"
	"github.com/return55/tirocinio/webDriver"

	"github.com/tebeka/selenium"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

//Costruisce l'albero delle citazioni a partire dal documento iniziale,
//posso decidere il numero dei levels da input, per ora la soglia la decido io.
//NOTA:
//Non conoscero' i documenti che citano le foglie del mio albero
func createCitationsTree_MA(wd selenium.WebDriver, startingPoint string, maxNumLv, threshold int, quit chan int8) bool {

	var allDoc []structures.MADocument
	var initialDoc structures.MADocument
	if t := strings.Contains(startingPoint, "https://academic.microsoft.com/#/detail/"); t {
		initialDoc = webDriver.GetInitialDocumentByURL_MA(wd, startingPoint)
		fmt.Println("URL partenza: ", startingPoint)
	} else {
		initialDoc = webDriver.GetInitialDocument_MA(wd, startingPoint)
	}

	allDoc = append(allDoc, initialDoc)
	fmt.Println("Doc: ", initialDoc)
	if initialDoc.LinkCitations == "" {
		fmt.Println("\n\nIL DOCUMENTO DI PARTENZA NON E' CITATO DA NESSUNO\n")
		return false
	}
	//il numero delle pagine in cui sono distribuiti i risultati
	numPages := int((initialDoc.NumCitations / structures.NumArticlePerPageMA) + 1)
	citeInitialDoc, _ := webDriver.GetCiteDocumentsByThreshold_MA(wd, initialDoc.LinkCitations, numPages, threshold)
	allDoc = append(allDoc, citeInitialDoc...)

	fmt.Println(allDoc)

	//fase neo4j
	conn := docDatabase.StartNeo4j()
	defer conn.Close()
	//pulisco il db
	docDatabase.CleanAll(conn)
	// SOLO PER QUANDO USO IL CAMPO GENERICO !!
	result, err := conn.ExecNeo("MERGE (doc:MAFieldOfStudy { name: 'Generic'})",
		map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	numResult, _ := result.RowsAffected()
	fmt.Printf("Creato campo generico : %d\n", numResult)
	// FINE

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
	for level := maxNumLv; level > 0 && parentDocs != nil; level-- {
		for _, doc := range parentDocs {
			select {
			case _, ok := <-quit:
				if ok {
					logger.Println("Stopped at level: ", maxNumLv-level+1)
					close(quit)
					return true
				} else {
					logger.Println("The channel has been closed unexpectedly")
					return false
				}
			default:
				//prima di esplorare un doc controllo se l'ho gia' esplorato
				if !docDatabase.AlreadyExplored(conn, doc.Title) {
					childDocs = append(childDocs, getFirstsNDoc_MA(wd, doc, conn, threshold)...)
				}
			}
		}
		parentDocs = childDocs
		childDocs = nil
		logger.Println("parent: ", parentDocs)
		logger.Println("child: ", childDocs)
		logger.Println("Ho finito il livello: ", maxNumLv-level+1)
	}
	return true
}

//Solo di supporto: le passo il doc di partenza e vado alla pagina di scholar con
//i documenti che lo citano e prendo quelli con numero citazioni > soglia, infine
//li aggiungo al database.
//"go run main_MA.go NUM_levels SOGLIA <URL_PRIMO_DOC>"
func getFirstsNDoc_MA(wd selenium.WebDriver, initialDoc structures.MADocument, conn bolt.Conn, threshold int) []structures.MADocument {

	numPages := int((initialDoc.NumCitations / structures.NumArticlePerPageMA) + 1)
	citeInitialDoc, numFigli := webDriver.GetCiteDocumentsByThreshold_MA(wd, initialDoc.LinkCitations, numPages, threshold)

	for docIndex := 0; docIndex < len(citeInitialDoc); docIndex++ {
		docDatabase.AddDocumentBasic_MA(conn, citeInitialDoc[docIndex], initialDoc.Title)
	}
	fmt.Println("Titolo: ", initialDoc.Title, " -- num figli: ", numFigli)
	return citeInitialDoc
}

/*
Il documento iniziale lo prendo da info_documenti_MA/GetInitialDocument_MA() facendo una ricerca nella home
di Microsoft Academic e prendendo il primo documento tra quelli restituiti.
Per cambiare il documento si puo' cambiare la parola da cercare nella funzione textBox.SendKeys() (riga 688)

Oppure se al main viene passato un terzo argomento, questo viene considerato l'URL della pagina di Academic del
primo documento.
*/
func main5() {

	service, wd := webDriver.StartSelenium(-1)

	defer service.Stop()
	defer wd.Quit()

	//Get user input
	reader := bufio.NewReader(os.Stdin)
	//could be an Academic URL or a phrase to search
	startingPoint := ""
	for startingPoint == "" {
		fmt.Println("Insert the URl of an article or a phrase (one or more words) to search:")
		startingPoint, _ = reader.ReadString('\n')
		startingPoint = strings.Replace(startingPoint, "\n", "", -1)
	}
	var maxNumLv int = 0
	var err error
	for maxNumLv <= 0 {
		fmt.Println("Please specify the most number of search iterations (e.g. 10):")
		str, _ := reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		maxNumLv, err = strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please insert a non negative number")
			continue
		}
	}
	var threshold int = 0
	for threshold <= 0 {
		fmt.Println("Now the least citations number (threashold) (e.g. 500):")
		str, _ := reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		threshold, err = strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please insert a non negative number")
			continue
		}
	}

	//Start collecting articles and writing them on neo4j, it stops when the user writes "stop"
	var wg sync.WaitGroup
	wg.Add(2)
	//if the user writes "stop", the second goroutine stops the first through this channel
	quit := make(chan int8, 1)
	//if the function ends by itself, the first goroutine closes the channel to stop the second

	var t bool
	go func() {
		t = createCitationsTree_MA(wd, startingPoint, maxNumLv, threshold, quit)
		wg.Done()
	}()
	fmt.Println("The search has begun")

	go func() {
		defer wg.Done()
		fmt.Println("Type \"stop\" to stop searching or wait until it ends:")
		stdin := make(chan string, 1)
		//this routine checks when the user insert some values
		go func() {
			str, _ := reader.ReadString('\n')
			stdin <- strings.Replace(strings.Replace(str, "\n", "", -1), " ", "", -1)
		}()
		for {
			time.Sleep(3 * time.Second)
			//if the channel (quit) is closed, the first routine has already terminate
			select {
			case _, ok := <-quit:
				if !ok {
					fmt.Println("Channel closed correctly")
					return
				} else {
					fmt.Println("The channel is open and someone writes on it ??")
				}
			case str, _ := <-stdin:
				if str == "stop" {
					quit <- 1
					close(stdin)
					return
				} else {
					fmt.Println("Type \"stop\" to stop searching or wait until it ends:")
					//this routine checks when the user insert some values
					go func() {
						str, _ := reader.ReadString('\n')
						stdin <- strings.Replace(strings.Replace(str, "\n", "", -1), " ", "", -1)
					}()
				}
			}
		}
	}()

	wg.Wait()

	if t {
		fmt.Println("ALL OK")
	} else {
		fmt.Println("There could be some problem, check the file: Quali_livelli_ho_fatto")
	}

	//Show the functionality
	for {
		fmt.Println("Select a function:\n0) Print the fields ranking")
		str, _ := reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		choice, err := strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please insert an integer value")
			continue
		}
		switch choice {
		case 0:
			choice = 0
			for choice <= 0 && choice != -1 {
				fmt.Println("Insert the number of fields to show (e.g. 10 for top10):	")
				str, _ := reader.ReadString('\n')
				str = strings.Replace(str, "\n", "", -1)
				choice, err = strconv.Atoi(str)
				if err != nil {
					fmt.Println("Please insert a non negative number")
				}
			}
			conn := docDatabase.StartNeo4j()
			ranking := docDatabase.FieldsRanking(conn, choice)
			fmt.Println("RANKING:\nSCORE\tFIELD")
			for field, score := range ranking {
				fmt.Println(strconv.Itoa(score) + "\t" + field)
			}
			conn.Close()

		}
	}

}

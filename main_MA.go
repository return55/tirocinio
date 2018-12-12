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

var (
	//mi serve per tenere traccia dei livelli
	fileMA, _      = os.OpenFile("Quali_livelli_ho_fatto", os.O_WRONLY, 0600)
	logger         = log.New(fileMA, "", 0)
	reader         = bufio.NewReader(os.Stdin)
	researchNumber = 1
)

//Costruisce l'albero delle citazioni a partire dal documento iniziale,
//posso decidere il numero dei levels da input, per ora la soglia la decido io.
//NOTA:
//Non conoscero' i documenti che citano le foglie del mio albero
func createCitationsTree_MA(wd selenium.WebDriver, startingPoint string, maxNumLv, threshold, perc, graphNumber int, quit chan int8) bool {

	var allDoc []structures.MADocument
	var initialDoc structures.MADocument
	if t := strings.Contains(startingPoint, "https://academic.microsoft.com/#/detail/"); t {
		initialDoc = webDriver.GetInitialDocumentByURL_MA(wd, startingPoint)
		fmt.Println("URL partenza: ", startingPoint)
	} else {
		fmt.Println("Parole da cercare: ", startingPoint)
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
	citeInitialDoc, _ := webDriver.GetCiteDocumentsByThreshold_MA(wd, initialDoc.LinkCitations, numPages, threshold, perc)
	allDoc = append(allDoc, citeInitialDoc...)

	fmt.Println(allDoc)

	//fase neo4j
	conn := docDatabase.StartNeo4j()
	defer conn.Close()
	//pulisco il db
	docDatabase.CleanAll(conn)
	// SOLO PER QUANDO USO IL CAMPO GENERICO !!
	/*result, err := conn.ExecNeo("MERGE (doc:MAFieldOfStudy { name: 'Generic'})",
		map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	numResult, _ := result.RowsAffected()
	fmt.Printf("Creato campo generico : %d\n", numResult)*/
	// FINE

	//aggiungo il documento iniziale e i suoi figli
	//if this condition is true, i have to create a new graph
	if graphNumber == researchNumber {
		researchNumber++
	}
	docDatabase.AddDocumentBasic_MA(conn, allDoc[0], "", string(graphNumber))
	for docIndex := 1; docIndex < len(allDoc); docIndex++ {
		docDatabase.AddDocumentBasic_MA(conn, allDoc[docIndex], allDoc[0].Title, string(graphNumber))
	}

	//sono i documenti ancora da esplorare ovvero i figli appena creati
	parentDocs := allDoc[1:]
	var childDocs []structures.MADocument
	for level := maxNumLv; level > 0 && parentDocs != nil; level-- {
		for _, doc := range parentDocs {
			select {
			case _, ok := <-quit:
				if ok {
					fmt.Println("Stopped at level: ", maxNumLv-level+1)
					close(quit)
					return true
				} else {
					fmt.Println("The channel has been closed unexpectedly")
					return false
				}
			default:
				//prima di esplorare un doc controllo se l'ho gia' esplorato
				if !docDatabase.AlreadyExplored(conn, doc.Title, string(graphNumber)) {
					childDocs = append(childDocs, getFirstsNDoc_MA(wd, doc, conn, threshold, perc, graphNumber)...)
				}
			}
		}
		parentDocs = childDocs
		childDocs = nil
		fmt.Println("parent: ", parentDocs)
		fmt.Println("child: ", childDocs)
		fmt.Println("Ho finito il livello: ", maxNumLv-level+1)
	}
	close(quit)
	return true
}

//Solo di supporto: le passo il doc di partenza e vado alla pagina di scholar con
//i documenti che lo citano e prendo quelli con numero citazioni > soglia, infine
//li aggiungo al database.
//"go run main_MA.go NUM_levels SOGLIA <URL_PRIMO_DOC>"
func getFirstsNDoc_MA(wd selenium.WebDriver, initialDoc structures.MADocument, conn bolt.Conn, threshold, perc, graphNumber int) []structures.MADocument {

	numPages := int((initialDoc.NumCitations / structures.NumArticlePerPageMA) + 1)
	citeInitialDoc, numFigli := webDriver.GetCiteDocumentsByThreshold_MA(wd, initialDoc.LinkCitations, numPages, threshold, perc)

	for docIndex := 0; docIndex < len(citeInitialDoc); docIndex++ {
		docDatabase.AddDocumentBasic_MA(conn, citeInitialDoc[docIndex], initialDoc.Title, string(graphNumber))
	}
	fmt.Println("Titolo: ", initialDoc.Title, " -- num figli: ", numFigli)
	return citeInitialDoc
}

//getUserInput gets the information about:
//startingPoint: Academic URl or list of words to search
//maxnumLv: the search result will be something similar to a tree, this value
//			is the max number tree levels
//threshold: the min citations number that a document must have to be considered
//perc: the citations number percentage that a document must have to be considered
//NOTE:
//condition to get an article (webDiver/info_documenti_MA.go/GetDocumentsFromPageBasic_MA())
// if numsCitations[i] < int64(threshold) || float32(numsCitations[i]) < float32(maxCit)*(float32(perc)/100)
func getUserInput() (string, int, int, int, int) {
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
	var perc int = 0
	for perc <= 0 || perc > 100 {
		fmt.Println("Finally the percentage on citations (e.g. 60):")
		str, _ := reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		perc, err = strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please insert a value between [1-100]")
			continue
		}
	}
	//choose which graph modify (previous research) or create a new one
	var graph int = 0
	if researchNumber == 1 {
		fmt.Println("There is no graph stored, a new one will be created")
		graph = researchNumber
	} else {
		for {
			fmt.Println("There are ", researchNumber-1, " graphs stored.")
			fmt.Println("Do you want to store the articles in a new graph?	")
			str, _ := reader.ReadString('\n')
			str = strings.Replace(str, "\n", "", -1)
			if str == "yes" {
				graph = researchNumber
				break
			}
			if str == "no" {
				for graph < 1 || graph >= researchNumber {
					fmt.Println("Choose one number between [1-", researchNumber-1, "]:	")
					str, _ := reader.ReadString('\n')
					str = strings.Replace(str, "\n", "", -1)
					graph, err = strconv.Atoi(str)
					if err != nil {
						fmt.Println("Please insert a number")
						continue
					}
				}
				break
			}
		}
	}

	return startingPoint, maxNumLv, threshold, perc, graph
}

func collectingArticles(wd selenium.WebDriver, startingPoint string, maxNumLv, threshold, perc, graph int) bool {
	//Start collecting articles and writing them on neo4j, it stops when the user writes "stop"
	var wg sync.WaitGroup
	wg.Add(2)
	//if the user writes "stop", the second goroutine stops the first through this channel
	quit := make(chan int8, 1)
	//if the function ends by itself, the first goroutine closes the channel to stop the second

	var t bool
	go func() {
		t = createCitationsTree_MA(wd, startingPoint, maxNumLv, threshold, perc, graph, quit)
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
	return true
}

/*
Il documento iniziale lo prendo da info_documenti_MA/GetInitialDocument_MA() facendo una ricerca nella home
di Microsoft Academic e prendendo il primo documento tra quelli restituiti.
Per cambiare il documento si puo' cambiare la parola da cercare nella funzione textBox.SendKeys() (riga 688)

Oppure se al main viene passato un terzo argomento, questo viene considerato l'URL della pagina di Academic del
primo documento.
*/
func main() {

	service, wd := webDriver.StartSelenium(-1)

	defer service.Stop()
	defer wd.Quit()

	//Show the functionality
	for {
		fmt.Println("Select a function:\n" +
			"0) Start searching\n" +
			"1) Print the fields ranking")
		str, _ := reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		choice, err := strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please insert an integer value")
			continue
		}
		switch choice {
		case 0:
			//Get user input
			startingPoint, maxNumLv, threshold, perc, graph := getUserInput()
			//print user's input
			logger.Println("Your input:\n")
			logger.Println("startingPoint = ", startingPoint)
			logger.Println("maxNumLv = ", maxNumLv)
			logger.Println("threshold = ", threshold)
			logger.Println("percentage on citations = ", perc)
			logger.Println("graph number = ", graph)

			t := collectingArticles(wd, startingPoint, maxNumLv, threshold, perc, graph)
			if t {
				fmt.Println("ALL OK")
			} else {
				fmt.Println("There could be some problem, check the file: Quali_livelli_ho_fatto")
			}
		case 1:
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
		default:
			fmt.Println("Please insert a correct value")
		}
	}

}

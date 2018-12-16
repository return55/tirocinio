package main

import (
	"bufio"
	"fmt"
	"go/build"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/return55/tirocinio/docDatabase"
	"github.com/return55/tirocinio/draw"
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
func createCitationsTree_MA(startingPoint string, maxNumLv, threshold, perc, graphNumber int, quit chan int8) bool {
	service, wd := webDriver.StartSelenium(-1)
	defer service.Stop()
	defer wd.Quit()

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
	//docDatabase.CleanAll(conn)
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
	docDatabase.AddDocumentBasic_MA(conn, allDoc[0], "", graphNumber)
	//if this condition is true, i have to create a new graph
	if graphNumber == researchNumber {
		researchNumber++
	}
	for docIndex := 1; docIndex < len(allDoc); docIndex++ {
		docDatabase.AddDocumentBasic_MA(conn, allDoc[docIndex], allDoc[0].URL, graphNumber)
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
				if !docDatabase.AlreadyExplored(conn, doc.URL, graphNumber) {
					childDocs = append(childDocs, getFirstsNDoc_MA(wd, doc, conn, threshold, perc, graphNumber)...)
				}
			}
		}
		parentDocs = childDocs
		childDocs = nil
		//fmt.Println("parent: ", parentDocs)
		//fmt.Println("child: ", childDocs)
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
		docDatabase.AddDocumentBasic_MA(conn, citeInitialDoc[docIndex], initialDoc.URL, graphNumber)
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
	var graphNumber int = 0
	if researchNumber == 1 {
		fmt.Println("There is no graph stored, a new one will be created")
		graphNumber = researchNumber
	} else {
		for {
			fmt.Println("There are ", researchNumber-1, " graphs stored.")
			fmt.Println("Do you want to store the articles in a new graph?	")
			str, _ := reader.ReadString('\n')
			str = strings.Replace(str, "\n", "", -1)
			if str == "yes" {
				graphNumber = researchNumber
				break
			}
			if str == "no" {
				for graphNumber < 1 || graphNumber >= researchNumber {
					fmt.Println("Choose one number between [1-", researchNumber-1, "]:	")
					str, _ := reader.ReadString('\n')
					str = strings.Replace(str, "\n", "", -1)
					graphNumber, err = strconv.Atoi(str)
					if err != nil {
						fmt.Println("Please insert a number")
						continue
					}
				}
				break
			}
		}
	}

	return startingPoint, maxNumLv, threshold, perc, graphNumber
}

func collectingArticles(startingPoint string, maxNumLv, threshold, perc, graphNumber int) bool {
	//Start collecting articles and writing them on neo4j, it stops when the user writes "stop"
	var wg sync.WaitGroup
	wg.Add(2)
	//if the user writes "stop", the second goroutine stops the first through this channel
	quit := make(chan int8, 1)
	//if the function ends by itself, the first goroutine closes the channel to stop the second

	var t bool
	go func() {
		t = createCitationsTree_MA(startingPoint, maxNumLv, threshold, perc, graphNumber, quit)
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

//printFieldsRanking print to stdOut the topN of the fields ordered by popularity (how many
//articles have this field of study) of one graph.
func printFieldsRanking() {
	//if there's no graph -> print an error message
	if researchNumber == 1 {
		fmt.Println("Sorry but the database is empty, first do a search")
		return
	}
	graphNumber := 0
	var err error
	for graphNumber <= 0 || graphNumber >= researchNumber {
		fmt.Println("What research do you mean?	[ 1 -", researchNumber-1, "]")
		str, _ := reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		graphNumber, err = strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please insert a number")
		}
	}
	numFields := 0
	for numFields <= 0 && numFields != -1 {
		fmt.Println("Insert the number of fields to show (-1 to see all fields):	")
		str, _ := reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		numFields, err = strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please insert a number")
		}
	}
	conn := docDatabase.StartNeo4j()
	defer conn.Close()
	_ = docDatabase.FieldsRanking(conn, numFields, graphNumber, true)
	/*fmt.Println("RANKING:\nSCORE\tFIELD")
	for field, score := range ranking {
		fmt.Println(strconv.Itoa(score) + "\t" + field)
	}*/
}

func cleanAll() {
	conn := docDatabase.StartNeo4j()
	defer conn.Close()
	docDatabase.CleanAll(conn)
}

func deleteGraph() {
	//if there's no graph -> print an error message
	if researchNumber == 1 {
		fmt.Println("Sorry but the database is empty, first do a search")
		return
	}
	graphNumber := 0
	var err error
	for graphNumber <= 0 || graphNumber >= researchNumber {
		fmt.Println("Which graph do you want to remove?	[ 1 -", researchNumber-1, "]")
		str, _ := reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		graphNumber, err = strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please insert a number")
		}
	}
	conn := docDatabase.StartNeo4j()
	defer conn.Close()
	if docDatabase.DeleteGraph(conn, graphNumber) {
		fmt.Println("Graph", graphNumber, "has been successfully deleted")
	} else {
		fmt.Println("No information has been deleted")
	}
}

//createDotFile gets the user inputs: file path
func createDotFile() {
	//if there's no graph -> print an error message
	if researchNumber == 1 {
		fmt.Println("Sorry but the database is empty, first do a search")
		return
	}
	fmt.Println("Insert the path of the new file:")
	filePath, _ := reader.ReadString('\n')
	filePath = strings.Replace(filePath, "\n", "", -1)
	graphNumber := 0
	var err error
	for graphNumber <= 0 || graphNumber >= researchNumber {
		fmt.Println("Which graph do you want to print?	[ 1 -", researchNumber-1, "]")
		str, _ := reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		graphNumber, err = strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please insert a number")
		}
	}
	draw.CreateFile(filePath, graphNumber)
}

//checkProject try to find the project directory from the
func checkProject(GOPATH string) bool {
	//if _, os.Stat(GOPATH+"/src/")
	return true
}

//dotEveryGraph creates a dot file for each graph in the database
func dotEveryGraph() bool { /*
		//get GOPATH to create the path to the project
		goPath := strings.Split(os.Getenv("GOPATH"), string(os.PathListSeparator))
		filePath := ""
		if len(goPath) == 0 || goPath[0] == "" {
			GOPATH := build.Default.GOPATH
			fmt.Println("No path (build), i try the deafult: ", GOPATH)
			if !checkProject(GOPATH) {
				fmt.Println("You don't have a GOPATH variable and i can't find the project directory strarting from", GOPATH)
				return
			}

		} else if len(goPath) == 1 {
			GOPATH := goPath[0]
			fmt.Println("One path:", GOPATH)
			if !checkProject(GOPATH) {
				fmt.Println("There's something wrong with your GOPATH variable () and i can't find the project directory")
				return
			}
		} else {
			//More than one path
			GOPATH := ""
			for _, path := range goPath {
				if checkProject(path) {
					GOPATH = path
					break
				}
			}
			//if no path is valid
			if GOPATH == "" {
				fmt.Println("You have more than one path in GOPATH variable but i can't find the project in any of them")
				return
			}
		}*/
	var str string
	for str != "yes" && str != "no" {
		fmt.Println("Do you want to use the default path ($GOPATH/src/github.com/return55/tirocinio/draw/fileDOT)?		")
		str, _ = reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
	}
	var directoryPath string
	if str == "yes" {
		GOPATH := build.Default.GOPATH
		directoryPath = GOPATH + "/src/github.com/return55/tirocinio/draw/fileDOT"
	} else {
		fmt.Println("Insert your directory's path:")
		directoryPath, _ = reader.ReadString('\n')
		directoryPath = strings.Replace(str, "\n", "", -1)
	}
	//little check
	if directoryPath == "" {
		panic("main/dotEveryGraph - something wrong with the path")
	}
	for i := 1; i < researchNumber; i++ {
		fmt.Println("path: " + directoryPath + "/" + strconv.Itoa(i) + ".dot")
		draw.CreateFile(directoryPath+"/"+strconv.Itoa(i)+".dot", i)
	}
	return true
}

//initialization queries the db to find the number of graphs (of search) and set
//the researchNumber's value
func initialization() {
	conn := docDatabase.StartNeo4j()
	defer conn.Close()
	researchNumber = docDatabase.GetResearchNumber(conn) + 1
}

/*
Il documento iniziale lo prendo da info_documenti_MA/GetInitialDocument_MA() facendo una ricerca nella home
di Microsoft Academic e prendendo il primo documento tra quelli restituiti.
Per cambiare il documento si puo' cambiare la parola da cercare nella funzione textBox.SendKeys() (riga 688)

Oppure se al main viene passato un terzo argomento, questo viene considerato l'URL della pagina di Academic del
primo documento.
*/
func main() {
	//Initialization: set researchNumber
	initialization()
	//if the user pass one command line argument (no matter the value), i start the script research
	if len(os.Args) > 1 {
		cleanAll()
		type data struct {
			startingPoint                          string
			maxNumLv, threshold, perc, graphNumber int
		}
		input := []data{
			{"https://academic.microsoft.com/#/detail/317187901", 10, 1000, 50, 1},
			{"https://academic.microsoft.com/#/detail/317187901", 10, 800, 50, 2},
			{"https://academic.microsoft.com/#/detail/317187901", 10, 600, 50, 3},
			{"https://academic.microsoft.com/#/detail/317187901", 10, 400, 50, 4},
			{"https://academic.microsoft.com/#/detail/317187901", 10, 150, 50, 5},
			{"https://academic.microsoft.com/#/detail/317187901", 10, 50, 50, 6},
		}
		for i, in := range input {
			quit := make(chan int8, 1)
			if createCitationsTree_MA(in.startingPoint, in.maxNumLv, in.threshold, in.perc, in.graphNumber, quit) {
				logger.Println("ALL GOOD with", i)
			} else {
				logger.Println("There is a problem with the iteration number", i)
				return
			}
		}
		return
	}
	//-------------------END SCRIPT------------------------------
	//Show the functionality
	for {
		fmt.Println("Select a function:\n" +
			"0) Start searching\n" +
			"1) Print the fields ranking\n" +
			"2) Clean db (delete all graphs)\n" +
			"3) Delete one search's results (one graph)\n" +
			"4) Print graph (create a new .dot file)\n" +
			"5) Print all graph available")
		str, _ := reader.ReadString('\n')
		str = strings.Replace(str, "\n", "", -1)
		choice, err := strconv.Atoi(str)
		if err != nil {
			fmt.Println("Please insert an integer value\n")
			continue
		}
		switch choice {
		case 0:
			//Get user input
			startingPoint, maxNumLv, threshold, perc, graphNumber := getUserInput()
			//print user's input
			logger.Println("Your input:\n")
			logger.Println("startingPoint = ", startingPoint)
			logger.Println("maxNumLv = ", maxNumLv)
			logger.Println("threshold = ", threshold)
			logger.Println("percentage on citations = ", perc)
			logger.Println("graph number = ", graphNumber)

			t := collectingArticles(startingPoint, maxNumLv, threshold, perc, graphNumber)
			if t {
				fmt.Println("ALL GOOD")
			} else {
				fmt.Println("There could be some problem, check the file: Quali_livelli_ho_fatto")
			}
		case 1:
			printFieldsRanking()
		case 2:
			cleanAll()
		case 3:
			deleteGraph()
		case 4:
			createDotFile()
		case 5:
			dotEveryGraph()
		default:
			fmt.Println("Please select one of the option above")
		}
		fmt.Println()
	}

}

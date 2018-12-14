package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/return55/tirocinio/docDatabase"
	"github.com/return55/tirocinio/structures"
)

func main3() {
	//Apro connessione neo4j
	conn := docDatabase.StartNeo4j()
	defer conn.Close()

	t := docDatabase.AlreadyExplored(conn, "Pyrolysis of Wood/Biomass for Bio-oil: A Critical Review", 1)

	fmt.Println(t)
}

func main5() { /*
		service, wd := webDriver.StartSelenium(-1)

		defer service.Stop()
		defer wd.Quit()
		initialDoc := webDriver.GetInitialDocument_MA(wd)

		fmt.Println("Link alle citazioni: ", initialDoc.LinkCitations)

		//citeInitialDoc, _ := webDriver.GetCiteDocuments_MA(wd, initialDoc.LinkCitations, 22)
		var citeInitialDoc []structures.MADocument
		allDoc := append(citeInitialDoc, initialDoc)
		allDoc[0], allDoc[len(allDoc)-1] = allDoc[len(allDoc)-1], allDoc[0]

		SaveDoc_MA(allDoc)
		webDriver.SaveDocuments(allDoc)
	*/

	/*stdinput := os.Stdin
	s, _ := stdinput.Stat()

	fmt.Println(s.Size())
	fmt.Println(reader.Buffered())
	time.Sleep(5 * time.Second)
	fmt.Println(s.Size())
	b, err := reader.Peek(1)
	fmt.Println("buff: ", b, "-", len(b), "-", err)
	str, _ := reader.ReadString('\n')
	fmt.Println(str)
	fmt.Println(reader.Buffered())
	fmt.Println(s.Size())
	return*/
	//Start collecting articles and writing them on neo4j, it stops when the user writes "stop"
	var wg sync.WaitGroup
	wg.Add(2)
	//if the user writes "stop", the second goroutine stops the first through this channel
	quit := make(chan int8, 1)
	//if the function ends by itself, the first goroutine closes the channel to stop the second
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(5 * time.Second)
			fmt.Println("sono ancora vivo ", i)
		}
		wg.Done()
	}()
	//scanner := bufio.NewScanner(os.Stdin)

	go func() {
		defer fmt.Println("finito")
		defer wg.Done()
		fmt.Println("Type \"stop\" to stop searching or wait until it ends:")
		stdin := make(chan string, 1)
		//this routine checks when the user insert some values
		go func() {
			fmt.Println("sono l'altro2")
			str, _ := reader.ReadString('\n')
			fmt.Println("cosa ho letto: -", str, "-")
			stdin <- strings.Replace(strings.Replace(str, "\n", "", -1), " ", "", -1)
		}()
		for {
			time.Sleep(5 * time.Second)
			//if the channel (quit) is closed, the first routine has already terminate
			select {
			case _, ok := <-quit:
				fmt.Println("sono l'altro")
				if !ok {
					fmt.Println("Channel closed correctly")
					return
				} else {
					fmt.Println("The channel is open")
				}
			case str, _ := <-stdin:
				fmt.Printf("-%s-", str)
				if str == "stop" {
					fmt.Println("-----------")
					quit <- 1
					fmt.Println("-----------")
					close(stdin)
					return
				} else {
					fmt.Println("Type \"stop\" to stop searching or wait until it ends:")
					go func() {
						fmt.Println("sono l'altro2")
						str, _ := reader.ReadString('\n')
						fmt.Println("cosa ho letto: -", str, "-")
						stdin <- strings.Replace(strings.Replace(str, "\n", "", -1), " ", "", -1)
					}()
					continue
				}
			}

		}
	}()
	/*
		//Pulisco il DB
		docDatabase.CleanAll(conn)

		//Aggiungo il primo doc
		docDatabase.AddDocument_MA(conn, allDoc[0], "")

		for i := 1; i < len(allDoc); i++ {
			fmt.Println("Titolo ", i, " : ", allDoc[i].Title)
			docDatabase.AddDocument_MA(conn, allDoc[i], allDoc[0].Title)
		}
	*/
	wg.Wait()
	return

}

func test(conn bolt.Conn) {
	_, _ = conn.ExecNeo("MATCH (n) RETURN n", nil)
}

//Salvo i documenti in formato Human Readable
func SaveDoc_MA(docs []structures.MADocument) {
	file, err := os.Create("Documenti.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, doc := range docs {
		file.WriteString("Titolo: " + doc.Title + "\n")
		file.WriteString("Url:\nPDF:\n")
		for _, pdf := range doc.Url.PDF {
			file.WriteString(pdf + "\n")
		}
		file.WriteString("WWW:\n")
		for _, www := range doc.Url.WWW {
			file.WriteString(www + "\n")
		}
		file.WriteString("Authors " + strconv.Itoa(len(doc.Authors)) + ":\n")
		for _, a := range doc.Authors {
			file.WriteString("Nome: " + a.Name + "\n")
			file.WriteString("Affiliazione: " + a.Affiliation + "\n")
		}
		file.WriteString("NumCitations: " + strconv.FormatInt(doc.NumCitations, 10) + "\n")
		file.WriteString("LinkCitations: " + doc.LinkCitations + "\n")
		file.WriteString("NumReferences: " + strconv.FormatInt(doc.NumReferences, 10) + "\n")
		file.WriteString("LinkReferences: " + doc.LinkReferences + "\n")
		file.WriteString("Abstract: " + doc.Abstract + "\n")
		file.WriteString("Date: " /*+ doc.Date*/ + "\n")
		file.WriteString("FieldsOfStudy " + strconv.Itoa(len(doc.FieldsOfStudy)) + ":" + "\n")
		for _, f := range doc.FieldsOfStudy {
			file.WriteString(f + "\n")
		}
	}
}

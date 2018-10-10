package main

import (
	"fmt"

	"github.com/return55/tirocinio/docDatabase"
	"github.com/return55/tirocinio/webDriver"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

func main() {
	service, wd := webDriver.StartSelenium(-1)

	defer service.Stop()
	defer wd.Quit()
	initialDoc := webDriver.GetInitialDocument_MA(wd)

	fmt.Println("Link alle citazioni: ",initialDoc.LinkCitations)
	
	citeInitialDoc, _ := webDriver.GetCiteDocuments_MA(wd, initialDoc.LinkCitations, 7)
	
	allDoc := append(citeInitialDoc, initialDoc)
	allDoc[0], allDoc[len(allDoc)-1] = allDoc[len(allDoc)-1], allDoc[0]
	
	//Apro connessione neo4j
	conn := docDatabase.StartNeo4j()
	defer conn.Close()
	//Aggiungo il prioìmo doc
	docDatabase.AddDocument_MA(conn, allDoc[0], "")
	
	for i:=1; i<len(allDoc); i++ {
		fmt.Println("Titolo ",i," : ", allDoc[i].Title)
		docDatabase.AddDocument_MA(conn, allDoc[i], allDoc[0].Title)
	}
	
	
	return

}

func test(conn bolt.Conn) {
	_, _ = conn.ExecNeo("MATCH (n) RETURN n", nil)
}

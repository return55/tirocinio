package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/return55/tirocinio/docDatabase"
	"github.com/return55/tirocinio/structures"
)

func main3() {
	//Apro connessione neo4j
	conn := docDatabase.StartNeo4j()
	defer conn.Close()

	t := docDatabase.AlreadyExplored(conn, "Pyrolysis of Wood/Biomass for Bio-oil: A Critical Review")

	fmt.Println(t)
}

func mainProva() { /*
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
	//Apro connessione neo4j
	conn := docDatabase.StartNeo4j()
	defer conn.Close()
	result, err := conn.ExecNeo("MATCH (doc:MADocumentBasic {title: 'Smashing the stack for fun and profit'}),"+
		" (field:MAFieldOfStudy {name: 'Generic'}) CREATE (doc)-[:"+strings.ToUpper("aoic")+"]->(field)",
		map[string]interface{}{})
	/*result, err := conn.ExecNeo("CREATE (doc:MAFieldOfStudy { name: 'Generic'})",
	map[string]interface{}{})*/
	if err != nil {
		panic(err)
	}
	numResult, _ := result.RowsAffected()
	fmt.Printf("Creato campo generico : %d\n", numResult)
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

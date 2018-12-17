package draw

import (
	"os"

	"github.com/return55/tirocinio/docDatabase"
)

//CreateFile creates a .dot file to show one or more graph with details
func CreateFile(filePath string, graphNumber int) {
	//creo un nuovo file (.dot) con le info sugli archi e i colori
	fOut, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	defer fOut.Close()
	defer fOut.Sync()

	//stringa di inzio file
	_, err = fOut.WriteString("digraph G {\n")
	if err != nil {
		panic(err)
	}

	conn := docDatabase.StartNeo4j()
	defer conn.Close()

	relations := docDatabase.GetGraphDocuments(conn, graphNumber)

	//scrivo i vari archi
	if relations != nil {
		for _, rel := range relations {
			fOut.WriteString("\t\t\"" + rel.SourceTitle + "\" -> \"" + rel.DestinationTitle + "\";\n")
		}
	}

	//stringa di fine file
	_, err = fOut.WriteString("}")
	if err != nil {
		panic(err)
	}

}

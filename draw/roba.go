package draw

import (
	"os"
	"strings"

	"github.com/return55/tirocinio/docDatabase"
)

//CreateFile creates a .dot file to show one or more graph with details
//If the fieldName != "" -> GetGraphDocuments return two more informations about the node:
//for each node: if it has fieldName between its fields (true) otherwise (false)
//thank to this information i can color the node
func CreateFile(filePath string, graphNumber int, fieldName, color string) {
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
			rel.SourceTitle = strings.Replace(rel.SourceTitle, "\"", "'", -1)
			rel.DestinationTitle = strings.Replace(rel.DestinationTitle, "\"", "'", -1)
			fOut.WriteString("\t\t\"" + rel.SourceTitle + "\" -> \"" + rel.DestinationTitle + "\"\n")
			if docDatabase.DoesDocumentHaveField(conn, rel.SourceTitle, fieldName, graphNumber) {
				fOut.WriteString("\"" + rel.SourceTitle + "\" [color = " + color + ", penwidth = 6.0];\n")
			}
			if docDatabase.DoesDocumentHaveField(conn, rel.DestinationTitle, fieldName, graphNumber) {
				fOut.WriteString("\"" + rel.DestinationTitle + "\" [color = " + color + ", penwidth = 6.0];\n")
			}
		}
	}

	//stringa di fine file
	_, err = fOut.WriteString("}")
	if err != nil {
		panic(err)
	}

}

//CreateFileFields creates a .dot file to show one or more graph with details
//shows the fields hierarchy.
//If the fieldName != "" -> GetGraphDocuments return two more informations about the node:
//for each node: if it has fieldName between its fields (true) otherwise (false)
//thank to this information i can color the node
func CreateFileFields(filePath string) {
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

	relations := docDatabase.GetGraphFields(conn)

	//scrivo i vari archi
	if relations != nil {
		for _, rel := range relations {
			rel.SourceTitle = strings.Replace(rel.SourceTitle, "\"", "'", -1)
			rel.DestinationTitle = strings.Replace(rel.DestinationTitle, "\"", "'", -1)
			fOut.WriteString("\t\t\"" + rel.SourceTitle + "\" -> \"" + rel.DestinationTitle + "\"\n")
		}
	}

	//stringa di fine file
	_, err = fOut.WriteString("}")
	if err != nil {
		panic(err)
	}

}

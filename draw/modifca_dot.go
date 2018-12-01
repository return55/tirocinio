package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

//struttura che contiene gli archi
type Graph struct {
	XMLName     xml.Name `xml:"graph"`
	Id          string   `xml:"id,attr"`
	Edgedefault string   `xml:"edgedefault,attr"`
	Edges       []Edge   `xml:"edge"`
}

type Edge struct {
	XMLName xml.Name `xml:"edge"`
	Id      string   `xml:"id,attr"`
	Source  string   `xml:"source,attr"`
	Target  string   `xml:"target,attr"`
	Label   string   `xml:"label,attr"`
	Datas   []Data   `xml:"data"`
}

type Data struct {
	XMLName xml.Name `xml:"data"`
	Key     string   `xml:"key,attr"`
}

//struttura con le info dell'arco che mi interessano
type Info struct {
	source string //source node
	target string //destination node
	field  string //field name (edge's label)
}

func main() {
	//apro il file
	xmlFile, err := os.Open("archi.xml")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened users.xml")
	defer xmlFile.Close()

	//leggo il contenuto
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// uso il tag del grafo ma sono presenti solo gli archi
	var allEdges Graph
	xml.Unmarshal(byteValue, &allEdges)

	// dizionario: idArcho -> info sull'arco
	dict := make(map[string]Info)
	for _, edge := range allEdges.Edges {
		dict[edge.Id] = Info{edge.Source, edge.Target, edge.Label}
	}

	//creo un nuovo file (.dot) con le info sugli archi e i colori
	fOut, err := os.Create("grafo_colorato.dot")
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

	//dizionario dei campi - colore
	colorMap := map[string]string{
		"CITE":                       "azure2",
		"COMPUTER_SCIENCE":           "blue",
		"COMPUTER_SECURITY":          "brown",
		"ANDROID__OPERATING_SYSTEM_": "coral",
		"REAL_TIME_COMPUTING":        "cyan",
	}
	//scrivo i vari archi
	for id, edge := range dict {
		fOut.WriteString("		" + edge.source + " -> " + edge.target + " [_graphml_id=" + id + ", color=" + colorMap[edge.field] + "];\n")
	}

	//stringa di fine file
	_, err = fOut.WriteString("}")
	if err != nil {
		panic(err)
	}

	//fmt.Println("numero archi: ", len(allEdges.Edges))

}

package draw

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

//struttura con le info dell'arco che mi interessano (id e' la chiave nella mappa)
type EdgeInfo struct {
	source string //source node
	target string //destination node
	field  string //field name (edge's label)
}

//abbinamento tra campo - colore
//dizionario dei campi - colore
var colorMap = map[string]string{
	"CITE":                       "azure2",
	"COMPUTER_SCIENCE":           "blue",
	"COMPUTER_SECURITY":          "brown",
	"ANDROID__OPERATING_SYSTEM_": "coral",
	"REAL_TIME_COMPUTING":        "cyan",
}

//Rimuove i nodi duplicati
func removeDuplicatesFromSlice(s []string) []string {
	m := make(map[string]bool)
	for _, item := range s {
		if _, ok := m[item]; ok {
			// duplicate item
			fmt.Println(item, "is a duplicate")
		} else {
			m[item] = true
		}
	}

	var result []string
	for item, _ := range m {
		result = append(result, item)
	}
	return result
}

//Scrivo un file in cui  gli archi sono colorati
func colorEdges(dict map[string]EdgeInfo) {
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

	//scrivo i vari archi
	for id, edge := range dict {
		fOut.WriteString("\t\t" + edge.source + " -> " + edge.target + " [_graphml_id=" + id + ", color=" + colorMap[edge.field] + "];\n")
	}

	//qui posso creare i vari cluster
	fOut.WriteString(oneCluster(dict, "COMPUTER_SCIENCE"))

	//stringa di fine file
	_, err = fOut.WriteString("}")
	if err != nil {
		panic(err)
	}
}

//Restituisce il codice per il cluster dei nodi di uno specifico campo
//Cerco i nodi che appartengono a un campo, li metto in una lista (credo non sia necessario controllare le ripetizioni)
//creo il codice
func oneCluster(dict map[string]EdgeInfo, fieldName string) string {
	//controllo che il nome sia valido
	t := false
	for field, _ := range colorMap {
		if field == fieldName {
			t = true
			break
		}
	}
	if !t {
		fmt.Println("Il nome del campo " + fieldName + " non e' disponibile.")
		return ""
	}
	//cerco i nodi che appartengono a fieldName
	var nodes []string
	for _, info := range dict {
		if info.field == fieldName {
			nodes = append(nodes, info.source, info.target)
		}
	}
	//rimuovo i duplicati
	nodes = removeDuplicatesFromSlice(nodes)
	//creo la stringa che contiene la lista dei nodi: "nodo1" "nodo2" ...
	code := "\""
	for _, node := range nodes {
		code += "\"" + node + "\" "
	}
	code += "\";"
	//creo il codice
	return "\t\tsubgraph " + fieldName + " {\n" +
		"\t\t\tnode [style=filled];\n" +
		"\t\t\t" + code + "\n" +
		"\t\t\tlabel = \"" + fieldName + "\";\n" +
		"\t\t\tcolor=" + colorMap[fieldName] + ";\n" +
		"\t\t}\n"

}

func main99() {
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
	dict := make(map[string]EdgeInfo)
	for _, edge := range allEdges.Edges {
		dict[edge.Id] = EdgeInfo{edge.Source, edge.Target, edge.Label}
	}

	//coloro gli archi e creo i cluster dei campi
	colorEdges(dict)

	//fmt.Println("numero archi: ", len(allEdges.Edges))

}

package docDatabase

import (
	"fmt"
	"os/exec"
	"reflect"
	"regexp"

	"github.com/return55/tirocinio/structures"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

//Aggiungo il documento, gli autori che lo hanno scritto (con le relative relazioni)
//urlStartDoc e' l'URL del documento che viene citato da doc:
//se e' "" allora doc e' il documento da cui e' partita l'esplorazione
func AddDocument(conn bolt.Conn, document structures.Document, urlStartDoc string) {
	//mappa di un singolo documento
	fieldsMap := make(map[string]interface{})
	//devo usare la reflection per accedere ai campi di Document
	//tramite delle stringhe (i nomi dei campi)
	doc := reflect.ValueOf(document)

	for _, field := range structures.FieldsName {
		if field != "Authors" {
			fieldsMap[field] = doc.FieldByName(field).Interface()
		}
	}
	//aggiungo il documento
	result, err := conn.ExecNeo("CREATE (doc:Document {url: {Url},"+
		" numCitedBy: {NumCitedBy}, linkCitedBy: {LinkCitedBy}})", fieldsMap)
	if err != nil {
		panic(err)
	}
	//aggiungo la relazione tra document e il documento che cita
	if urlStartDoc != "" {
		_, err := conn.ExecNeo("MATCH (newDoc:Document {url: {NewUrl}}), (citedDoc:Document {url: {UrlStartDoc}})"+
			"CREATE (newDoc)-[:CITE]->(citedDoc)",
			map[string]interface{}{"NewUrl": document.Url, "UrlStartDoc": urlStartDoc})
		if err != nil {
			panic(err)
		}
	}
	numResult, _ := result.RowsAffected()
	fmt.Printf("CREATED DOCUMENT: %d\n", numResult) // CREATED ROWS: 1

	//creo gli autori e aggiungo le relazioni tra autori e il documento attuale
	for _, author := range document.Authors {
		//faccio una query per controllare se l'autore e' presente nel db,
		//se non e' presente lo aggiungo
		result, err := conn.ExecNeo("MERGE (author:Author {name : {Name}})",
			map[string]interface{}{"Name": author})
		if err != nil {
			panic(err)
		}
		numResult, _ := result.RowsAffected()
		fmt.Printf("CREATED AUTHOR: %d\n", numResult) // CREATED ROWS: 1

		//aggiungo la relazione: documento -[scritto_da]-> autore
		result, err = conn.ExecNeo("MATCH (doc:Document {url : {Url}}), (author:Author {name : {Name}})"+
			"CREATE (doc)-[:WRITTEN_BY]->(author)",
			map[string]interface{}{"Url": document.Url, "Name": author})
		if err != nil {
			panic(err)
		}
		numResult, _ = result.RowsAffected()
		fmt.Printf("CREATED REALATION: %d\n", numResult) // CREATED ROWS: 1
	}
}

//Crea una nuova connessione con Neo4j
func StartNeo4j() bolt.Conn {
	driver := bolt.NewDriver()
	conn, err := driver.OpenNeo("bolt://127.0.0.1:7687")
	if err != nil {
		panic(err)
	}
	return conn
}

//Crea un pool di connessioni con Neo4j
//come e' diviso pool:
//array di connesioni : pos=0 (connessione di Concurrency)
// pos = [1,numThread] (connessioni per i Thread)
func StartPoolNeo4j(connectionNumber int) []bolt.Conn {
	driver, err := bolt.NewClosableDriverPool("bolt://127.0.0.1:7687", connectionNumber)
	if err != nil {
		panic(err)
	}
	//creo l'array delle connessioni
	pool := make([]bolt.Conn, connectionNumber)
	for i := 0; i < connectionNumber; i++ {
		conn, err := driver.OpenPool()
		if err != nil {
			panic(err)
		}
		pool[i] = conn
	}

	return pool
}

//Pulisce il database da tutti i nodi e le relazioni
func CleanAll(conn bolt.Conn) {
	_, err := conn.ExecNeo("MATCH (n), ()-[r]-() DELETE n,r", nil)
	if err != nil {
		//se l'errore e' dovuto alla mancanza della memoria heap (il db e' troppo grosso)
		//NON FUNZIONANO PIU' PERCHE' USO ENTERPRISE, NON COMMUNITY
		if t, _ := regexp.MatchString(".*OutOfMemoryError.*", err.Error()); t {
			if err := exec.Command("rm", "-fr", "docDatabase/neo4j-community-3.3.5/data/databases/graph.db").Run(); err != nil {
				panic(err)
			}
			if err := exec.Command("mkdir", "docDatabase/neo4j-community-3.3.5/data/databases/graph.db").Run(); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	fmt.Println("All clean!!")
}

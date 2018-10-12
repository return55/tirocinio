package docDatabase

import (
	"fmt"
	"reflect"

	"github.com/return55/tirocinio/structures"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

//Aggiungo il documento, gli autori che lo hanno scritto (con le relative relazioni) e
//tutte le informazioni dei doc MA (eccetto la data che non riesco a prendere)
//urlStartDoc e' l'URL del documento che viene citato da doc:
//se e' "" allora doc e' il documento da cui e' partita l'esplorazione
//NOTA:
//Per ora Source lo spezzo in 2 proprieta': sourceWWW, sourcePDF;
//poi dovro' creare un nodo a parte con una relazione che lo lega al documento.
func AddDocument_MA(conn bolt.Conn, document structures.MADocument, titleStartDoc string) {
	//mappa di un singolo documento
	fieldsMap := make(map[string]interface{})
	//devo usare la reflection per accedere ai campi di Document
	//tramite delle stringhe (i nomi dei campi)
	doc := reflect.ValueOf(document)

	for _, field := range structures.FieldsNameMA {
		if field != "Authors" && field != "Url" {
			//pe capire qualce campo e' vuoto
			fmt.Println("Campo: ", field)
			fieldsMap[field] = doc.FieldByName(field).Interface()
		}
	}
	//aggiungo il documento
	result, err := conn.ExecNeo("CREATE (doc:MADocument {title: {Title},"+
		/*" sourceWWW: {Url.WWW}, sourcePDF: {Url.PDF},*/ " numCitations: {NumCitations},"+
		" linkCitations: {LinkCitations}, numReferences: {NumReferences}, abstract: {Abstract},"+
		" date: {Date}, fieldsOfStudy: {FieldsOfStudy}})", fieldsMap)
	if err != nil {
		panic(err)
	}
	//aggiungo la relazione tra document e il documento che cita
	if titleStartDoc != "" {
		_, err := conn.ExecNeo("MATCH (newDoc:MADocument {title: {Title}}), (citedDoc:MADocument {title: {TitleStartDoc}})"+
			"CREATE (newDoc)-[:CITE]->(citedDoc)",
			map[string]interface{}{"Title": document.Title, "TitleStartDoc": titleStartDoc})
		if err != nil {
			panic(err)
		}
	}
	numResult, _ := result.RowsAffected()
	fmt.Printf("CREATED DOCUMENT: %d\n", numResult) // CREATED ROWS: 1

	//creo gli autori e aggiungo le relazioni tra autori e il documento attuale
	//NOTA:
	//L'affiliazione e' una proprieta' della relazione CITE.
	for _, author := range document.Authors {
		//faccio una query per controllare se l'autore e' presente nel db,
		//se non e' presente lo aggiungo
		result, err := conn.ExecNeo("MERGE (author:MAAuthor {name : {Name}})",
			map[string]interface{}{"Name": author.Name})
		if err != nil {
			panic(err)
		}
		numResult, _ := result.RowsAffected()
		fmt.Printf("CREATED AUTHOR: %d\n", numResult) // CREATED ROWS: 1

		//aggiungo la relazione: documento -[scritto_da]-> autore
		result, err = conn.ExecNeo("MATCH (doc:MADocument {title: {Title}}), (author:MAAuthor {name : {Name}})"+
			"CREATE (doc)-[:MA_WRITTEN_BY {affiliation: {Affiliation}}]->(author)",
			map[string]interface{}{"Title": document.Title, "Name": author.Name, "Affiliation": author.Affiliation})
		if err != nil {
			panic(err)
		}
		numResult, _ = result.RowsAffected()
		fmt.Printf("CREATED REALATION: %d\n", numResult) // CREATED ROWS: 1
	}
}

//Crea un nodo che ha solo: titolo, numCitazioni e linkCitazioni.
//Niente autori, ne sources
func AddDocumentBasic_MA(conn bolt.Conn, document structures.MADocument, titleStartDoc string) {
	//mappa di un singolo documento
	fieldsMap := make(map[string]interface{})
	//devo usare la reflection per accedere ai campi di Document
	//tramite delle stringhe (i nomi dei campi)
	doc := reflect.ValueOf(document)

	fieldsMap["Titolo"] = doc.FieldByName("Titolo").Interface()
	fieldsMap["NumCitations"] = doc.FieldByName("NumCitations").Interface()
	fieldsMap["LinkCitations"] = doc.FieldByName("LinkCitations").Interface()

	//aggiungo il documento
	result, err := conn.ExecNeo("CREATE (doc:MADocumentBasic {title: {Title},"+
		" numCitations: {NumCitations}, linkCitations: {LinkCitations}", fieldsMap)
	if err != nil {
		panic(err)
	}
	//aggiungo la relazione tra document e il documento che cita
	if titleStartDoc != "" {
		_, err := conn.ExecNeo("MATCH (newDoc:MADocumentBasic {title: {Title}}), (citedDoc:MADocumentBasic {title: {TitleStartDoc}})"+
			"CREATE (newDoc)-[:CITE]->(citedDoc)",
			map[string]interface{}{"Title": document.Title, "TitleStartDoc": titleStartDoc})
		if err != nil {
			panic(err)
		}
	}
	numResult, _ := result.RowsAffected()
	fmt.Printf("CREATED DOCUMENT: %d\n", numResult) // CREATED ROWS: 1

}

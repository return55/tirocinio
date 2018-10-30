package docDatabase

import (
	"fmt"
	"io"
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

	fieldsMap["Title"] = doc.FieldByName("Title").Interface()
	fieldsMap["NumCitations"] = doc.FieldByName("NumCitations").Interface()
	fieldsMap["LinkCitations"] = doc.FieldByName("LinkCitations").Interface()
	fieldsMap["FieldsOfStudy"] = doc.FieldByName("FieldsOfStudy").Interface()

	//aggiungo il documento se non e' presente
	result, err := conn.ExecNeo("MERGE (doc:MADocumentBasic {title: {Title}, "+
		"numCitations: {NumCitations}, linkCitations: {LinkCitations}, "+
		"fieldsOfStudy: {FieldsOfStudy}})", fieldsMap)
	if err != nil {
		panic(err)
	}
	numResult, _ := result.RowsAffected()
	fmt.Printf("CREATED DOCUMENT: %d\n", numResult)

	//aggiungo la relazione tra document e il documento che cita
	//e dico che non e' la radice (isRoot = false)
	if titleStartDoc != "" {
		_, err := conn.ExecNeo("MATCH (newDoc:MADocumentBasic {title: {Title}}), (citedDoc:MADocumentBasic {title: {TitleStartDoc}}) "+
			"CREATE (newDoc)-[:CITE]->(citedDoc) SET newDoc.isRoot = false",
			map[string]interface{}{"Title": document.Title, "TitleStartDoc": titleStartDoc})
		if err != nil {
			panic(err)
		}
	} else {
		//dico che il doc iniziale e' la radice dell'albero
		_, err := conn.ExecNeo("MATCH (initialDoc:MADocumentBasic {title: {Title}}) SET initialDoc.isRoot = true",
			map[string]interface{}{"Title": document.Title})
		if err != nil {
			panic(err)
		}
	}

}

//Controllo se il documento e' gia stato esplorato:
//ha gia' dei doc che lo citano.
//NOTA:
//Non e' proprio vero perche' un doc potrebbe non avere figli che
//soddisfano la soglia minima, ma a me va benem cosi'
func AlreadyExplored(conn bolt.Conn, title string) bool {
	rows, err := conn.QueryNeo("MATCH (doc:MADocumentBasic {title: {Title}}), (otherDoc:MADocumentBasic) "+
		"WHERE (otherDoc)-[:CITE]->(doc) "+
		"RETURN COUNT(otherDoc)", map[string]interface{}{"Title": title})
	if err != nil {
		panic(err)
	}

	numDocInterface, _, err := rows.NextNeo()
	_ = rows.Close()
	if err != nil {
		if err == io.EOF {
			return false
		}
		panic(err)
	}

	return reflect.ValueOf(numDocInterface[0]).Int() > 0

}

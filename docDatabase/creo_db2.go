package docDatabase

import (
	"fmt"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	//"github.com/johnnadratowski/golang-neo4j-bolt-driver/structures/graph"
	"github.com/tirocinio/structures"
	"reflect"
	"io"
	"os"
)

func CreoDB2(allDoc []structures.Document) {
	driver := bolt.NewDriver()
	conn, err := driver.OpenNeo("bolt://127.0.0.1:7687")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Here we prepare a new statement. This gives us the flexibility to
	// cancel that statement without any request sent to Neo
	// stmt, err := conn.PrepareNeo("CREATE (doc:Document {url: {Url}, authors: {Authors}," +
	// 	"numCitedBy: {NumCitedBy}, linkCitedBy: {LinkCitedBy}})")
	// if err != nil {
	// 	panic(err)
	// }

	//Aggiungo i documenti al database:
	//Documento: url, numCitedBy, linkCitedBy
	//Autore: string(nome cognome alla buona)
	//Relazioni:
	//doc -[cita]-> doc | doc -[scritto_da]-> aut | aut-[cita]->aut
	for docIndex := 0; docIndex < len(allDoc); docIndex++ {
		//mappa di un singolo documento
		fieldsMap := make(map[string]interface{})
		//devo usare la reflection per accedere ai campi di Document
		//tramite delle stringhe (i nomi dei campi)
		doc := reflect.ValueOf(allDoc[docIndex])

		for _, field := range structures.FieldsName {
			if field != "Authors" {
				fieldsMap[field] = doc.FieldByName(field).Interface()
			}
		}
		result, err := conn.ExecNeo("CREATE (doc:Document {url: {Url},"+
			" numCitedBy: {NumCitedBy}, linkCitedBy: {LinkCitedBy}})", fieldsMap)
			
		if err != nil {
			panic(err)
		}
		numResult, _ := result.RowsAffected()
		fmt.Printf("CREATED ROWS: %d\n", numResult) // CREATED ROWS: 1
		
		//creo gli autori e aggiungo le relazioni tra autori e documenti
		for _, author := range allDoc[docIndex].Authors {
			//faccio una query per controllare se l'autore e' presente nel db
			row, err := conn.QueryNeo("MATCH (author:Author) WHERE author.name = {Name} RETURN author.name", 
							map[string]interface{}{"Name" : author})
			if err != nil{
				panic(err)
			}
			_, _, err := row.NextNeo()
			//se l'autore non e' presente nel db lo aggiungo
			if err == io.EOF{
				_, errCreate := conn.ExecNeo("CREATE (author:Author {name: {Name}})", map[string]interface{}{"Name" : author})
				if errCreate != nil{
					panic(errCreate)
				}
			}else if err != nil{
					panic(err)
			}
			//aggiungo la relazione: documento -[scritto_da]-> autore
			result, err := conn.ExecNeo("CREATE (doc:Document {url: {Url},"+
				" numCitedBy: {NumCitedBy}, linkCitedBy: {LinkCitedBy}})", fieldsMap)
			
			if err != nil {
				panic(err)
			}
			numResult, _ := result.RowsAffected()
			fmt.Printf("CREATED ROWS: %d\n", numResult) // CREATED ROWS: 1
		
		}		

	}

	// result, err := stmt.ExecNeo(mapDocs)
	// //*****result, err := stmt.ExecNeo(map[string]interface{}{"foo": 1, "bar": 2.2})
	// if err != nil {
	// 	panic(err)
	// }

	/*	// Lets get the node
		data, rowsMetadata, _, _ := conn.QueryNeoAll("MATCH (n:NODE) RETURN n.foo, n.bar", nil)
		fmt.Printf("COLUMNS: %#v\n", rowsMetadata["fields"].([]interface{}))    // COLUMNS: n.foo,n.bar
		fmt.Printf("FIELDS: %d %f\n", data[0][0].(int64), data[0][1].(float64)) // FIELDS: 1 2.2

		// oh cool, that worked. lets blast this baby and tell it to run a bunch of statements
		// in neo concurrently with a pipeline
		results, _ := conn.ExecPipeline([]string{
			"MATCH (n:NODE) CREATE (n)-[:REL]->(f:FOO)",
			"MATCH (n:NODE) CREATE (n)-[:REL]->(b:BAR)",
			"MATCH (n:NODE) CREATE (n)-[:REL]->(z:BAZ)",
			"MATCH (n:NODE) CREATE (n)-[:REL]->(f:FOO)",
			"MATCH (n:NODE) CREATE (n)-[:REL]->(b:BAR)",
			"MATCH (n:NODE) CREATE (n)-[:REL]->(z:BAZ)",
		}, nil, nil, nil, nil, nil, nil)
		for _, result := range results {
			numResult, _ := result.RowsAffected()
			fmt.Printf("CREATED ROWS: %d\n", numResult) // CREATED ROWS: 2 (per each iteration)
		}

		data, _, _, _ = conn.QueryNeoAll("MATCH (n:NODE)-[:REL]->(m) RETURN m", nil)
		for _, row := range data {
			fmt.Printf("NODE: %#v\n", row[0].(graph.Node)) // Prints all nodes
		}

		result, _ = conn.ExecNeo(`MATCH (n) DETACH DELETE n`, nil)
		numResult, _ = result.RowsAffected()
		fmt.Printf("Rows Deleted: %d", numResult) // Rows Deleted: 13*/
}

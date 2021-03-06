package docDatabase

import (
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/return55/tirocinio/structures"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

//FieldsRanking queries the db to retrieve the top <numFields> field.
//If <numFields> == -1 -> there's no limit
//The field score is the number of edges from one article to another whose
//property "name" is equal to the field.
//If print -> print to stdOut the ranking
func FieldsRanking(conn bolt.Conn, numFields, graphNumber int, print bool) map[string]int {
	var graphNumberInterface interface{} = graphNumber
	query := "match (f:MAFieldOfStudy)<-[r:HAS_FIELD]-(d:MADocumentBasic {searchId: {GraphNumber}})" +
		"return f.name as name, count(distinct(d)) as score " +
		"order by score desc "
	if numFields >= 0 {
		query += "limit " + strconv.FormatInt(int64(numFields), 10)
	}
	rows, err := conn.QueryNeo(query, map[string]interface{}{"GraphNumber": graphNumberInterface})
	if err != nil {
		panic(err)
	}

	ranking := make(map[string]int)
	if print {
		fmt.Println("RANKING:\nSCORE\tFIELD")
	}
	for row, _, err := rows.NextNeo(); err != io.EOF; row, _, err = rows.NextNeo() {
		ranking[row[0].(string)] = (int)(row[1].(int64))
		if print {
			fmt.Println(strconv.Itoa((int)(row[1].(int64))) + "\t" + row[0].(string))
		}

	}
	_ = rows.Close()
	return ranking
}

//DeleteGraph remove from the database all documents and relations relative to a specific research
func DeleteGraph(conn bolt.Conn, graphNumber int) bool {
	var graphNumberInterface interface{} = graphNumber
	result, err := conn.ExecNeo("match (d:MADocumentBasic {searchId: {GraphNumber}})-[r]->() DELETE d,r",
		map[string]interface{}{"GraphNumber": graphNumberInterface})
	if err != nil {
		panic(err)
	}
	numResult, _ := result.RowsAffected()
	if numResult < 0 {
		panic("docDatabase/DeleteGraph : returns a negative value")
	} else if numResult == 0 {
		return false
	} else {
		return true
	}
}

//GetResearchNumber returns the number of graphs (of searches) in the db
func GetResearchNumber(conn bolt.Conn) int {
	rows, err := conn.QueryNeo("MATCH (n:MADocumentBasic) RETURN MAX(n.searchId)",
		map[string]interface{}{})
	if err != nil {
		panic(err)
	}

	numDocInterface, _, err := rows.NextNeo()
	_ = rows.Close()
	if err != nil {
		panic(err)
	}
	if numDocInterface[0] == nil {
		fmt.Println("There are no articles")
		return 0
	} else {
		fmt.Println("The db contains ", reflect.ValueOf(numDocInterface[0]).Int(), " searches")
		return int(reflect.ValueOf(numDocInterface[0]).Int())
	}

}

//GetGraphDocuments get the info aboute "CITE" relation:
//		(title1)-[CITE]->(title2)
//for all the documents of a graph (or all db if graphNumber == -1)
func GetGraphDocuments(conn bolt.Conn, graphNumber int) []structures.CiteRelation {
	var rows bolt.Rows
	var err error
	//look for only one graph
	if graphNumber > 0 {
		var graphNumberInterface interface{} = graphNumber
		rows, err = conn.QueryNeo("match (s:MADocumentBasic {searchId: {GraphNumber}})-[:CITE]->(d:MADocumentBasic {searchId: {GraphNumber}})"+
			"return s.title, d.title", map[string]interface{}{"GraphNumber": graphNumberInterface})
		if err != nil {
			panic(err)
		}
		//look for all the db
	} else if graphNumber == -1 {
		rows, err = conn.QueryNeo("match (s:MADocumentBasic)-[:CITE]->(d:MADocumentBasic)"+
			"return s.title, d.title", map[string]interface{}{})
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("The number of graph can only be positive or -1 to consider all db")
		return nil
	}

	var relations []structures.CiteRelation
	for row, _, err := rows.NextNeo(); err != io.EOF; row, _, err = rows.NextNeo() {
		relations = append(relations, structures.CiteRelation{
			SourceTitle:      row[0].(string),
			DestinationTitle: row[1].(string),
		})
	}
	_ = rows.Close()
	return relations

}

//GetGraphFields get the info aboute "CITE2" relation:
//		(field1)-[CITE2]->(field2)
//for all the fields of all db (for now) ///////of a graph (or all db if graphNumber == -1)
func GetGraphFields(conn bolt.Conn) []structures.CiteRelation {
	var rows bolt.Rows
	var err error

	rows, err = conn.QueryNeo("MATCH (f:MAFieldOfStudy2)-[r]->(f2:MAFieldOfStudy2) RETURN f.name, f2.name",
		map[string]interface{}{})
	if err != nil {
		panic(err)
	}

	var relations []structures.CiteRelation
	for row, _, err := rows.NextNeo(); err != io.EOF; row, _, err = rows.NextNeo() {
		relations = append(relations, structures.CiteRelation{
			SourceTitle:      row[0].(string),
			DestinationTitle: row[1].(string),
		})
	}
	_ = rows.Close()
	return relations

}

//DoesDocumentHaveField returns true if the document (title) has the field of
//study specified, false otherwise.
//NOTA: dovrei usare l'URL del documento non il suo titolo
func DoesDocumentHaveField(conn bolt.Conn, title, fieldName string, graphNumber int) bool {
	var rows bolt.Rows
	var err error
	//look for only one graph
	if graphNumber > 0 {
		var graphNumberInterface interface{} = graphNumber
		rows, err = conn.QueryNeo(
			"MATCH (f:MAFieldOfStudy {name: {FieldName}})<-[r:HAS_FIELD]-(d:MADocumentBasic {title: {Title}, searchId: {GraphNumber}})"+
				"RETURN COUNT(d)",
			map[string]interface{}{"Title": title, "GraphNumber": graphNumberInterface, "FieldName": fieldName})
		if err != nil {
			panic(err)
		}
		//look for all the db
	} else if graphNumber == -1 {
		rows, err = conn.QueryNeo(
			"MATCH (f:MAFieldOfStudy {name: {FieldName}})<-[r:HAS_FIELD]-(d:MADocumentBasic {title: {Title}})"+
				"RETURN COUNT(d)",
			map[string]interface{}{"Title": title, "FieldName": fieldName})
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("The number of graph can only be positive or -1 to consider all db")
		return false
	}

	hasField, _, err := rows.NextNeo()
	_ = rows.Close()
	if err != nil {
		if err == io.EOF {
			return false
		}
		panic(err)
	}

	return reflect.ValueOf(hasField[0]).Int() > 0
}

//ONLY FOR GRAPH NUMBER 5
//1) Create relations between articles that have the same field
//2) Create a new node for each field
//3) Create relations between these nodes
func PrepareDBForDotFileFields() {
	conn := StartNeo4j()
	defer conn.Close()

	CreateDocumentsRelations(conn)
	CreateNewFields(conn)
	CreateFieldsRelations(conn)
}

//------------------THESE FUNCTIONS ARE ONLY FOR PrepareDBForDotFileFields()--------------------

//Creates a relation "CITE_FIELD" from two documents that have the same field and are linked by the
//relation "CITE", this relation has the field's name as property.
func CreateDocumentsRelations(conn bolt.Conn) {
	_, err := conn.QueryNeo(
		"match (sorgente:MADocumentBasic {searchId: 5})-[s]->(f:MAFieldOfStudy)<-[d]-(destinazione:MADocumentBasic {searchId: 5})"+
			"where (sorgente)-[:CITE]->(destinazione)"+
			"and sorgente <> destinazione"+
			"merge (sorgente)-[r:CITE_FIELD {field: f.name}]->(destinazione)"+
			"return sorgente.title as sorg, r.field, destinazione.title as dest",
		map[string]interface{}{})
	if err != nil {
		panic(err)
	}
}

//Creates a new node (label:"MAFieldOfStudy2") for each field "MAFieldOfStudy"
func CreateNewFields(conn bolt.Conn) {
	_, err := conn.QueryNeo("match (f:MAFieldOfStudy)"+
		"merge (f2:MAFieldOfStudy2 {name: f.name})"+
		"return count(f) as num_campi_originali, count(f2) as num_campi_nuovi",
		map[string]interface{}{})
	if err != nil {
		panic(err)
	}
}

//Creates a relation "CITE2" between two "MAFieldOfStudy2" whose names appear
//consecutively in two "CITE_FIELD"
func CreateFieldsRelations(conn bolt.Conn) {
	_, err := conn.QueryNeo(
		"match (sorgente:MADocumentBasic {searchId: 5})-[f1:CITE_FIELD]->(meta:MADocumentBasic {searchId: 5})-[f2:CITE_FIELD]->(destinazione:MADocumentBasic {searchId: 5}), (field1:MAFieldOfStudy2 ), (field2:MAFieldOfStudy2)"+
			"where f1.field <> f2.field and sorgente <> meta and meta <> destinazione and sorgente <> destinazione"+
			"and field1.name = f1.field and field2.name = f2.field"+
			"merge (field1)-[r:CITE2]->(field2)"+
			"return count(r) as nuove_relazioni",
		map[string]interface{}{})
	if err != nil {
		panic(err)
	}
}

//-----------------------------------------------------------------------------------------------

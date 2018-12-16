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

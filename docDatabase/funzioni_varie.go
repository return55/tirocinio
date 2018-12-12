package docDatabase

import (
	"io"
	"strconv"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

//FieldsRanking queries the db to retrieve the top <numFields> field.
//If <numFields> == -1 -> there's no limit
//The field score is the number of edges from one article to another whose
//property "name" is equal to the field.
func FieldsRanking(conn bolt.Conn, numFields int) map[string]int {
	query := "match (f:MAFieldOfStudy)<-[r]-(d:MADocumentBasic) " +
		"return type(r) as name, count(distinct(d)) as score " +
		"order by score desc "
	if numFields >= 0 {
		query += "limit " + strconv.FormatInt(int64(numFields), 10)
	}
	rows, err := conn.QueryNeo(query, map[string]interface{}{})
	if err != nil {
		panic(err)
	}

	var ranking map[string]int
	for row, _, err := rows.NextNeo(); err != io.EOF; row, _, err = rows.NextNeo() {
		ranking[row[0].(string)] = (int) (row[1].(int64))
	}
	return ranking
}

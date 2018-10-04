package main

import (
	"fmt"

	"github.com/return55/tirocinio/docDatabase"
	"github.com/return55/tirocinio/webDriver"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
)

func main() {
	service, wd := webDriver.StartSelenium(-1)

	defer service.Stop()
	defer wd.Quit()
	_ = webDriver.GetInitialDocument_MA(wd)

	//fmt.Println(docs)
	return

	_, _ = webDriver.StartSelenium(-1)

	fmt.Println("111111111111111111111111111111111111111111")
	allDoc := webDriver.LoadDocuments(47)
	webDriver.PrintDocuments(allDoc)

	fmt.Println("222222222222222222222222222222222222222222")
	pool := docDatabase.StartPoolNeo4j(6)
	for i, conn := range pool {
		docDatabase.AddDocument(conn, allDoc[i], "")
	}

}

func test(conn bolt.Conn) {
	_, _ = conn.ExecNeo("MATCH (n) RETURN n", nil)
}

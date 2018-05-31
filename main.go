package main

import (
	//"fmt"
	"github.com/tirocinio/docDatabase"
	"github.com/tirocinio/structures"
	"github.com/tirocinio/webDriver"
)

func main() {
	var initialDoc structures.Document
	var citeInitialDoc []structures.Document

	service, wd := webDriver.StartSelenium()

	defer service.Stop()
	defer wd.Quit()

	initialDoc = webDriver.GetInitialDocument(service, wd, structures.URLScholar)

	citeInitialDoc = webDriver.GetCiteDocument(service, wd, initialDoc)

	allDoc := append(citeInitialDoc, initialDoc)
	allDoc[0], allDoc[10] = allDoc[10], allDoc[0]
	
	/*webDriver.SaveDocuments(nil)
	webDriver.LoadDocuments(allDoc)
	webDriver.PrintDocuments(allDoc)*/
	

	docDatabase.CreoDB2(allDoc)

}

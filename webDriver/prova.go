package webDriver
/*
 import (
 	"fmt"
 	"strings"
 	"github.com/tirocinio/structures"
 )

 func main(){

 	//fmt.Println(strings.Replace("H Balakrishnan, VN Padmanabhan… ", "\u2026", "", -1))
	saveDocuments([]structures.Document{})
 }
**/
/*
func getDocument(scholarURL string){
	// Navigate to the simple playground interface.
	if err := wd.Get(scholarURL); err != nil {
		panic(err)
	}

	// Get a reference to the text box.
	textBox, err := wd.FindElement(selenium.ByID, "gs_hdr_tsi")
	if err != nil {
		panic(err)
	}

	if err := textBox.SendKeys(`TCP performance`); err != nil{
		panic(err)
	}


	//-------------------premo il pulsante cerca------------------------------------------------
	searchButton, err:= wd.FindElement(selenium.ByID, "gs_hdr_tsb")
	if err != nil{
		panic(err)
	}

	if err:=searchButton.Click(); err!=nil {
		panic(err)
	}

	/*url, err := wd.CurrentURL()
	if err!= nil {
		panic(fmt.Sprintf("\ncurrent url: %s\n", url))
	}

	//---------------------------------cerco le info dei documenti-------------------------------------------
	urls, err:= wd.FindElements(selenium.ByXPATH, "//div/h3/a")
	for _, l := range urls {
		url, _:=l.GetAttribute("href")
		fmt.Println(url)
	}

	authors, err:= wd.FindElements(selenium.ByXPATH, "//div[@class='gs_a']")
	for _, a := range authors{
		author, _:=a.Text()
		fmt.Println(author)
	}

	fmt.Println("------------------------------------------")
	//altro = citato da + related + versioni
	altro, err:=wd.FindElements(selenium.ByXPATH, "//div[@class='gs_fl']/a")
	if err!=nil {
		panic("sto cercando i link di: citato da + related + versioni")
	}

	for _, a := range altro{
		testo, _:= a.Text()
		if t, _:= regexp.MatchString("Citato da.*", testo); t{
			words := strings.Split(testo, " ")
			fmt.Println("\nCitato da: ", words[2])
		}
		link, _:= a.GetAttribute("href")
		fmt.Println(testo, "\n", link, "\n")
	}

}
*/

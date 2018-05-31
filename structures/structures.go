package structures

type Document struct {
	Url         string
	Authors     []string
	NumCitedBy  uint16
	LinkCitedBy string
}

var FieldsName = []string{"Url", "Authors", "NumCitedBy", "LinkCitedBy"}

const URLScholar = "https://scholar.google.com/"

const SaveFilePath = "documenti.txt"

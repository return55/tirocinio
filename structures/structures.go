package structures

type Document struct {
	Url         string
	Authors     []string
	NumCitedBy  uint16
	LinkCitedBy string
}

var FieldsName = []string{"Url", "Authors", "NumCitedBy", "LinkCitedBy"}

const URLScholar = "https://scholar.google.com/"

const SaveFilePath = "Documenti.txt"

const MaxReadableDoc = 3000

//Valore da cui ricavo il numero di porta specifico del thread.
// thread_port = threadBasePort + id_del_thread
const ThreadBasePort = 23513

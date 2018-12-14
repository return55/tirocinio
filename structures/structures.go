package structures

//Document e' utilizzata per i documenti presi da Google Scholar
//!!!!!!!!! CAMBIA CitedBy CON Citations !!!!!!!!!!!!!!
type Document struct {
	Url         string   //Singolo link a una sorgente della pubblicazione
	Authors     []string //Alcuni nomi degli autori del documento
	NumCitedBy  uint16   //Numero dei documenti che lo citano
	LinkCitedBy string   //URL a quei documenti
}

//MADocument e' utilizzata per i documenti presi da Microsoft Academic
type MADocument struct {
	Title          string
	URL            string
	Url            sources  //URL dei vari sorgenti disponibili
	Authors        []Author //Nomi, cognomi e affiliazioni dei vari autori
	NumCitations   int64
	LinkCitations  string //Link alla pagina di Academy con i documenti che lo citano
	NumReferences  int64
	LinkReferences string //Link alla pagina di Academy con i documenti che cita
	Abstract       string
	Date           string //time.Time //Data pubblicazione
	FieldsOfStudy  []string
}

//Gli URL di un documento posso essere di 2 tipi: PDF o link a siti Internet
type sources struct {
	PDF []string
	WWW []string
}

//Oltre al nome e cognome dell'autore, memorizzo anche le informazioni sulla sua
//affiliazione che puo' essere diversa da pubblicazione a pubblicazione ma unica
//per ognuna
type Author struct {
	Name        string //Contenuto: "nome cognome"
	Affiliation string //Ente con cui l'autore ha collaborato per scrivere la pubblicazione
}

var FieldsName = []string{"Url", "Authors", "NumCitedBy", "LinkCitedBy"}

var FieldsNameMA = []string{"Title", "Url", "Authors", "NumCitations", "LinkCitations", "NumReferences",
	"LinkReferences", "Abstract", "Date", "FieldsOfStudy"}

const URLScholar = "https://scholar.google.com/"

const URLAcademic = "https://academic.microsoft.com/"

const NumArticlePerPageMA = 8

const SaveFilePath = "DocumentiSerialize.txt"

const MaxReadableDoc = 3000

//Valore da cui ricavo il numero di porta specifico del thread.
// thread_port = threadBasePort + id_del_thread
const ThreadBasePort = 23513

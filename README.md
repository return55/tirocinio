# Document Collector from Scholar
!! Questo codice e' stato sviluppato e testato solo su Debian 9 !!
## Prerequisiti:
Sulla macchina devono essere presenti: la jre 8 e il pacchetto "xvfb" per permettere al web driver (Selenium) di funzionare.
### Firefox
Selenium vuole che in '/usr/bin/firefox' sia presente una versione di firefox non inferiore alla 57.
E' necessario scaricarlo dal sito ufficiale e creare un link alleseguibile:
* sudo ln -s /path-to-firefox-directory/firefox-bin /usr/bin/firefox
### Altri
* Scarica Go dal sito e configuralo (anche GOBIN)
* Scarica e configura git:      user.name, user.email

```
go get github.com/return55/tirocinio
```
## Utilizzo
Al momento per usare il progetto e' necessario andare nella directory del preogetto (dovrebbe essere "go/src/github.com/return55/tirocinio"):
* avvia Neo4j: ./docDatabase/neo4j-community-3.3.5/bin/neo4j start  (stop per fermare)
* lanciare "go run main.go n" con n=numero intero >0 per creare un database con n documenti che citano un documento preimpostato

## Note
### Neo4j
* Se le prestazioni di neo4j sono scarse o se da errore per mancanza di memoria heap, puo' essere utile modificare nel file:  
"docDatabase/neo4j-community-3.3.5/conf/neo4j.conf" il campo "dbms.memory.heap.max_size" e dare al dbms piu' memoria.
* All'avvio di neo4j, il dbms si potrebbe lamentare del max numero di file open. E' possibile modificarne il valore tra le
impostazioni di sicurezza ma anche facendolo non ho notato cambiamenti nelle prestazioni.

## Prerequisiti:
Sulla macchina devono essere presenti: la jre 8 e il pacchetto "xvfb" per permettere al web driver (Selenium) di funzionare.
### Firefox
Selenium vuole che in '/usr/bin/firefox' sia presente una versione di firefox non inferiore alla 57.
E' necessario scaricarlo dal sito ufficiale e creare un link all'eseguibile:
* sudo ln -s /path-to-firefox-directory/firefox-bin /usr/bin/firefox
### Altri
* Scarica Go dal sito e configuralo (anche GOBIN)
* Scarica e configura git:      user.name, user.email

```
go get github.com/return55/tirocinio
```
## Utilizzo
Al momento per usare il progetto e' necessario andare nella directory del preogetto (dovrebbe essere "go/src/github.com/return55/tirocinio"):
* avvia Neo4j: ./docDatabase/neo4j-enterprise-3.4.7/bin/neo4j start  (stop per fermare)
* lancia "go run main_MA.go num_livelli soglia" con:
    num_livelli = quanti livelli avra' (al massimo) l'albero delle citazioni
    soglia = numero di citazioni minime che un documento deve avere per essere preso
    (fai riferimento al file "main_MA.go" per piu' dettagli)

## Note
### Neo4j
* Se le prestazioni di neo4j sono scarse o se da errore per mancanza di memoria heap, puo' essere utile modificare nel file:  
"docDatabase/neo4j-community-3.3.5/conf/neo4j.conf" il campo "dbms.memory.heap.max_size" e dare al dbms piu' memoria.
* All'avvio di neo4j, il dbms si potrebbe lamentare del max numero di file open. E' possibile modificarne il valore tra le
impostazioni di sicurezza ma anche facendolo non ho notato cambiamenti nelle prestazioni.
* Per vedere il grafo:
** Vai all'indirizzo "localhost:7687", accedi a neo4j (se richiesto: username: neo4j - password: neo4j)
** Esegui "match(n) return n"
### Microsoft Academic
* I file con la sigla MA alla fine sono specifici per gli articoli di Academic, tuttavia non sono sufficienti.
Alcune delle funzionalita' di base sono nei rispettivi file senza sigla (es creo_db.go - creo_db_MA.go)
### Main
* main.go (solo per Scholar) ha le seguenti funzionalita':
    firstN <n> : Prendo i primi n articoli che citano quello iniziale.
    everFirst <n> : Prendo per n volte il primo tra gli articoli che citano quello precedente.
    thread <numThreads> <docPerLink> <lenLinkList> : guarda la funzione concurrency() per piu' dettagli.
* main_MA.go (solo per Academic):
    <num_livelli> <soglia> : quanti livelli avra' (al massimo) l'albero delle citazioni e numero di citazioni minime che un documento         deve avere per essere preso.


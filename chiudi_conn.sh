#Serve se invochi "go run main.go thread ..." e il programma va in crash.
#Chiude tutte le connessioni dei thread creati.
PORTA_INIZIO=23514
NUM_PROC=$(( $1 - 1 ))

for porta in $(seq $PORTA_INIZIO $(( $PORTA_INIZIO + $NUM_PROC )));
do
	kill -9 $(netstat -enlp | grep $porta | tr -s " " | cut -d " " -f 9 | cut -d "/" -f 1);
done

kill -9 $(netstat -enlp | grep 8080  | tr -s " " | cut -d " " -f 9 | cut -d "/" -f 1);

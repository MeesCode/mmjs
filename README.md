# mmjs (Mees' Mp3 Jukebox System)

Een mogelijke vervanging voor de huidige mp3 bak software (mjs). De grootste reden voor het bestaan hiervan is dat de mensen achter de originele mp3 bak software geen lid meer zijn en er geen onderhoud meer gepleegd word. Het voordeel van deze versie is dat het geschreven is in go, waardoor het makkelijker is om aan te passen.

## installeren en starten

Het programma kan in 3 modus draaien: filesystem, database en index. filesystem kan je gebruiken zonder enige voorbereiding, echter deze is niet snel genoeg voor gebruik op de Bolk aangezien de muziekbibliotheek te groot is. Database modus maakt gebruik van een mysql database, deze kan lokaal draaien maar ook op een externe server. De index modus scant het bestandssysteem en vult de gekoppelde database met tracks. Er is een database file meegeleverd, deze kan je gebruiken om een mysql scheme mee te initialiseren. 
Het indexeren op de Bolk kan een paar uur duren, houd daar rekening mee.

```config.json.example``` is een voorbeeld van een config file die je kan gebruiken in plaats van commmand line arguments.

### compilen vanaf source
``` 
$ go mod vendor
$ go build
```

### programma starten
normaal
```
$ ./mmjs <path>
```
met een config file
```
$ ./mmjs -c config.json /pub/mp3
```
alle opties
```
$ ./mmjs --help
```

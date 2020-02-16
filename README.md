# mp3bak2

Een mogelijke vervanging voor de huidige mp3 bak software. De grootste reden voor het bestaan hiervan is dat de mensen achter de originele mp3 bak software geen lid meer zijn en er geen onderhoud meer gepleegd word. Het voordeel van deze versie is dat het geschreven is in go, waardoor het makkelijker is om aan te passen.

## installeren en starten

### setup database
Er is een database file meegeleverd, deze kan je gebruiken om een mysql scheme mee te initialiseren. Hierna moet je index modus gebruiken om de database te populaten. Tenzij je filesystem modus gebruikt zal je dit moeten doen. Als je database modus gebruikt is het aan te raden deze eens in de zoveel tijd te updaten doormiddel van index modus. (let op: deze haalt je database niet leeg, voegt alleen entries toe)

in ```./globals/config.go.example``` staat een voorbeeld van een ```config.go``` bestand waar je de database credentials kan invullen.

### compilen vanaf source
``` 
$ go get .
$ go build -o mp3bak2 *.go
$ ./mp3bak2
```
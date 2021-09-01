package plugins

import (
    "log"
	"github.com/tarm/serial"
	"github.com/MeesCode/mmjs/audioplayer"
	"github.com/MeesCode/mmjs/globals"
)


func Coinslot() {

	config := &serial.Config{
		Name: globals.Config.Serial.Port,
		Baud: 9600,
		Size: 8,
	}

	stream, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 1)

	for {
		_, err = stream.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		
		audioplayer.Nextsong()
	}
}
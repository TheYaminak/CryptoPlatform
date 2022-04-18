package handlers

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"os"
)

func GetLogs() func(http.ResponseWriter, *http.Request) {
	return func(resWriter http.ResponseWriter, req *http.Request) {
		file, err := os.Open("log.log")
		if err != nil {
			log.Printf("GetLogs() os.Open(log.log) err: %v\n", err)

			return
		}
		defer file.Close()

		buf := new(bytes.Buffer)

		fileScanner := bufio.NewScanner(file)
		for fileScanner.Scan() {
			if fileScanner.Text() == "" {
				continue
			}

			_, err = buf.WriteString(fileScanner.Text() + "\n")
			if err != nil {
				log.Printf("GetLogs() buf.WriteString(trade) err: %v\n", err)
			}
		}

		err = fileScanner.Err()
		if err != nil {
			log.Printf("Error while reading file: %s\n", err)

			return
		}

		if len(buf.Bytes()) == 0 {
			_, err = resWriter.Write([]byte("No logs"))
			if err != nil {
				log.Printf("resWriter.Write([]byte(No logs)) err: %v\n", err)
			}

			return
		}

		_, err = resWriter.Write(buf.Bytes())
		if err != nil {
			log.Printf("GetLogs() resWriter.Write(buf.Bytes()) err: %v\n", err)
		}
	}
}

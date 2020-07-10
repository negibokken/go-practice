package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
	<html>
		<head>
			<title>チャット</title>
		</head>
		<body>
			チャットしましょう！
		</body>
	</html>
`))
	})

	fmt.Println("Listening :8080!")
	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

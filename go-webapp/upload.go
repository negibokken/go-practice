package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

func uploaderHandler(w http.ResponseWriter, req *http.Request) {
	userId := req.FormValue("userid")
	file, header, err := req.FormFile("avatarFile")
	if err != nil {
		log.Println("Form data error")
		io.WriteString(w, err.Error())
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Uploaded data read error")
		io.WriteString(w, err.Error())
		return
	}
	filename := filepath.Join("avatars", userId+filepath.Ext(header.Filename))
	fmt.Println("filename: ", filename)
	err = ioutil.WriteFile(filename, data, 0777)
	if err != nil {
		log.Println("Form data write error")
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, "成功")
}

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type SubmissionInfo struct {
	CurrentProgram string
	Programs       []string
}

func main() {
	programlist := os.Getenv("SNW_PROGRAMLIST")
	currentprogram := os.Getenv("SNW_PROGRAM")
	if currentprogram == "" || programlist == "" {
		fmt.Println("Please ensure SNW_PROGRAM (current term/program) and SNW_PROGRAMLIST ('|' delimited list of program options) environment variables are set")
		os.Exit(1)
	}

	splitprogramlist := strings.Split(programlist, "|")
	templateData := SubmissionInfo{
		CurrentProgram: currentprogram,
		Programs:       splitprogramlist,
	}
	tmpl := template.Must(template.ParseFiles("template.html"))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { tmpl.Execute(w, templateData) })
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/submit", handleSubmission)
	fmt.Println("Listening on port 8080")
	log.Fatalln(http.ListenAndServe(":8080", mux))
}

func handleSubmission(w http.ResponseWriter, r *http.Request) {
	fmt.Println("submission received")
	r.ParseForm()
	fmt.Printf("formy stuff: %v", r.Form)

	tmpl2 := template.Must(template.ParseFiles("submit.html"))

	//put in sendgrid stuff here
	fmt.Println("past template")
	var err error
	err = nil
	fmt.Println("past err stuff")

	if err != nil {
		fmt.Println("error submitting info w/sendgrid, ", err, time.Now().Format(time.RFC850))
		err = tmpl2.Execute(w, struct {
			Message string
		}{
			"error during submission process",
		})
		if err != nil {
			fmt.Println("error rendering failed submission template: ", err)
		}
		http.Redirect(w, r, "/", 301)
	}

	fmt.Println("submission successful at", time.Now().Format(time.RFC850))
	err = tmpl2.Execute(w, struct {
		Message string
	}{
		"Submission Received",
	})
	if err != nil {
		fmt.Println("error rendering successful submission template: ", err)
		http.Redirect(w, r, "/", 301)
	}
}

func jsonResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprint(w, message)
}

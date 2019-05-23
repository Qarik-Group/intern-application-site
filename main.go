package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"net/smtp"

	"github.com/jhunt/vcaptive"
)

type SubmissionInfo struct {
	CurrentProgram  string
	Programs      []string
}

func main() {
	//bunch of ugly env variable checking
	programlist := os.Getenv("SNW_PROGRAMLIST")
	currentprogram := os.Getenv("SNW_PROGRAM")
	if currentprogram == "" || programlist == "" {
		fmt.Println("Please ensure SNW_PROGRAM (current term/program) and SNW_PROGRAMLIST ('|' delimited list of program options) environment variables are set")
		os.Exit(1)
	}
	if os.Getenv("VCAP_SERVICES") == "" {
		fmt.Fprintf(os.Stderr, "VCAP_SERVICES: not found in environment\n")
		os.Exit(1)
	}

	services, err := vcaptive.ParseServices(os.Getenv("VCAP_SERVICES"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "VCAP_SERVICES: %s\n", err)
		os.Exit(1)
	}

	instance, found := services.Tagged("smtp")
	if !found {
		fmt.Fprintf(os.Stderr, "VCAP_SERVICES: no 'smtp' service found smtp\n")
		os.Exit(2)
	}

	muser, found := instance.GetString("username")
	if !found {
		fmt.Fprintf(os.Stderr, "VCAP_SERVICES: '%s' service is missing required credential username\n", instance.Label)
		os.Exit(2)
	}

	mpass, found := instance.GetString("password")
	if !found {
		fmt.Fprintf(os.Stderr, "VCAP_SERVICES: '%s' service is missing required credential password\n", instance.Label)
		os.Exit(2)
	}

	mhost, found := instance.GetString("hostname")
	if !found {
		fmt.Fprintf(os.Stderr, "VCAP_SERVICES: '%s' service is missing required credential hostname\n", instance.Label)
		os.Exit(2)
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
	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("submission received, parsing form ...")

		r.ParseForm()
		var formData map[string]string
		for k, v := range r.Form {
			formData[k] = strings.Join(v,"")
		}

		tmpl2 := template.Must(template.ParseFiles("submit.html"))

		auth := smtp.PlainAuth("", muser+"@"+mhost, mpass, mhost)
		to := []string{"aanelli@starkandwayne.com"}
		msg := []byte("To: aanelli@starkandwayne.com\r\n" +
			"Subject: discount Gophers!\r\n" +
			"\r\n" +
			"This is the email body.\r\n")
		err := smtp.SendMail(mhost+":25", auth, "aanelli@starkandwayne.com", to, msg)
		if err != nil {
			fmt.Println("erorr sending smtp message with sendgrid: ", err, time.Now().Format(time.RFC850))
			w.WriteHeader(500)
		}

		fmt.Println("submission successful at:", time.Now().Format(time.RFC850))
		err = tmpl2.Execute(w, struct {Message string}{"Submission Received",})
		if err != nil {
			fmt.Println("error rendering successful submission template: ", err)
			w.WriteHeader(500)
		}
    })
	fmt.Println("Listening on port 8080")
	log.Fatalln(http.ListenAndServe(":8080", mux))
}

func jsonResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprint(w, message)
}

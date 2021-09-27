// Before using this tool you need to register, to do it use: ./naruken init then answer on the questions
// To submit the flag that you found use following command: ./naruken submit -flag <flag>
// Naruken tool will create the json file where will be stored a unique user id that will be randomly generated.
// This tool will interact with the API server - https://api.narukoshin.me. The data will be stored in the SQLiTE database that will be strongly protected, maybe with encryption
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const folderName string = ".narukeFolder"

// TODO: Create function that will create the folder and settings.json file and \
// ...verify if the folder and file exists

func main(){
	if !(len(os.Args) > 1){
		fmt.Println("Usage: ./naruken <command> [options...]")
		fmt.Printf("\nAvailable commands:\ninit	Creates your CTF account\n		-name Your full name (example: \"Naru Koshin\")\n		-course Your course of study (example: 4KT)\n\nsubmit	When you found the flag, you need to submit it\n		-flag	The flag that you found in your scope\n\nscore	You can view the scoreboard\n")
		return
	}
	command := os.Args[1]
	switch command {
		case "init":
			cmdInit()
		case "submit":
			cmdSubmit()
		case "score":
			cmdScore()
		case "end":
			cmdEnd()
		default:
			fmt.Println("Unknown command " + command)
	}
}

type CTFMemberData struct {
	UID			string	`json:"user_id"`
	Name 		string	`json:"name"`
	Course		string	`json:"course"`
}

type CTFMemberUID struct {
	UID			string 	`json:"uid"`
}

type CTFSubmitFlag struct {
	Flag		string `json:"flag"`
	UID			string `json:"user_id"`
}

// Registering the CTF member in the system then retrieving the unique user id from the server
func cmdInit(){
	var (
		name,
		course string
	)
	init := flag.NewFlagSet("init", flag.ExitOnError)
	init.StringVar(&name, "name", "", "Your full name")
	init.StringVar(&course, "course", "", "Your course of study")
	init.Parse(os.Args[2:])

	if name == "" || course == "" {
		fmt.Printf("Please provide the required flags\n\n*Before the registration ensure that your data provided is correct, otherwise, your participation will be declined\n")
		fmt.Printf("\nRequired flags:\n-name		Your full name (example: \"Naru Koshin\")\n-course		Your course of study (example: 4KT)\n")
		return
	}

	// Creating one time registration.
	// When user will run this command, this piece of code will check if settings.json exists and if not then that will be created
	if _, err := os.Stat(folderName + "/settings.json"); os.IsNotExist(err) {
		err := os.Mkdir(folderName, 0777)
		if err != nil {
			log.Fatal(err)
		}
		file, err := os.Create(folderName + "/settings.json")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	} else {
		fmt.Println("You are already registered for the CTF.\nIf there is any mistake, please contact your CTF organizer.")
		return

	}

	// Registering the CTf participient in the server
	participient := CTFMemberData {
		Name: name,
		Course: course,
	}
	participient.Register()
}

func (p *CTFMemberData) Register(){
	js, err := json.Marshal(p)
	if err  != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", 
		"https://api.narukoshin.me/Register/068b661109426b3e284c0ef892eba655a776177cd248a4294c24c861c0573748", 
		bytes.NewReader(js))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// getting the UID from the server
		var memberUID CTFMemberUID
		err = json.NewDecoder(resp.Body).Decode(&memberUID)
		if err != nil {
			log.Fatal(err)
		}
		// Joining together user ID with another information that I already have
		p.UID = memberUID.UID
		file, err := os.OpenFile(folderName + "/settings.json", os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		json.NewEncoder(file).Encode(p)
		fmt.Println("You successfuly registered for the CTF... Good Luck and Have Fun. :)")
	}	
}

// When user submits the flag
func cmdSubmit(){
	var _flag string
	submit := flag.NewFlagSet("submit", flag.ExitOnError)
	submit.StringVar(&_flag, "flag", "", "The flag that you found in vulnerable site")
	submit.Parse(os.Args[2:])
	if _flag == "" {
		fmt.Printf("Please provide the required flags\n\nRequired flags:\n-flag		The flag that you found in vulnerable site\r\n")
		return
	}
	// Checking if the user is executed the init command
	var memberData CTFMemberData
	if _, err := os.Stat(folderName + "/settings.json"); os.IsNotExist(err) {
		fmt.Println("To submit the flag, please run the init command at first.")
		return
	} else {
		// reading the data from settings.json
		file, err := os.OpenFile(folderName + "/settings.json", os.O_RDONLY, 0777)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		err = json.NewDecoder(file).Decode(&memberData)
		if err != nil {
			fmt.Println("Failed to verify your data, please contact your CTF organizer.")
			return
		}
	}
	// Checking if the flag is in correct format
	if ok, _ := regexp.MatchString("^NARU\\{([A-Z0-9]+){32,}\\}$", strings.TrimSpace(_flag)); !ok {
		fmt.Println("Wrong flag format provided.")
		return
	}
	flag := CTFSubmitFlag{
		Flag: _flag,
		UID: memberData.UID,
	}
	js, err := json.Marshal(flag)
	if err != nil {
		log.Fatal(err)
	}
	// Sending the request to the server
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("POST", "https://api.narukoshin.me/Submit/c86eca4b5dcc75000c39e9963fa05bc9aca0b20d0e8118ac849ea59ade081bf1", bytes.NewReader(js))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Receiving the response from the server
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(body))
	}
}

// The admin management interface
func cmdScore(){
	if _, err := os.Stat(folderName + "/settings.json"); os.IsNotExist(err) {
		fmt.Println("You need to register for the CTF to use this command.")
		return
	}
	// getting the data from the API server
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", "https://api.narukoshin.me/suka", nil)
	if err != nil {
		log.Fatal(err)
	}

	type member struct {
		Name string `json:"name"`
		Course string `json:"course"`
		RegDate string `json:"registered_at"`
		Points string `json:"points"`
	}
	var members []member

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	json.NewDecoder(resp.Body).Decode(&members)

	fmt.Println(" ______________________________________________________")
	fmt.Println("|[ID] [FULL NAME] [COURSE] [POINTS] [REGISTRATION DATE]|")
	fmt.Println(" ￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣￣")
	count := 0
	for _, member := range members {
		count++
		date, err := time.Parse(time.RFC3339, member.RegDate)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[%d] {%s} {%s} {%s} {%s}\n", count, member.Name, member.Course, member.Points, date)
	}
}

// Deleting the files that the tool created
func cmdEnd(){
	if _, err := os.Stat(folderName); !os.IsNotExist(err) {
		// Deleting the config file
		err = os.Remove(folderName + "/settings.json")
		if err != nil {
			log.Fatal(err)
		}
		// Deleting the directory
		err = os.Remove(folderName)
		if err != nil {
			log.Fatal(err)
		}
	}
}
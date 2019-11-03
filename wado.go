package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/radovskyb/watcher"
)

// Config structu
type Config struct {
	Rules []Rule `json:"rules"`
}

//Rule struct
type Rule struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Command      []string `json:"command"`
	ShellCommand string   `json:"shellCommand"`
	EventFilter  string   `json:"eventFilter"`
	FileFilter   string   `json:"fileFilter"`
}

func executeCommand(rule Rule, event watcher.Event) (output string, err error) {
	var cmd *exec.Cmd
	data := map[string]string{
		"fileRelative": event.Name(),
		"fileAbsolute": event.Path,
		"fileName":     event.Name(),
		"dir":          rule.Path,
		"rule":         rule.Name,
	}

	if rule.ShellCommand != "" {
		t := template.Must(template.New("").Parse(rule.ShellCommand))
		buf := bytes.Buffer{}
		t.Execute(&buf, data)
		cmd = exec.Command("bash", "-c", buf.String())
	} else {
		cmd = exec.Command(rule.Command[0], rule.Command[1:]...)
	}
	out, err := cmd.CombinedOutput()

	return string(out), err
}

func watch(w *watcher.Watcher, rule Rule) {
	for {
		select {
		case event := <-w.Event:
			log.Println("Queue length", len(w.Event))
			log.Println("Event from rule ", rule.Name, event) // Print the event's info.
			output, err := executeCommand(rule, event)
			if err != nil {
				log.Println("Cmd failed: ", err)
			}
			log.Printf("** Cmd finished %s\n", output)
		case err := <-w.Error:
			log.Fatalln("Watcher failed", rule.Name, err)
		case <-w.Closed:
			log.Println("watcher stopped ", rule)
			return
		}
	}
}

func (rule Rule) String() string {
	if rule.Name != "" {
		return rule.Name
	}
	return "Watch Path " + rule.Path

}

func main() {
	configFile := flag.String("c", "wado.json", "a path to wado config file")
	flag.Parse()
	var config Config

	jsonFile, err := os.Open(*configFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Fatalln("Error loading config file. Please specify -c param to set custom config file location", err)
	}
	log.Println("Successfully processed config from: ", *configFile)
	configFileContents, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalln("Error reading config file", err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	err = json.Unmarshal(configFileContents, &config)
	if err != nil {
		log.Fatalf("Could not parse config file %s", err)
	}

	log.Printf("Config file contains %d rules\n", len(config.Rules))
	for i, v := range config.Rules {
		log.Printf("Rule: %d, %s", i, v)
		w := watcher.New()

		// Watch this folder for changes.
		if err := w.AddRecursive(v.Path); err != nil {
			log.Fatalln(err)
		}

		// Print a list of all of the files and folders currently
		// being watched and their paths.
		for path, f := range w.WatchedFiles() {
			log.Printf("\tRule watching file: %s: %s\n", path, f.Name())
		}

		// log.Printf("starting watch goroutine")
		go watch(w, v)
		// log.Printf("watcher goroutine started...")
		// Start the watching process - it'll check for changes every 100ms.
		go func() {
			if err := w.Start(time.Millisecond * 2000); err != nil {
				log.Fatalln(err)
			}
		}()
	}

	for {

	}

}

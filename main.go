package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	type Platform struct {
		Nome    string `json:"nome"`
		Arquivo string `json:"arquivo"`
	}

	http.HandleFunc("/platforms/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/platforms/"):]
		out, err := exec.Command("/usr/bin/ia", "list", id).Output()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to execute command: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		var result []map[string]string
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasSuffix(line, ".cue") ||
				strings.HasSuffix(line, ".bin") ||
				strings.HasSuffix(line, ".7z") ||
				strings.HasSuffix(line, ".zip") ||
				strings.HasSuffix(line, ".chd") ||
				strings.HasSuffix(line, ".pbp") ||
				strings.HasSuffix(line, ".cdi") {

				filename := path.Base(line)
				name := strings.TrimSuffix(filename, path.Ext(filename))

				result = append(result, map[string]string{
					"nome":    name,
					"arquivo": line,
				})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		// Parse URL parameters
		params := r.URL.Query()
		platform := strings.TrimPrefix(r.URL.Path, "/download/")
		arquivo := params.Get("arquivo")

		// Check if platform and arquivo are valid
		if platform == "" || arquivo == "" {
			http.Error(w, "Invalid URL parameters", http.StatusBadRequest)
			return
		}

		// Execute wget command in /roms/:platform directory
		cmd := exec.Command("wget", "-b", "-o", "/roms/ports/amberserver/download.log", "https://archive.org/download/"+arquivo)
		cmd.Dir = fmt.Sprintf("/roms/%s", platform)
		err := cmd.Start()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to execute command: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		err = cmd.Wait()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to execute command: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Remove wget.log file
		err = os.Remove(fmt.Sprintf("/roms/ports/amberserver/download.log"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to remove file: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// Return success message
		fmt.Fprintf(w, "Download completed for arquivo=%s in /roms/%s\n", arquivo, platform)
	})

	http.HandleFunc("/progress", func(w http.ResponseWriter, r *http.Request) {
		content, err := ioutil.ReadFile("download.log")
		if err != nil {
			fmt.Fprint(w, "0%")
			return
		}

		var highestPercent int
		re := regexp.MustCompile(`\d+%`)
		matches := re.FindAll(content, -1)
		for _, match := range matches {
			percentStr := strings.TrimRight(string(match), "%")
			percent, err := strconv.Atoi(percentStr)
			if err != nil {
				continue
			}
			if percent > highestPercent {
				highestPercent = percent
			}
		}

		fmt.Fprintf(w, "%d%%", highestPercent)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

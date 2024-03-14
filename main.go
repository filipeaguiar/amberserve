package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func downloadIA() error {
	iaPath := "/roms/ports/amberserver/ia"
	if _, err := os.Stat(iaPath); os.IsNotExist(err) {
		fmt.Println("Downloading ia executable...")
		err := exec.Command("wget", "-O", iaPath, "https://archive.org/download/ia-pex/ia").Run()
		if err != nil {
			return fmt.Errorf("failed to download ia executable: %s", err)
		}
		err = os.Chmod(iaPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to set ia executable permissions: %s", err)
		}
	}
	return nil
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	platform := strings.TrimPrefix(r.URL.Path, "/download/")
	platform = strings.TrimSuffix(platform, r.URL.RawQuery)
	arquivo := r.URL.Query().Get("arquivo")
	arquivo = strings.ReplaceAll(arquivo, "%20", " ")
	exec.Command("echo", "0%", ">", "/roms/amberserver/download.log").Run()
	cmd := exec.Command("/roms/ports/amberserver/downloader.sh", platform, arquivo)
	fmt.Println(cmd.String())
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(w, "Arquivo baixado com sucesso!")
}

func main() {
	err := downloadIA()
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	type Platform struct {
		Nome    string `json:"nome"`
		Arquivo string `json:"arquivo"`
	}

	http.HandleFunc("/platforms/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POSTS, OPTIONS")
		w.Header().Set("Access-control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		id := r.URL.Path[len("/platforms/"):]
		out, err := exec.Command("/roms/ports/amberserver/ia", "list", id).Output()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to execute command: %s", err), http.StatusInternalServerError)
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

	http.HandleFunc("/download/", downloadHandler)

	http.HandleFunc("/progress", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POSTS, OPTIONS")
		w.Header().Set("Access-control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		content, err := ioutil.ReadFile("/roms/ports/amberserver/download.log")
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

	if err = http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

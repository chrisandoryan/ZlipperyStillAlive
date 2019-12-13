package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"time"

	"github.com/go-macaron/renders"
	"github.com/go-macaron/session"
	"github.com/nwaples/rardecode"
	"gopkg.in/macaron.v1"
)

func writeNewFile(fileName string, rr *rardecode.Reader, headerMode os.FileMode) {
	fmt.Println(fileName)
	newFile, err := os.Create(fileName)
	defer newFile.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(reflect.TypeOf(newFile))

	_, err = io.Copy(newFile, rr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Println("Whoa.")
	return
}

func extract(rarName string, destDir string) {
	theRar, err := os.Open(rarName)

	if err != nil {
		fmt.Println(err)
		return
	}

	rr, err := rardecode.NewReader(theRar, "")
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		header, err := rr.Next()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("headerName", header.Name)
		writeNewFile(filepath.Join(destDir, header.Name), rr, header.Mode())
	}
	defer theRar.Close()
}

func putTimeScript(sessPath string) {
	timeScript := "#!/bin/sh\n\ncurrent_date_time=\"`date \"+%Y-%m-%d %H:%M:%S\"`\";\necho $current_date_time;"

	f, err := os.Create(filepath.Join(sessPath, "time.sh"))

	if err != nil {
		fmt.Println(err)
	}

	_, err = f.WriteString(timeScript)

	if err != nil {
		fmt.Println(err)
	}

	f.Sync()
}

func initialize(sessID string) {
	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	sessDir := filepath.Join(baseDir, "storage", sessID)

	if _, err := os.Stat(sessDir); os.IsNotExist(err) {
		os.Mkdir(sessDir, 0755)
		putTimeScript(sessDir)
	}

	if err != nil {
		fmt.Println(err)
	}

}

func getServerTime(sessID string) string {
	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	pathToTimeScript := filepath.Join(baseDir, "storage", sessID, "time.sh")

	cmd := exec.Command("/bin/sh", pathToTimeScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("cmd.Run() failed with %s\n", err)
	}

	return string(out)
}

func main() {
	m := macaron.Classic()
	m.Use(renders.Renderer(
		renders.Options{
			Directory:       "templates",                // Specify what path to load the templates from.
			Extensions:      []string{".tmpl", ".html"}, // Specify extensions to load for templates.
			IndentJSON:      true,                       // Output human readable JSON
			IndentXML:       true,                       // Output human readable XML
			HTMLContentType: "text/html",                // Output XHTML content type instead of default "text/html"
		}))
	m.Use(macaron.Static("public",
		macaron.StaticOptions{
			Prefix:      "public",
			SkipLogging: true,
			IndexFile:   "index.html",
			Expires: func() string {
				return time.Now().Add(24 * 60 * time.Minute).UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
			},
		}))
	m.Use(macaron.Static("storage",
		macaron.StaticOptions{
			Prefix:      "storage",
			SkipLogging: true,
			Expires: func() string {
				return time.Now().Add(24 * 60 * time.Minute).UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
			},
		}))
	m.Use(macaron.Recovery())
	m.Use(session.Sessioner())

	m.Get("/", func(r renders.Render, s session.Store) {
		initialize(s.ID())
		time := getServerTime(s.ID())
		fmt.Println(time)
		r.HTML(200, "pages/index.html", map[string]interface{}{"Time": time})
	})

	m.Post("/upload", func(w http.ResponseWriter, r *http.Request, s session.Store) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		sessID := s.ID()
		if err := r.ParseMultipartForm(2048); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sender := r.FormValue("sender")
		receiver := r.FormValue("receiver")

		uploadedFile, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer uploadedFile.Close()

		baseDir, err := os.Getwd()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filename := handler.Filename
		fmt.Println(filename)

		if sender != "" && receiver != "" {
			filename = fmt.Sprintf("%s%s", sender, filepath.Ext(handler.Filename))
		}

		uploadDir := filepath.Join(baseDir, "storage", sessID, "uploads")

		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.Mkdir(uploadDir, 0755)
		}

		imgDir := filepath.Join(uploadDir, "graphics")
		etcDir := filepath.Join(uploadDir, "others")

		if _, err := os.Stat(imgDir); os.IsNotExist(err) {
			os.Mkdir(imgDir, 0755)
		}

		if _, err := os.Stat(etcDir); os.IsNotExist(err) {
			os.Mkdir(etcDir, 0755)
		}

		if filepath.Ext(handler.Filename) == ".rar" {
			fileLocation := filepath.Join(etcDir, filename)
			targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer targetFile.Close()

			if _, err := io.Copy(targetFile, uploadedFile); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			extract(fileLocation, etcDir)

			err = os.Remove(fileLocation)
			if err != nil {
				fmt.Println(err)
				return
			}

			w.Write([]byte("Done! Your archive has been extracted to " + etcDir + " </br><a href=\"/\">Go back</a>"))

		} else if filepath.Ext(handler.Filename) == ".jpg" || filepath.Ext(handler.Filename) == ".png" {
			fileLocation := filepath.Join(imgDir, filename)
			targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer targetFile.Close()

			if _, err := io.Copy(targetFile, uploadedFile); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Write([]byte("Done! Your image has been stored in " + fileLocation + " </br><a href=\"/\">Go back</a>"))

		} else if filepath.Ext(handler.Filename) == ".zip" {
			w.Write([]byte("Zip is not yet supported!</br><a href=\"/\">Go back</a>"))
		}

	})

	m.Run()
}

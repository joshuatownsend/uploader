package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sontags/env"
)

const (
	form = `<!DOCTYPE html>
	<html lang="en">
	  <head>
	    <title>File Upload</title>
	  </head>
	  <body>
	    <div class="container">
	      <h1>File Upload</h1>
	      <div class="message">{{.}}</div>
	      <form class="form-signin" method="post" action="/" enctype="multipart/form-data">
	          <fieldset>
	            <input type="file" name="myfiles" id="myfiles" multiple="multiple">
	            <input type="submit" name="submit" value="Submit">
	        </fieldset>
	      </form>
	    </div>
	  </body>
	</html>`
)

//Display the named template
func displayHandler(templates *template.Template, message interface{}) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var out bytes.Buffer
		templates.ExecuteTemplate(&out, "Form", message)
		c.Data(http.StatusOK, "text/html", out.Bytes())
	})
}

func uploadHandler(templates *template.Template, dest string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {

		reader, err := c.Request.MultipartReader()

		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}

		//copy each part to destination.
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			//if part.FileName() is empty, skip this iteration.
			if part.FileName() == "" {
				continue
			}
			dst, err := os.Create(dest + part.FileName())
			defer dst.Close()

			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := io.Copy(dst, part); err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		var out bytes.Buffer
		templates.ExecuteTemplate(&out, "Form", "Uploaded successfully...")
		c.Data(http.StatusOK, "text/html", out.Bytes())

	})
}

func main() {
	var port, dest string
	env.Var(&port, "PORT", "8989", "Port that should be binded")
	env.Var(&dest, "DEST", "/tmp/", "Where uploaded files will be placed")
	env.Parse("U")

	templates := template.Must(template.New("Form").Parse(form))

	log.Println("listening on port", port)
	log.Println("writing data to", dest)

	r := gin.Default()
	r.GET("/", displayHandler(templates, nil))
	r.POST("/", uploadHandler(templates, dest))
	r.Run(":" + port)
}

package serveImage

import(
	// "encoding/json"
	// "fmt"
    
	// "encoding/json"
	// "sort"
	"net/http"
	// "os"
	// "path/filepath"
	// "sathya-narayanan23/crudapp/database"
	// "sathya-narayanan23/crudapp/users/user"

    // "strconv"
    
	"github.com/gorilla/mux"
)


func ServeImagenews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Construct the file path based on your image storage setup
	filePath := "banner/" + filename

	// Serve the image
	http.ServeFile(w, r, filePath)
}

func ServeImagelogos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Construct the file path based on your image storage setup
	filePath := "uploads/logos/" + filename

	// Serve the image
	http.ServeFile(w, r, filePath)
}

func ServeImageuploads(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Construct the file path based on your image storage setup
	filePath := "uploads/" + filename

	// Serve the image
	http.ServeFile(w, r, filePath)
}

func ServeImageimage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Construct the file path based on your image storage setup
	filePath := "image/" + filename

	// Serve the image
	http.ServeFile(w, r, filePath)
}

func ServeImageqrcode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Construct the file path based on your image storage setup
	filePath := "qrcodes/" + filename

	// Serve the image
	http.ServeFile(w, r, filePath)
}


func ServeImagepdfFilesends(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Construct the file path based on your image storage setup
	filePath := "uploads/pdfFilesends/" + filename

	// Serve the image
	http.ServeFile(w, r, filePath)
}

func ServeImagenew(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["filePath"]

	// Serve the image
	http.ServeFile(w, r, filePath)
}





package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	//FormatHeader is the Header name for extracting format
	FormatHeader = "X-Format"
	//CodeHeader is the Header name for extracting status code
	CodeHeader = "X-Code"
	//ContentType Header name for content type
	ContentType = "Content-Type"
	//OriginalURI is the header name to the original url
	OriginalURI = "X-Original-URI"
	//Namespace is the header name to the namespace nginx controller is running
	Namespace = "X-Namespace"
	//IngressName is the header name to the name of the ingress
	IngressName = "X-Ingress-Name"
	//ServiceName is the header name to  the current service with errors
	ServiceName = "X-Service-Name"
	//ServicePort is the header name to the port of the service with errors
	ServicePort = "X-Service-Port"
	//RequestID is the header name to the request id of the request that came in
	RequestID = "X-Request-ID"
	//ErrFilesPath is just the path to the error pages
	ErrFilesPath = "ERROR_FILES_PATH"
)

var debug = os.Getenv("DEBUG")
var isDeBugMode = debug != ""

func healthHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
}

func loadEnv() {
	if isDeBugMode {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file", err)
		}
		log.Printf("App Running in debug mode")
	}
}

func addHeadersToResInDebugMode(res http.ResponseWriter, req *http.Request) {

	if isDeBugMode {
		log.Print("Adding headers")
		res.Header().Set(FormatHeader, req.Header.Get(FormatHeader))
		res.Header().Set(CodeHeader, req.Header.Get(CodeHeader))
		res.Header().Set(ContentType, req.Header.Get(ContentType))
		res.Header().Set(OriginalURI, req.Header.Get(OriginalURI))
		res.Header().Set(Namespace, req.Header.Get(Namespace))
		res.Header().Set(IngressName, req.Header.Get(IngressName))
		res.Header().Set(ServiceName, req.Header.Get(ServiceName))
		res.Header().Set(ServicePort, req.Header.Get(ServicePort))
		res.Header().Set(RequestID, req.Header.Get(RequestID))
	}
}
func getFormat(req *http.Request) string {
	format := req.Header.Get(FormatHeader)
	if format == "" {
		format = "text/html"
		log.Printf("Using default format %s", format)
	}
	return format
}

func getExtension(format string) string {
	var ext = "html"
	cext, err := mime.ExtensionsByType(format)

	if err != nil {
		log.Printf("error getting extension using  %s", format)
	} else if len(cext) == 0 {
		log.Printf("no media type extension using %s", format)
	} else if len(cext) > 0 {
		ext = cext[0]
	}

	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

func getAbsoluteErrorPath(path string, code int, ext string) string {
	return fmt.Sprintf("%v/%v%v", path, code, ext)
}

func getStatusCode(req *http.Request) int {
	errCode := req.Header.Get(CodeHeader)
	code, err := strconv.Atoi(errCode)

	if err != nil {
		code = 404
		log.Printf("error getting status code %v with error %v", code, err)
	}
	return code
}

func getBaseErrorFilePath() string {
	path := os.Getenv(ErrFilesPath)
	if path == "" {
		path = "./errors"
	}
	return path
}

func errorHandler(res http.ResponseWriter, req *http.Request) {
	addHeadersToResInDebugMode(res, req)
	format := getFormat(req)
	ext := getExtension(format)
	code := getStatusCode(req)
	file := getAbsoluteErrorPath(getBaseErrorFilePath(), code, ext)

	res.WriteHeader(code)

	f, err := os.Open(file)

	if err != nil {
		log.Printf("error opening file %v", err)
		http.NotFound(res, req)
		return
	}

	log.Printf("serving custom error for code  %v and format %v and file %v", code, format, file)

	io.Copy(res, f)
	defer f.Close()
}
func main() {
	loadEnv()

	address := ":8080"

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", errorHandler)
	log.Printf("Listening on %s", address)

	log.Fatal(http.ListenAndServe(address, nil))
}

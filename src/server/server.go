package main

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

type ServerOptions struct {
	Port              int
	Burst             int
	Concurrency       int
	HttpCacheTtl      int
	HttpReadTimeout   int
	HttpWriteTimeout  int
	CORS              bool
	Gzip              bool
	AuthForwarding    bool
	EnableURLSource   bool
	EnablePlaceholder bool
	Address           string
	PathPrefix        string
	ApiKey            string
	Mount             string
	CertFile          string
	KeyFile           string
	Authorization     string
	Placeholder       string
	PlaceholderImage  []byte
	AlloweOrigins     []*url.URL
	MaxAllowedSize    int
	CFDistributionId  string
}

func Server(o ServerOptions) error {
	addr := o.Address + ":" + strconv.Itoa(o.Port)

	handler := NewLog(NewServerMux(o), os.Stdout)

	server := &http.Server{
		Addr:           addr,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    time.Duration(o.HttpReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(o.HttpWriteTimeout) * time.Second,
	}

	return listenAndServe(server, o)
}

func listenAndServe(s *http.Server, o ServerOptions) error {
	if o.CertFile != "" && o.KeyFile != "" {
		return s.ListenAndServeTLS(o.CertFile, o.KeyFile)
	}
	return s.ListenAndServe()
}

func join(o ServerOptions, route string) string {

	return path.Join(o.PathPrefix, route)
}

// NewServerMux creates a new HTTP server route multiplexer.
func NewServerMux(o ServerOptions) http.Handler {
	mux := http.NewServeMux()

	mux.Handle(join(o, "/"), Middleware(indexController, o))
	mux.Handle(join(o, "/form"), Middleware(formController, o))
	mux.Handle(join(o, "/health"), Middleware(healthController, o))
	mux.Handle(join(o, "/uploadfile"), Middleware(uploadFileController, o))
	mux.Handle(join(o, "/attachments"), Middleware(attachmentsController, o))
	mux.Handle(join(o, "/public-upload-file"), UnAuthMiddleware(publicUploadFileController, o))
	mux.Handle(join(o, "/upload-by-email"), UnAuthMiddleware(uploadByEmailController, o))

	image := ImageMiddleware(o)
	mux.Handle(join(o, "/resize"), image(Resize))
	mux.Handle(join(o, "/enlarge"), image(Enlarge))
	mux.Handle(join(o, "/extract"), image(Extract))
	mux.Handle(join(o, "/crop"), image(Crop))
	mux.Handle(join(o, "/rotate"), image(Rotate))
	mux.Handle(join(o, "/flip"), image(Flip))
	mux.Handle(join(o, "/flop"), image(Flop))
	mux.Handle(join(o, "/thumbnail"), image(Thumbnail))
	mux.Handle(join(o, "/zoom"), image(Zoom))
	mux.Handle(join(o, "/convert"), image(Convert))
	mux.Handle(join(o, "/watermark"), image(Watermark))
	mux.Handle(join(o, "/info"), image(Info))
	//kiem tra midleware cho nhung api khong co token
	o.ApiKey = ""
	image2 := ImageMiddleware(o)
	mux.Handle(join(o, "/profile"), image2(Profile))
	mux.Handle(join(o, "/fromexternalUpload"), image2(fromexternalUpload))

	return mux
}

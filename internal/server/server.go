package server

import (
	_ "embed"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/patrick246/partdb-csv/internal/auth"
	"github.com/patrick246/partdb-csv/internal/query"
	"log"
	"net/http"
	"strconv"
)

//go:embed index.html
var indexPage string

type Server struct {
	server      http.Server
	querier     *query.Querier
	linkBaseUrl string
	auth        auth.Authenticator
}

func NewServer(
	port uint,
	baseUrl string,
	querier *query.Querier,
	auth auth.Authenticator,
) *Server {
	mux := http.NewServeMux()

	srv := Server{
		server: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
		querier:     querier,
		linkBaseUrl: baseUrl,
		auth:        auth,
	}

	mux.HandleFunc("/parts.csv", srv.authenticationMiddleware(http.HandlerFunc(srv.handlePartsRequest)))
	mux.HandleFunc("/locations.csv", srv.authenticationMiddleware(http.HandlerFunc(srv.handleLocationRequest)))
	mux.HandleFunc("/", srv.authenticationMiddleware(http.HandlerFunc(srv.handleIndex)))
	return &srv
}

func (srv *Server) ListenAndServe() error {
	return srv.server.ListenAndServe()
}

func (srv *Server) authenticationMiddleware(next http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		username, password, ok := request.BasicAuth()
		if !ok {
			writer.Header().Set("WWW-Authenticate", `Basic realm=PartDB-CSV, charset="UTF-8"`)
			httpError(writer, http.StatusUnauthorized, "Authentication required")
			return
		}

		err := srv.auth.Authenticate(request.Context(), username, password)
		if errors.Is(err, auth.ErrUsernamePasswordMismatch) {
			writer.Header().Set("WWW-Authenticate", `Basic realm=PartDB-CSV, charset="UTF-8"`)
			httpError(writer, http.StatusUnauthorized, "Authentication required")
			return
		}
		if err != nil {
			log.Printf("authentication error, message=%v", err)
			httpError(writer, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		next.ServeHTTP(writer, request)
	}
}

func (srv *Server) handleIndex(writer http.ResponseWriter, _ *http.Request) {
	_, _ = writer.Write([]byte(indexPage))
}

func (srv *Server) handlePartsRequest(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		httpError(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	startId, err := strconv.ParseInt(request.URL.Query().Get("startID"), 10, 64)
	if err != nil {
		startId = 0
	}

	parts, err := srv.querier.GetPartData(request.Context(), startId)
	if err != nil {
		log.Printf("error getting data: %s", err)
		httpError(writer, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	writer.Header().Set("Content-Type", "text/csv")
	writer.Header().Set("Content-Disposition", "attachment; filename=parts.csv")
	csvWriter := csv.NewWriter(writer)
	err = csvWriter.Write([]string{
		"id",
		"name",
		"comment",
		"description",
		"instock",
		"Lagerplatz",
		"Link",
	})
	if err != nil {
		log.Printf("error writing csv: %v", err)
		return
	}
	for _, row := range parts {
		err = csvWriter.Write([]string{
			fmt.Sprintf("%d", row.ID),
			row.Name,
			row.Comment,
			row.Description,
			fmt.Sprintf("%d", row.Instock),
			row.Lagerplatz,
			fmt.Sprintf("%s/show_part_info.php?pid=%d", srv.linkBaseUrl, row.ID),
		})
		if err != nil {
			log.Printf("error writing csv: %v", err)
			return
		}
	}
	csvWriter.Flush()
}

func (srv *Server) handleLocationRequest(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		httpError(writer, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	startId, err := strconv.ParseInt(request.URL.Query().Get("startID"), 10, 64)
	if err != nil {
		startId = 0
	}

	locations, err := srv.querier.GetLocationData(request.Context(), startId)
	if err != nil {
		log.Printf("error getting data: %s", err)
		httpError(writer, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	writer.Header().Set("Content-Type", "text/csv")
	writer.Header().Set("Content-Disposition", "attachment; filename=locations.csv")
	csvWriter := csv.NewWriter(writer)
	err = csvWriter.Write([]string{
		"id",
		"name",
		"comment",
		"Lagerort",
		"Link",
	})
	if err != nil {
		log.Printf("error writing csv: %v", err)
		return
	}
	for _, row := range locations {
		err = csvWriter.Write([]string{
			fmt.Sprintf("%d", row.ID),
			row.Name,
			row.Comment,
			row.Lagerort,
			fmt.Sprintf("%s/show_location_parts.php?lid=%d", srv.linkBaseUrl, row.ID),
		})
		if err != nil {
			log.Printf("error writing csv: %v", err)
			return
		}
	}
	csvWriter.Flush()
}

func httpError(writer http.ResponseWriter, status int, message string) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(status)
	_, _ = writer.Write([]byte(message))
}

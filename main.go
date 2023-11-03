package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Album struct {
	AlbumId  int
	Title    string
	ArtistId int
}

var AlbumList []Album

func albumsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "/Users/sd-m004/Library/DBeaverData/workspace6/.metadata/sample-database-sqlite-1/Chinook.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		AlbumList := []Album{}
		sqlGet := "SELECT AlbumId, Title, ArtistId FROM Album"
		rows, err := db.Query(sqlGet)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var Title sql.NullString
			var AlbumId, ArtistId sql.NullInt16
			rows.Scan(&AlbumId, &Title, &ArtistId)
			AlbumList = append(AlbumList, Album{
				AlbumId:  int(AlbumId.Int16),
				Title:    Title.String,
				ArtistId: int(ArtistId.Int16),
			})
		}
		AlbumJSON, err := json.Marshal(AlbumList)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(AlbumJSON)
		return
	}
}

func findID(ID int) (*Album, int) {
	db, err := sql.Open("sqlite3", "/Users/sd-m004/Library/DBeaverData/workspace6/.metadata/sample-database-sqlite-1/Chinook.db")
	if err != nil {
		fmt.Println(err)
		return nil, 0
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Println("Database connection failed:", err)
		return nil, 0
	}
	sqlGet := "SELECT AlbumId, Title, ArtistId FROM Album"
	rows, err := db.Query(sqlGet)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var Title sql.NullString
		var AlbumId, ArtistId sql.NullInt16
		rows.Scan(&AlbumId, &Title, &ArtistId)
		AlbumList = append(AlbumList, Album{
			AlbumId:  int(AlbumId.Int16),
			Title:    Title.String,
			ArtistId: int(ArtistId.Int16),
		})
	}

	for i, album := range AlbumList {
		if album.AlbumId == ID {
			return &album, i
		}
	}
	return nil, 0
}
func albumHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "/Users/sd-m004/Library/DBeaverData/workspace6/.metadata/sample-database-sqlite-1/Chinook.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	urlPathSement := strings.Split(r.URL.Path, "Album/")
	ID, err := strconv.Atoi(urlPathSement[len(urlPathSement)-1])
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	album, _ := findID(ID)
	if album == nil {
		http.Error(w, fmt.Sprintf("no album with id %d", ID), http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		albumJSON, err := json.Marshal(album)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-type", "application/json")
		w.Write(albumJSON)
	case http.MethodPost:
		var newAlbum Album
		Bodybyte, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(Bodybyte, &newAlbum)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sqlGet := "INSERT INTO Album (AlbumId, Title, ArtistId) VALUES(?, ?, ?)"
		_, err = db.Exec(sqlGet, newAlbum.AlbumId, newAlbum.Title, newAlbum.ArtistId)
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusCreated)
		return
	case http.MethodPut:
		var updateAlbum Album
		byteBody, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(byteBody, &updateAlbum)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if updateAlbum.AlbumId != ID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sqlGet := `UPDATE Album
		SET Title=?, ArtistId=?
		WHERE AlbumId=?`

		_, err = db.Exec(sqlGet, updateAlbum.Title, updateAlbum.ArtistId, ID)

		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
		return
	case http.MethodDelete:
		var Deletelbum Album
		byteBody, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(byteBody, &Deletelbum)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if Deletelbum.AlbumId != ID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sqlDelete := `DELETE FROM Album
		WHERE AlbumId=?`

		_, err = db.Exec(sqlDelete, Deletelbum.AlbumId)

		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
		return
	
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func enableCorsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		handler.ServeHTTP(w, r)
	})
}

func main() {
	albumItemHandler := http.HandlerFunc(albumHandler)
	albumListHandler := http.HandlerFunc(albumsHandler)
	http.Handle("/Album/", enableCorsMiddleware(albumItemHandler))
	http.Handle("/Albums", enableCorsMiddleware(albumListHandler))
	http.ListenAndServe(":5000", nil)
}

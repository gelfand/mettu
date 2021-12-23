package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/gelfand/mettu/repo"
	"golang.org/x/net/context"
)

var (
	db *repo.DB

	rpcAddr = "ws://127.0.0.1:8545"

	userIDs = map[string]string{
		"75408122edc81988b92988054d2b4339f88e01d3efb7ec55cd275a558be71ac2": "Sats",
		"9b1e8a94fcdb88c8391ec1200718b3ddd73fb631b9c6b5d56619852a47833665": "Eugene",
		"0a041b9462caa4a31bac3567e0b6e6fd9100787db2ab433d96f6d178cabfce90": "user1",
		"6025d18fe48abd45168528f18a82e265dd98d421a7084aa09f61b341703901a3": "user2",
		"5860faf02b6bc6222ba5aca523560f0e364ccd8b67bee486fe8bf7c01d492ccb": "user3",
		"5a39bead318f306939acb1d016647be2e38c6501c58367fdb3e9f52542aa2442": "user4",
	}
	allow = map[string]string{
		"Sats":   "75408122edc81988b92988054d2b4339f88e01d3efb7ec55cd275a558be71ac2",
		"Eugene": "9b1e8a94fcdb88c8391ec1200718b3ddd73fb631b9c6b5d56619852a47833665",
		"user1":  "0a041b9462caa4a31bac3567e0b6e6fd9100787db2ab433d96f6d178cabfce90",
		"user2":  "6025d18fe48abd45168528f18a82e265dd98d421a7084aa09f61b341703901a3",
		"user3":  "5860faf02b6bc6222ba5aca523560f0e364ccd8b67bee486fe8bf7c01d492ccb",
		"user4":  "5269ef980de47819ba3d14340f4665262c41e933dc92c1a27dd5d01b047ac80e",
		"user5":  "5a39bead318f306939acb1d016647be2e38c6501c58367fdb3e9f52542aa2442",
	}
)

func hashSha(v string) []byte {
	sh := sha256.New224()
	_, _ = sh.Write([]byte(v))
	dat := sh.Sum(nil)
	return dat
}

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookies := r.Cookies()
		userID, hash := "", ""
		for i := range cookies {
			if cookies[i].Name == "Allowed" {
				hash = cookies[i].Value
			}
			if cookies[i].Name == "UserID" {
				userID = cookies[i].Value
			}
		}
		if hash != "" && userID != "" {
			pass, ok := allow[userID]
			if !ok {
				http.Handler(http.HandlerFunc(authHandler)).ServeHTTP(w, r)
				return
			}
			if bytes.Equal(hashSha(pass), hashSha(hash)) {
				http.Handler(http.HandlerFunc(authHandler)).ServeHTTP(w, r)
				return
			}
			if r.URL.Path == "/auth" {
				http.Handler(http.HandlerFunc(mainHandler)).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		queryVal, ok := r.URL.Query()["key"]
		if !ok || len(queryVal) == 0 {
			http.Handler(http.HandlerFunc(authHandler)).ServeHTTP(w, r)
			return
		}
		lenQueryVal := len(queryVal)
		if lenQueryVal == 0 {
			return
		}
		key := queryVal[lenQueryVal-1]
		if _, ok = userIDs[key]; !ok {
			return
		}
		fmt.Println(strings.ReplaceAll(key, "\n", ""))
		sh := sha256.New()
		sh.Write([]byte(key))
		head := sh.Sum([]byte{})
		fmt.Println(hash, key)
		http.SetCookie(w, &http.Cookie{
			Name:     "Allowed",
			Value:    fmt.Sprintf("%x", head),
			SameSite: http.SameSiteStrictMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "UserID",
			Value:    userIDs[key],
			SameSite: http.SameSiteStrictMode,
		})

		if r.URL.Path == "/auth" {
			http.Handler(http.HandlerFunc(mainHandler)).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		os.Exit(0)
	}()
	defer cancel()

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	db, err = repo.NewDB(homedir + "/.mettu/")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	go runCaching(ctx, db)

	mux := http.NewServeMux()
	mux.Handle("/static/css/", http.StripPrefix("/static/css/", http.FileServer(http.Dir("static/css"))))
	mux.Handle("/static/js/", http.StripPrefix("/static/js/", http.FileServer(http.Dir("static/js"))))
	mux.Handle("/auth", basicAuth(http.HandlerFunc(authHandler)))
	mux.Handle("/wallets", basicAuth(http.HandlerFunc(walletsHandler)))
	mux.Handle("/tokens", basicAuth(http.HandlerFunc(tokensHandler)))
	mux.Handle("/patterns", basicAuth(http.HandlerFunc(patternsHandler)))
	mux.Handle("/exchanges", basicAuth(http.HandlerFunc(exchangesHandler)))
	mux.Handle("/swaps", basicAuth(http.HandlerFunc(swapsHandler)))
	mux.Handle("/", basicAuth(http.HandlerFunc(mainHandler)))

	server := &http.Server{
		Addr:    "0.0.0.0:443",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServeTLS("certs/cdn.pem", "certs/cdn-key.pem"))
}

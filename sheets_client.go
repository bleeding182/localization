package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

type sheetData struct {
	sheet   *sheets.ValueRange
	sheetID string
}

var clientIDJson = []byte(`
{
	"installed": {
		"client_id":"142458847207-7hodvf2hl8ib6ga2smj0iveagbk23ih5.apps.googleusercontent.com",
		"project_id":"linear-elf-198922",
		"auth_uri":"https://accounts.google.com/o/oauth2/auth",
		"token_uri":"https://accounts.google.com/o/oauth2/token",
		"auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs",
		"client_secret":"1DGJhF7vtd-wgSA8BkG2d-nf",
		"redirect_uris":[
			"urn:ietf:wg:oauth:2.0:oob",
			"http://localhost"
			]
		}
	}
`)

func loadSpreadSheet() ([]*sheetData, error) {
	ctx := context.Background()

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/sheets.googleapis.com-go-quickstart.json
	config, err := google.ConfigFromJSON(clientIDJson, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
		panic(err)
	}
	client := getClient(ctx, config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
		panic(err)
	}

	sheetsInfo, err := srv.Spreadsheets.Get(*sheetID).Do()
	if err != nil {
		log.Fatalf("Unable to read sheets information: %v", err)
		panic(err)
	}

	results := make(chan *sheetData)
	for _, sheet := range sheetsInfo.Sheets {
		readRange := fmt.Sprintf("%v!A:M", sheet.Properties.Title)
		go func(readRange string, sheetId string) {
			res, err := srv.Spreadsheets.Values.Get(*sheetID, readRange).Do()
			if err != nil {
				panic(err)
			}
			results <- &sheetData{res, sheetId}
		}(readRange, strconv.FormatInt(sheet.Properties.SheetId, 10))
	}

	sheetData := make([]*sheetData, 0, len(sheetsInfo.Sheets))

	for r := range results {
		sheetData = append(sheetData, r)
		if len(sheetData) == cap(sheetData) {
			close(results)
		}
	}

	return sheetData, nil
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("sheets.googleapis.com-github-bleeding182-localization.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

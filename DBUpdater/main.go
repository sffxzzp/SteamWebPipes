package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	_ "modernc.org/sqlite"
)

type (
	SteamWebPipes struct {
		dbfile string
	}
	AppJson struct {
		AppList struct {
			Apps Apps `json:"apps"`
		} `json:"applist"`
	}
	Apps []App
	App  struct {
		AppID int    `json:"appid"`
		Name  string `json:"name"`
	}
	SubJson struct {
		Success     int    `json:"success"`
		ResultsHTML string `json:"results_html"`
		TotalCount  int    `json:"total_count"`
		Start       int    `json:"start"`
	}
	Subs []Sub
	Sub  struct {
		SubID int
		Name  string
	}
)

func newPipes(dbfile string) *SteamWebPipes {
	return &SteamWebPipes{
		dbfile: dbfile,
	}
}

func (s *SteamWebPipes) getAppData() *Apps {
	fmt.Println("Updating apps data...")
	url := "https://api.steampowered.com/ISteamApps/GetAppList/v0002/"
	var appdata AppJson
	json.Unmarshal(httpRead(httpGet(url)), &appdata)
	return &appdata.AppList.Apps
}

func (s *SteamWebPipes) updateApps() {
	db, err := sql.Open("sqlite", s.dbfile)
	isError(err)
	defer db.Close()
	_, err = db.Exec("DELETE FROM Apps")
	isError(err)
	_, err = db.Exec("PRAGMA synchronous = OFF")
	isError(err)
	data := s.getAppData()
	tx, err := db.Begin()
	isError(err)
	stmt, err := tx.Prepare("INSERT INTO Apps(AppID, Name, LastKnownName) VALUES (?, ?, ?)")
	isError(err)
	apps := make(map[int]string)
	for _, app := range *data {
		apps[app.AppID] = app.Name
	}
	for appid, name := range apps {
		if strings.TrimSpace(name) == "" {
			continue
		}
		_, err := stmt.Exec(appid, name, "")
		isError(err)
	}
	err = tx.Commit()
	isError(err)
}

func (s *SteamWebPipes) getSubsData() *Subs {
	fmt.Println("Updating subs data...")
	start := 0
	page := 999999
	var subs Subs
	for page > 0 {
		url := fmt.Sprintf("https://store.steampowered.com/search/results/?query&start=%d&count=100&dynamic_data=&sort_by=_ASC&category1=996&infinite=1&ignore_preferences=1", start)
		var subdata SubJson
		json.Unmarshal(httpRead(httpGet(url)), &subdata)
		fmt.Printf("%d / %d\n", start, subdata.TotalCount)
		page = int(math.Ceil(float64(subdata.TotalCount-start) / 100))
		start += 100
		doc := parseHTML(subdata.ResultsHTML)
		doc.Find("a").Each(func(_ int, node *goquery.Selection) {
			subid, exists := node.Attr("data-ds-packageid")
			if exists {
				title := node.Find("span.title").Text()
				subidint, err := strconv.Atoi(subid)
				isError(err)
				if strings.TrimSpace(title) != "" {
					subs = append(subs, Sub{
						SubID: subidint,
						Name:  title,
					})
				}
			}
		})
	}
	return &subs
}

func (s *SteamWebPipes) updateSubs() {
	db, err := sql.Open("sqlite", s.dbfile)
	isError(err)
	defer db.Close()
	_, err = db.Exec("DELETE FROM Subs")
	isError(err)
	_, err = db.Exec("PRAGMA synchronous = OFF")
	isError(err)
	data := s.getSubsData()
	tx, err := db.Begin()
	var rowStr []string
	var vals []interface{}
	isError(err)
	subs := make(map[int]string)
	for _, sub := range *data {
		subs[sub.SubID] = sub.Name
	}
	for subid, name := range subs {
		rowStr = append(rowStr, "(?, ?)")
		vals = append(vals, subid, name)
	}
	sqlStr := "INSERT INTO Subs (SubID, LastKnownName) VALUES " + strings.Join(rowStr, ", ")
	stmt, err := tx.Prepare(sqlStr)
	isError(err)
	_, err = stmt.Exec(vals...)
	isError(err)
	err = tx.Commit()
	isError(err)
}

func main() {
	pipes := newPipes("./database.db")
	pipes.updateApps()
	pipes.updateSubs()
}

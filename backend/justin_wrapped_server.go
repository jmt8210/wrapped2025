package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	YEAR = "2025"
)

type Track struct {
	Timestamp        string `json:"ts"`
	Platform         string `json:"platform"`
	MillisPlayed     int    `json:"ms_played"`
	Country          string `json:"conn_country"`
	IP               string `json:"ip_addr"`
	Name             string `json:"master_metadata_track_name"`
	AlbumArtist      string `json:"master_metadata_album_artist_name"`
	AlbumName        string `json:"master_metadata_album_album_name"`
	URI              string `json:"spotify_track_uri"`
	EpisodeName      string `json:"episode_name"`
	EpisodeShowName  string `json:"episode_show_name"`
	EpisodeURI       string `json:"spotify_episode_uri"`
	ReasonStart      string `json:"reason_start"`
	ReasonEnd        string `json:"reason_end"`
	Shuffle          bool   `json:"shuffle"`
	Skipped          bool   `json:"skipped"`
	Offline          bool   `json:"offline"`
	OfflineTimestamp int    `json:"offline_timestamp"`
	Incognito        bool   `json:"incognito_mode"`
}

type MinutesPerDay struct {
	Date  string
	Count int
}

type TopSong struct {
	Name         string
	Artist       string
	Plays        int
	Skips        int
	LengthMillis int
}

type TopArtist struct {
	Name         string
	Plays        int
	Skips        int
	LengthMillis int
}

type Stats struct {
	TotalTracks int
	MinsPerDay  []MinutesPerDay
	TopSongs    []TopSong
	TopArtists  []TopArtist
	MostSkipped TopSong
}

func main() {

	// Random port cuz why not
	PORT := 29228

	http.HandleFunc("/get-stats", statsHandler)

	// Start the server on port 8080
	fmt.Println("Starting server on :" + strconv.Itoa(PORT))
	if err := http.ListenAndServe(":"+strconv.Itoa(PORT), nil); err != nil {
		fmt.Println("Failed to serve on" + strconv.Itoa(PORT))

	}
}

func returnError(w http.ResponseWriter, e error) {
	fmt.Fprintf(w, `{"error": "%s"}`, e.Error())
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		r.ParseMultipartForm(10 << 20)

		// Stat lookup is disabled, consider adding db for cached lookup
		// user := r.FormValue("user")
		// if user == "" {
		// 	fmt.Println(`"user" field is empty but is required`)
		// 	returnError(w, errors.New(`"user" field is empty but is required`))
		// 	return
		// }

		// stats := lookupStats(user)
		// if stats != nil {
		// 	fmt.Printf("Found existing stats for user %s, returning\n", user)
		//
		// 	resJson, err := json.Marshal(stats)
		// 	if err != nil {
		// 		fmt.Printf("Error marshalling response json %s", err.Error())
		// 		returnError(w, err)
		// 		return
		// 	}
		//
		// 	fmt.Fprintf(w, string(resJson))
		// 	return
		// }

		file, _, err := r.FormFile("file")
		defer file.Close()
		if err != nil {
			fmt.Printf("Error reading file from multipart form data: %s\n", err.Error())
			returnError(w, err)
			return
		}

		buffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(buffer, file); err != nil {
			fmt.Printf("Couldn't copy file contents into buffer: %s\n", err.Error())
			returnError(w, err)
			return
		}

		reader := bytes.NewReader(buffer.Bytes())
		zipReader, err := zip.NewReader(reader, int64(len(buffer.Bytes())))
		if err != nil {
			fmt.Printf("Error creating zip reader: %s", err.Error())
			returnError(w, err)
			return
		}

		var files []fs.File

		for _, zippedFile := range zipReader.File {
			if strings.Contains(zippedFile.Name, YEAR) {
				f, err := zipReader.Open(zippedFile.Name)
				if err != nil {
					fmt.Printf("Error opening zipped file %s: %s", zippedFile.Name, err.Error())
					returnError(w, err)
					return
				}

				files = append(files, f)
			}
		}

		stats := processTracks(files)

		resJson, err := json.Marshal(stats)
		if err != nil {
			fmt.Printf("Error marshalling response json %s", err.Error())
			returnError(w, err)
			return
		}

		io.Writer.Write(w, resJson)

	default:
		fmt.Fprintf(w, "Invalid method")
	}
}

// Disable db functions
// func dbConnection() *sql.DB {
// 	db, err := sql.Open("sqlite3", "stats.db")
//
// 	if err != nil {
// 		fmt.Println()
// 	}
//
// 	return db
// }

// func lookupStats(user string) *Stats {
// 	result, err := dbConnection().Query("select * from stats where user = $1", user)
// 	if err != nil {
// 		fmt.Printf("Couldn't get stats for user %s: %s\n", user, err.Error())
// 	}
//
// 	if result.Next() {
// 		var stats Stats
// 		err := result.Scan(&stats.User, &stats.TotalTracks)
// 		if err != nil {
// 			fmt.Printf("Couldn't parse stats for user %s\n", user)
// 		}
// 		return &stats
// 	}
//
// 	return nil
// }

// func insertStats(stats *Stats) {
// 	result, err := dbConnection().Exec("insert into stats values ($1, $2)", stats.User, stats.TotalTracks)
// 	if err != nil {
// 		fmt.Printf("Couldn't insert stats for user %s: %s\n", stats.User, err.Error())
// 	}
//
// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		fmt.Printf("Couldn't get rows affected: %s\n", err.Error())
// 	}
//
// 	if rowsAffected != 1 {
// 		fmt.Printf("Wrong number of rows affected")
// 	}
// }

func processTracks(files []fs.File) *Stats {

	var tracks []Track

	for _, file := range files {
		data, err := io.ReadAll(file)

		fmt.Println(file)

		if err != nil {
			fmt.Printf("Error reading file %s: %s\n", file, err.Error())
		}

		var currentTracks []Track

		err = json.Unmarshal(data, &currentTracks)

		if err != nil {
			fmt.Printf("Error unmarshalling track data %s\n", err.Error())
		} else {
			// Filter out tracks with no name
			var filteredTracks []Track
			for _, track := range currentTracks {
				if track.Name != "" {
					filteredTracks = append(filteredTracks, track)
				}
			}
			tracks = append(tracks, filteredTracks...)
		}
	}

	topSongs := topSongs(tracks)
	topArtists := topArtists(tracks)

	return &Stats{
		TotalTracks: len(tracks),
		MinsPerDay:  processMinsPerDay(tracks),
		TopSongs:    topSongs[:5],
		TopArtists:  topArtists[:5],
		MostSkipped: mostSkipped(topSongs),
	}
}

func processMinsPerDay(tracks []Track) []MinutesPerDay {
	dayMap := make(map[string]MinutesPerDay)
	for _, track := range tracks {
		t, err := time.Parse(time.RFC3339, track.Timestamp)
		if err != nil {
			fmt.Printf("Couldn't parse time %s", track.Timestamp)
		}
		_, month, date := t.Date()
		dateString := fmt.Sprintf("%d-%d", month, date)
		if _, ok := dayMap[dateString]; !ok {
			dayMap[dateString] = MinutesPerDay{
				Date:  dateString,
				Count: track.MillisPlayed / 60000,
			}
		} else {
			dayMap[dateString] = MinutesPerDay{
				Date:  dateString,
				Count: dayMap[dateString].Count + track.MillisPlayed/60000,
			}
		}
	}

	start, err := time.Parse(time.RFC3339, fmt.Sprintf("%s-01-01T00:00:00.000Z", YEAR))
	if err != nil {
		fmt.Printf("Error parsing first day of %s (THIS SHOULD NOT HAPPEN): %s\n", YEAR, err.Error())
	}
	end := start.AddDate(1, 0, 0).AddDate(0, -1, 0)
	var days []MinutesPerDay
	for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
		_, month, date := d.Date()
		dateString := fmt.Sprintf("%d-%d", month, date)

		// Zero listening day, input zero
		if _, ok := dayMap[dateString]; !ok {
			dayMap[dateString] = MinutesPerDay{
				Date:  dateString,
				Count: 0,
			}
		}

		days = append(days, dayMap[dateString])
	}
	return days
}

func topSongs(tracks []Track) []TopSong {
	songMap := make(map[string]TopSong)
	for _, track := range tracks {
		skippedAdd := 0
		if track.Skipped {
			skippedAdd = 1
		}
		if _, ok := songMap[track.Name]; !ok {
			songMap[track.Name] = TopSong{
				Name:         track.Name,
				Artist:       track.AlbumArtist,
				Plays:        1,
				LengthMillis: track.MillisPlayed,
				Skips:        skippedAdd,
			}
		} else {
			prevSongStats := songMap[track.Name]
			songMap[track.Name] = TopSong{
				Name:         track.Name,
				Artist:       track.AlbumArtist,
				Plays:        prevSongStats.Plays + 1,
				LengthMillis: prevSongStats.LengthMillis + track.MillisPlayed,
				Skips:        prevSongStats.Skips + skippedAdd,
			}
		}
	}

	// Collect map values into slice and sort by # plays
	songs := slices.Collect(maps.Values(songMap))
	sort.Slice(songs, func(i, j int) bool {
		return songs[i].Plays > songs[j].Plays
	})

	return songs

}

func topArtists(tracks []Track) []TopArtist {
	artistMap := make(map[string]TopArtist)
	for _, track := range tracks {
		skippedAdd := 0
		if track.Skipped {
			skippedAdd = 1
		}
		if _, ok := artistMap[track.AlbumArtist]; !ok {
			artistMap[track.AlbumArtist] = TopArtist{
				Name:         track.AlbumArtist,
				Plays:        1,
				LengthMillis: track.MillisPlayed,
				Skips:        skippedAdd,
			}
		} else {
			prevArtistStats := artistMap[track.AlbumArtist]
			artistMap[track.AlbumArtist] = TopArtist{
				Name:         track.AlbumArtist,
				Plays:        prevArtistStats.Plays + 1,
				LengthMillis: prevArtistStats.LengthMillis + track.MillisPlayed,
				Skips:        prevArtistStats.Skips + skippedAdd,
			}
		}
	}

	// Collect map values into slice and sort by # plays
	artists := slices.Collect(maps.Values(artistMap))
	sort.Slice(artists, func(i, j int) bool {
		return artists[i].Plays > artists[j].Plays
	})

	return artists
}

func mostSkipped(topSongs []TopSong) TopSong {
	var mostSkippedSong TopSong
	mostSkippedCount := 0
	for _, song := range topSongs {
		if song.Skips > mostSkippedCount {
			mostSkippedSong = song
			mostSkippedCount = song.Skips
		}
	}

	return mostSkippedSong
}

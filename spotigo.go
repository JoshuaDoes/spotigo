package spotigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

type Client struct {
	Host string
	Pass string
}

type Track struct {
	Artist      string
	ArtURL      string
	Description string
	StreamURL   string
	Title       string
	TrackID     string
	Duration    int
}

type SpotigoTrackInfo struct {
	Name        string           `json:"name"`        //Track name
	TrackNumber int              `json:"number"`      //Track number
	DiscNumber  int              `json:"disc_number"` //Disc number
	Duration    int              `json:"duration"`    //Track duration in milliseconds
	Album       SpotigoAlbumInfo `json:"album"`       //Album
}
type SpotigoTrackOEmbed struct {
	ThumbnailURL string `json:"thumbnail_url"` //Thumbnail
}
type SpotigoAlbumInfo struct {
	Name   string              `json:"name"`   //Album name
	Artist []SpotigoArtistInfo `json:"artist"` //Artist list (main, ft.)
	Label  string              `json:"label"`  //Album record label
	Date   SpotigoDate         `json:"date"`   //Album release date
}
type SpotigoArtistInfo struct {
	Name string `json:"name"` //Artist name
}
type SpotigoDate struct {
	Year  int
	Month int
	Day   int
}

type SpotigoSearch struct {
	Results SpotigoSearchResults `json:"results"` //Search results
}
type SpotigoSearchResults struct {
	Tracks SpotigoSearchTracks `json:"tracks"` //Tracks
}
type SpotigoSearchTracks struct {
	Hits []SpotigoSearchHit `json:"hits"` //Hits
}
type SpotigoSearchHit struct {
	Album    SpotigoSearchHitAlbum    `json:"album"`
	Artists  []SpotigoSearchHitArtist `json:"artists"`
	ImageURL string                   `json:"image"`
	Name     string                   `json:"name"`
	URI      string                   `json:"uri"`
	ID       string                   `json:"-"` //Track ID, album ID, etc
	Duration int                      `json:"duration"`
}
type SpotigoSearchHitAlbum struct {
	//Artists //Unknown data type
	ImageURL string `json:"image"`
	Name     string `json:"name"`
	URI      string `json:"uri"`
}
type SpotigoSearchHitArtist struct {
	ImageURL string `json:"image"`
	Name     string `json:"name"`
	URI      string `json:"uri"`
}

func (c *Client) SearchTracks(query string) (*SpotigoSearchTracks, error) {
	query = url.QueryEscape(query)

	searchJSON, err := http.Get(fmt.Sprintf("http://%s/search/?query=%s&pass=%s", c.Host, query, c.Pass))
	if err != nil {
		return nil, err
	}

	trackResults := &SpotigoSearchTracks{}

	err = unmarshal(searchJSON, trackResults)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(trackResults.Hits); i++ {
		regex := regexp.MustCompile("^(spotify:)(.*)(:)([a-zA-Z0-9]+)(.*)$")
		id := regex.FindStringSubmatch(trackResults.Hits[i].URI)
		trackResults.Hits[i].ID = id[len(id)-2]
	}

	return trackResults, nil
}

func (c *Client) GetTrackInfo(url string) (*Track, error) {
	regex := regexp.MustCompile("^(https:\\/\\/open.spotify.com\\/track\\/|spotify:track:)([a-zA-Z0-9]+)(.*)$")
	trackID := regex.FindStringSubmatch(url)
	if len(trackID) <= 0 {
		return nil, errors.New("error finding track ID")
	}

	data := &Track{}
	data.TrackID = trackID[len(trackID)-2]

	trackJSON, err := http.Get(fmt.Sprintf("http://%s/track/%s?pass=%s", c.Host, data.TrackID, c.Pass))
	if err != nil {
		return data, err
	}
	trackOEmbedJSON, err := http.Get(fmt.Sprintf("https://embed.spotify.com/oembed?url=spotify:track:%s", data.TrackID))
	if err != nil {
		return data, err
	}

	trackInfo := &SpotigoTrackInfo{}
	err = unmarshal(trackJSON, trackInfo)
	if err != nil {
		return data, fmt.Errorf("%v", trackInfo)
	}
	if trackInfo.Name == "" {
		return data, errors.New("error getting track info")
	}

	trackOEmbed := &SpotigoTrackOEmbed{}
	err = unmarshal(trackOEmbedJSON, trackOEmbed)
	if err != nil {
		return data, fmt.Errorf("%v", trackOEmbedJSON)
	}
	if trackOEmbed.ThumbnailURL == "" {
		return data, errors.New("error getting track thumbnail")
	}

	data.Artist = trackInfo.Album.Artist[0].Name
	if len(trackInfo.Album.Artist) > 1 {
		data.Artist += " ft. " + trackInfo.Album.Artist[1].Name
		if len(trackInfo.Album.Artist) > 2 {
			for _, artist := range trackInfo.Album.Artist[2:] {
				data.Artist += ", " + artist.Name
			}
		}
	}
	data.StreamURL = fmt.Sprintf("http://%s/download/%s?pass=%s", c.Host, data.TrackID, c.Pass)
	data.Title = trackInfo.Name
	data.Duration = trackInfo.Duration
	data.ArtURL = trackOEmbed.ThumbnailURL

	return data, nil
}

// Gets the json from the API and assigns the data to the target.
// The target being a QueryResult struct
func unmarshal(body *http.Response, target interface{}) error {
	defer body.Body.Close()
	return json.NewDecoder(body.Body).Decode(target)
}

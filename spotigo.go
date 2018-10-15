package spotigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	Duration    int              `json:"duration"`    //Track duration
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

func (c *Client) GetTrackInfo(url string) (*Track, error) {
	regex := regexp.MustCompile("^(https:\\/\\/open.spotify.com\\/track\\/|spotify:track:)([a-zA-Z0-9]+)(.*)$")
	trackID := regex.FindStringSubmatch(url)
	if len(trackID) <= 0 {
		return nil, errors.New("error finding track ID")
	}

	trackJSON, err := http.Get(fmt.Sprintf("http://%s/track/%s?pass=%s", c.Host, trackID[len(trackID)-2], c.Pass))
	if err != nil {
		return nil, err
	}
	trackOEmbedJSON, err := http.Get(fmt.Sprintf("https://embed.spotify.com/oembed?url=spotify:track:%s", trackID[len(trackID)-2]))
	if err != nil {
		return nil, err
	}

	trackInfo := &SpotigoTrackInfo{}
	err = unmarshal(trackJSON, trackInfo)
	if err != nil {
		return nil, errors.New("error unmarshalling track info")
	}
	if trackInfo.Name == "" {
		return nil, errors.New("error getting track info")
	}

	trackOEmbed := &SpotigoTrackOEmbed{}
	err = unmarshal(trackOEmbedJSON, trackOEmbed)
	if err != nil {
		return nil, fmt.Errorf("%v", trackOEmbedJSON)
	}
	if trackOEmbed.ThumbnailURL == "" {
		return nil, errors.New("error getting track thumbnail")
	}

	data := &Track{}
	data.Artist = trackInfo.Album.Artist[0].Name
	if len(trackInfo.Album.Artist) > 1 {
		data.Artist += " ft. " + trackInfo.Album.Artist[1].Name
		if len(trackInfo.Album.Artist) > 2 {
			for _, artist := range trackInfo.Album.Artist[2:] {
				data.Artist += ", " + artist.Name
			}
		}
	}
	data.StreamURL = fmt.Sprintf("http://%s/download/%s?pass=%s", c.Host, trackID[len(trackID)-2], c.Pass)
	data.Title = trackInfo.Name
	data.TrackID = trackID[len(trackID)-2]
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

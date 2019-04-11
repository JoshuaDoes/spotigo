package spotigo

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Client struct {
	Host string
	Pass string
}

type Track struct {
	Artist    string
	Artists   []Artist
	Title     string
	Duration  int
	StreamURL string
	ArtURL    string
	TrackID   string
	URI       string
}

type Artist struct {
	Name      string
	TopTracks []*Track
	Albums    []*Album
	Singles   []*Album
	ArtURL    string
	ArtistID  string
	URI       string
}

type Album struct {
	Title   string
	Artist  string
	Artists []Artist
	Discs   []*Disc
	ArtURL  string
	AlbumID string
	URI     string
}

type Disc struct {
	Number int
	Tracks []*Track
}

type SpotigoTrackInfo struct {
	Gid         string              `json:"gid"`
	Name        string              `json:"name"`        //Track name
	TrackNumber int                 `json:"number"`      //Track number
	DiscNumber  int                 `json:"disc_number"` //Disc number
	Duration    int                 `json:"duration"`    //Track duration in milliseconds
	Album       SpotigoAlbumInfo    `json:"album"`       //Album
	Artist      []SpotigoArtistInfo `json:"artist"`      //Artist
}
type SpotigoEmbed struct {
	ThumbnailURL string `json:"thumbnail_url"` //Thumbnail
}
type SpotigoAlbumInfo struct {
	Gid    string              `json:"gid"`
	Name   string              `json:"name"`   //Album name
	Artist []SpotigoArtistInfo `json:"artist"` //Artist list (main, ft.)
	Discs  []SpotigoDisc       `json:"disc"`   //Virtual CD list
	Date   SpotigoDate         `json:"date"`   //Album release date
}
type SpotigoDisc struct {
	Number int          `json:"number"` //Virtual CD number
	Tracks []SpotigoGid `json:"track"`  //Virtual CD track list
}
type SpotigoArtistInfo struct {
	Gid       string             `json:"gid"`
	Name      string             `json:"name"`         //Artist name
	TopTracks []SpotigoTopTracks `json:"top_track"`    //Top tracks
	Albums    []SpotigoAlbums    `json:"album_group"`  //Albums
	Singles   []SpotigoAlbums    `json:"single_group"` //Single tracks inside albums
}
type SpotigoTopTracks struct {
	Tracks []SpotigoGid `json:"track"`
}
type SpotigoAlbums struct {
	Albums []SpotigoGid `json:"album"`
}
type SpotigoDate struct {
	Year  int
	Month int
	Day   int
}

type SpotigoGid struct {
	Gid string //Spotify GID
	ID  string
}

func (gid *SpotigoGid) GetID() (string, error) {
	str, err := base64.StdEncoding.DecodeString(gid.Gid)
	if err != nil {
		return "", err
	}

	id := ConvertTo62(str)

	return id, nil
}

type SpotigoPlaylist struct {
	Gid        string                    `json:"gid"`
	Length     int                       `json:"length"`
	Attributes SpotigoPlaylistAttributes `json:"attributes"`
	Contents   SpotigoPlaylistContents   `json:"contents"`

	//Additional data added by Spotigo
	UserID      string `json:"-"`
	PlaylistID  string `json:"-"`
	PlaylistURI string `json:"-"`
	ImageURL    string `json:"-"`
}
type SpotigoPlaylistAttributes struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type SpotigoPlaylistContents struct {
	Position  int                   `json:"pos"`
	Truncated bool                  `json:"truncated"`
	Items     []SpotigoPlaylistItem `json:"items"`
}
type SpotigoPlaylistItem struct {
	TrackURI   string                        `json:"uri"`
	Attributes SpotigoPlaylistItemAttributes `json:"attributes"`
}
type SpotigoPlaylistItemAttributes struct {
	AddedBy   string `json:"added_by"`
	Timestamp int64  `json:"timestamp"`
}

type SpotigoSearch struct {
	Results SpotigoSearchResults `json:"results"` //Search results
}
type SpotigoSearchResults struct {
	Tracks    SpotigoSearchHits `json:"tracks"`    //Tracks
	Albums    SpotigoSearchHits `json:"albums"`    //Albums
	Artists   SpotigoSearchHits `json:"artists"`   //Artists
	Playlists SpotigoSearchHits `json:"playlists"` //Playlists
}
type SpotigoSearchHits struct {
	Hits []SpotigoSearchHit `json:"hits"` //Hits
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

type SpotigoSearchHit struct {
	Album          SpotigoSearchHitAlbum    `json:"album"`
	Artists        []SpotigoSearchHitArtist `json:"artists"`
	ImageURL       string                   `json:"image"`
	Name           string                   `json:"name"`
	URI            string                   `json:"uri"`
	ID             string                   `json:"-"` //Track ID, album ID, etc
	Duration       int                      `json:"duration"`
	FollowersCount int                      `json:"followersCount"`
	Author         string                   `json:"author"`
}

func (hit *SpotigoSearchHit) GetType() string {
	regex := regexp.MustCompile("^spotify:(track|artist|album|user):([a-zA-Z0-9]+).*$")
	hitURI := regex.FindStringSubmatch(hit.URI)
	if len(hitURI) <= 0 {
		return ""
	}

	return hitURI[1]
}

func (hit *SpotigoSearchHit) GetID() []string {
	switch hit.GetType() {
	case "track", "artist", "album":
		regex := regexp.MustCompile("^spotify:(track|artist|album):([a-zA-Z0-9]+)$")
		hitURI := regex.FindStringSubmatch(hit.URI)
		if len(hitURI) <= 0 {
			return make([]string, 0)
		}
		return []string{hitURI[2]}
	case "user":
		regex := regexp.MustCompile("^spotify:user:(\\w\\S+):playlist:(\\w\\S+)$")
		hitURI := regex.FindStringSubmatch(hit.URI)
		if len(hitURI) <= 0 {
			return make([]string, 0)
		}
		user, _ := url.QueryUnescape(hitURI[1])
		return []string{user, hitURI[2]}
	}
	return make([]string, 0)
}

func (c *Client) Search(query string) (*SpotigoSearch, error) {
	query = url.QueryEscape(query)

	searchJSON, err := http.Get(fmt.Sprintf("http://%s/search/?query=%s&pass=%s", c.Host, query, c.Pass))
	if err != nil {
		return nil, err
	}

	searchResults := &SpotigoSearch{}

	err = unmarshal(searchJSON, searchResults)
	if err != nil {
		return nil, err
	}

	return searchResults, nil
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
		return data, err
	}
	if trackInfo.Name == "" {
		return data, errors.New("error getting track info")
	}

	trackOEmbed := &SpotigoEmbed{}
	err = unmarshal(trackOEmbedJSON, trackOEmbed)
	if err != nil {
		return data, err
	}
	if trackOEmbed.ThumbnailURL == "" {
		return data, errors.New("error getting track thumbnail")
	}

	data.Artist = trackInfo.Artist[0].Name
	if len(trackInfo.Artist) > 1 {
		data.Artist += " ft. " + trackInfo.Artist[1].Name
		if len(trackInfo.Artist) > 2 {
			for _, artist := range trackInfo.Artist[2:] {
				data.Artist += ", " + artist.Name
			}
		}
	}
	for _, trackArtist := range trackInfo.Artist {
		artistGid := &SpotigoGid{Gid: trackArtist.Gid}
		artistID, err := artistGid.GetID()
		if err != nil {
			continue
		}
		artist, err := c.GetArtistInfo("spotify:artist:" + artistID)
		if err == nil {
			data.Artists = append(data.Artists, *artist)
		}
	}
	data.StreamURL = fmt.Sprintf("http://%s/download/%s?pass=%s", c.Host, data.TrackID, c.Pass)
	data.Title = trackInfo.Name
	data.Duration = trackInfo.Duration
	data.ArtURL = trackOEmbed.ThumbnailURL

	return data, nil
}

func (c *Client) GetArtistInfo(url string) (*Artist, error) {
	regex := regexp.MustCompile("^(https:\\/\\/open.spotify.com\\/artist\\/|spotify:artist:)([a-zA-Z0-9]+)(.*)$")
	artistID := regex.FindStringSubmatch(url)
	if len(artistID) <= 0 {
		return nil, errors.New("error finding artist ID")
	}

	data := &Artist{}
	data.ArtistID = artistID[len(artistID)-2]

	artistJSON, err := http.Get(fmt.Sprintf("http://%s/artist/%s?pass=%s", c.Host, data.ArtistID, c.Pass))
	if err != nil {
		return data, err
	}
	artistOEmbedJSON, err := http.Get(fmt.Sprintf("https://embed.spotify.com/oembed?url=spotify:artist:%s", data.ArtistID))
	if err != nil {
		return data, err
	}

	artistInfo := &SpotigoArtistInfo{}
	err = unmarshal(artistJSON, artistInfo)
	if err != nil {
		return data, err
	}
	if artistInfo.Name == "" {
		return data, errors.New("error getting artist info")
	}

	artistOEmbed := &SpotigoEmbed{}
	err = unmarshal(artistOEmbedJSON, artistOEmbed)
	if err != nil {
		return data, err
	}
	if artistOEmbed.ThumbnailURL == "" {
		return data, errors.New("error getting artist thumbnail")
	}

	data.Name = artistInfo.Name
	if len(artistInfo.TopTracks) > 0 {
		for _, topTrack := range artistInfo.TopTracks[0].Tracks {
			trackID, err := topTrack.GetID()
			if err != nil {
				continue
			}
			/*trackInfo, err := c.GetTrackInfo("spotify:track:" + trackID)
			if err == nil {
				data.TopTracks = append(data.TopTracks, trackInfo)
			}*/
			data.TopTracks = append(data.TopTracks, &Track{TrackID: trackID, URI: "spotify:track:" + trackID})
		}
	}
	/*for _, albumGroup := range artistInfo.Albums {
		for _, album := range albumGroup.Albums {
			albumID, err := album.GetID()
			if err != nil {
				continue
			}
			albumInfo, err := c.GetAlbumInfo("spotify:album:" + albumID)
			if err == nil {
				data.Albums = append(data.Albums, albumInfo)
			}
		}
	}
	for _, singleGroup := range artistInfo.Singles {
		for _, album := range singleGroup.Albums {
			albumID, err := album.GetID()
			if err != nil {
				continue
			}
			albumInfo, err := c.GetAlbumInfo("spotify:album:" + albumID)
			if err == nil {
				data.Albums = append(data.Albums, albumInfo)
			}
		}
	}*/
	data.ArtURL = artistOEmbed.ThumbnailURL

	return data, nil
}

func (c *Client) GetAlbumInfo(url string) (*Album, error) {
	regex := regexp.MustCompile("^(https:\\/\\/open.spotify.com\\/album\\/|spotify:album:)([a-zA-Z0-9]+)(.*)$")
	albumID := regex.FindStringSubmatch(url)
	if len(albumID) <= 0 {
		return nil, errors.New("error finding album ID")
	}

	data := &Album{}
	data.AlbumID = albumID[len(albumID)-2]

	albumJSON, err := http.Get(fmt.Sprintf("http://%s/album/%s?pass=%s", c.Host, data.AlbumID, c.Pass))
	if err != nil {
		return data, err
	}
	albumOEmbedJSON, err := http.Get(fmt.Sprintf("https://embed.spotify.com/oembed?url=spotify:album:%s", data.AlbumID))
	if err != nil {
		return data, err
	}

	albumInfo := &SpotigoAlbumInfo{}
	err = unmarshal(albumJSON, albumInfo)
	if err != nil {
		return data, err
	}
	if albumInfo.Name == "" {
		return data, errors.New("error getting album info")
	}

	albumOEmbed := &SpotigoEmbed{}
	err = unmarshal(albumOEmbedJSON, albumOEmbed)
	if err != nil {
		return data, err
	}
	if albumOEmbed.ThumbnailURL == "" {
		return data, errors.New("error getting album thumbnail")
	}

	data.Title = albumInfo.Name
	if len(albumInfo.Artist) > 1 {
		data.Artist += " ft. " + albumInfo.Artist[1].Name
		if len(albumInfo.Artist) > 2 {
			for _, artist := range albumInfo.Artist[2:] {
				data.Artist += ", " + artist.Name
			}
		}
	}
	for _, albumArtist := range albumInfo.Artist {
		artistGid := &SpotigoGid{Gid: albumArtist.Gid}
		artistID, err := artistGid.GetID()
		if err != nil {
			continue
		}
		artist, err := c.GetArtistInfo("spotify:artist:" + artistID)
		if err == nil {
			data.Artists = append(data.Artists, *artist)
		}
	}
	for discN, albumDisc := range albumInfo.Discs {
		disc := &Disc{Number: discN}
		for _, track := range albumDisc.Tracks {
			trackID, err := track.GetID()
			if err != nil {
				continue
			}
			/*trackInfo, err := c.GetTrackInfo("spotify:track:" + trackID)
			if err == nil {
				disc.Tracks = append(disc.Tracks, trackInfo)
			}*/
			disc.Tracks = append(disc.Tracks, &Track{TrackID: trackID, URI: "spotify:track:" + trackID})
		}
		data.Discs = append(data.Discs, disc)
	}
	data.ArtURL = albumOEmbed.ThumbnailURL

	return data, nil
}

func (c *Client) GetPlaylist(url string) (*SpotigoPlaylist, error) {
	regex := regexp.MustCompile("^(https:\\/\\/open.spotify.com\\/user\\/|spotify:user:)(\\w\\S+)(\\/playlist\\/|:playlist:)(\\w\\S+)(.*)$")
	matches := regex.FindStringSubmatch(url)
	if len(matches) <= 0 {
		return nil, errors.New("error finding playlist/user ID")
	}

	userID := matches[len(matches)-4]
	playlistID := matches[len(matches)-2]

	data := &SpotigoPlaylist{}
	data.UserID = userID
	data.PlaylistID = playlistID

	playlistJSON, err := http.Get(fmt.Sprintf("http://%s/playlist/spotify:user:%s:playlist:%s?pass=%s", c.Host, userID, playlistID, c.Pass))
	if err != nil {
		return data, err
	}

	err = unmarshal(playlistJSON, data)
	if err != nil {
		return data, err
	}

	return data, nil
}

// Gets the json from the API and assigns the data to the target.
// The target being a QueryResult struct
func unmarshal(body *http.Response, target interface{}) error {
	defer body.Body.Close()
	return json.NewDecoder(body.Body).Decode(target)
}

func ConvertTo62(raw []byte) string {
	bi := big.Int{}
	bi.SetBytes(raw)
	rem := big.NewInt(0)
	base := big.NewInt(62)
	zero := big.NewInt(0)
	result := ""

	for bi.Cmp(zero) > 0 {
		_, rem = bi.DivMod(&bi, base, rem)
		result += string(alphabet[int(rem.Uint64())])
	}

	for len(result) < 22 {
		result += "0"
	}
	return reverse(result)
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

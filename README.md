# spotigo
Single file library for accessing SpotigoWeb (currently not publicly accessible)

# Installing
`go get github.com/JoshuaDoes/spotigo`

# Example
```go
package main

import "fmt"
import "github.com/JoshuaDoes/spotigo"

func main() {
	//Initialize a new client
	c := &spotigo.Client{Host: "spotigo host here", pass: "spotigo pass key here"}

	//Get metadata and an audio URL
	res, err := c.GetTrackInfo("https://open.spotify.com/track/1Czngy8R5LkQPs3CCkuKjF")

	if err != nil {
		panic(err)
	}
	
	fmt.Println("Track ID: " + res.TrackID)
	fmt.Println("Title: " + res.Title)
	fmt.Println("Artist(s): " + res.Artist)
	fmt.Println("Stream URL: " + res.StreamURL)
	fmt.Println("Artwork URL: " + res.ArtURL)
}
```
### Output

```
> go run main.go

< Track ID
< Title
< Artist(s)
< Stream URL
< Artwork URL
```

## License
The source code for spotigo is released under the MIT License. See LICENSE for more details.

## Donations
All donations are appreciated and helps me stay awake at night to work on this more. Even if it's not much, it helps a lot in the long run!
You can find the donation link here: [Donation Link](https://paypal.me/JoshuaDoes)
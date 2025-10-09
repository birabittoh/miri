# miri

An easy-to-use library to download music from Deezer, heavily ~stolen~ inspired by [GoDeez](https://github.com/mathismqn/godeez).

## Example usage

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/birabittoh/miri"
)

func main() {
	ctx := context.Background()

	arlCookie := os.Getenv("ARL_COOKIE")
	secretKey := os.Getenv("SECRET_KEY")

	opt := miri.SearchOptions{Limit: 1, Query: "eminem"}
	res, err := miri.SearchTracks(ctx, opt)
	if err != nil {
		log.Fatalf("failed to search tracks: %v", err)
	}
	if len(res) == 0 {
		log.Fatal("no tracks found")
	}

	track := res[0]

	cfg, err := miri.NewConfig(arlCookie, secretKey)
	if err != nil {
		log.Fatalf("failed to create config: %v", err)
	}

	c, err := miri.New(ctx, cfg) // new miri client
	if err != nil {
		log.Fatalf("failed to create Miri client: %v", err)
	}

	data, err := c.DownloadTrackByID(ctx, track.ID)
	if err != nil {
		log.Fatalf("failed to download track: %v", err)
	}

	if len(data) == 0 {
		log.Fatal("downloaded data is empty")
	}

	coverURL := track.CoverURL("xl")
	println(coverURL)
}
```

## Variables

Here are the key variables you need to set in your config object:

1. `ArlCookie`
* **What is it?**: The `arl_cookie` is a session cookie used for authentication with Deezer. Without this cookie, the downloader cannot access your account to retrieve playlists, albums, or songs.
* **How to retrieve it**:
	1.	Open your browser and log in to your Deezer account.
	2.	Open the Developer Tools (right-click on the page and select “Inspect” or press F12).
	3.	Navigate to the Application tab (in Chrome/Edge) or Storage tab (in Firefox).
	4.	In the left panel, look for Cookies and select `https://www.deezer.com`.
	5.	Find the arl cookie and copy its value.

2. `SecretKey`
* **What is it?**: The `secret_key` is a cryptographic value used to decrypt Deezer’s media files.
* **Where to find it?**: While we cannot provide the specific secret_key in this documentation, it can be found online through various sources or developer communities that focus on Deezer.

## License

miri is provided under the MIT License.

---

> ⚠️ This tool is provided for educational and personal use only. Please ensure your usage complies with Deezer’s Terms of Service.

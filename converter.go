package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/egonik-unlp/spotify-to-youtube/spotify-to-youtube/spotify_auth"
	"github.com/zmb3/spotify"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/youtube/v3"
)

var (
	method = flag.String("method", "list", "The API method to execute. (List is the only method that this sample currently supports.")

	channelId              = flag.String("channelId", "", "Retrieve playlists for this channel. Value is a YouTube channel ID.")
	hl                     = flag.String("hl", "", "Retrieve localized resource metadata for the specified application language.")
	maxResults             = flag.Int64("maxResults", 5, "The maximum number of playlist resources to include in the API response.")
	mine                   = flag.Bool("mine", true, "List playlists for authenticated user's channel. Default: false.")
	onBehalfOfContentOwner = flag.String("onBehalfOfContentOwner", "", "Indicates that the request's auth credentials identify a user authorized to act on behalf of the specified content owner.")
	pageToken              = flag.String("pageToken", "", "Token that identifies a specific page in the result set that should be returned.")
	part                   = flag.String("part", "snippet", "Comma-separated list of playlist resource parts that API response will include.")
	playlistId             = flag.String("playlistId", "", "Retrieve information about this playlist.")
)

type Snippet struct {
	title       string `json:"title"`
	description string `json: "description"`
}

type reqString struct {
	snippet Snippet `json:"snippet"`
}

func make_playlist_string(title, description string) string {
	snippet := Snippet{title: title, description: description}
	reqstring := reqString{snippet: snippet}
	bytes, err := json.Marshal(reqstring)
	if err != nil {
		panic("problemas con el jsonstirng")
	}
	return string(bytes)

}

func main() {
	flag.Parse()
	client := getClient(youtube.YoutubeForceSslScope)
	service, err := youtube.New(client)
	youtube.NewPlaylistsService(service)

	if err != nil {
		log.Fatalf("ERROR %v", err)
	}
	// redirectURL := "http://localhost:8888/callback"
	spotifyClient := spotify_auth.Authenticate()
	playlistID := "4RNxYgx8c1WuDV7MItXel2"
	playlist, err := spotifyClient.GetPlaylist(spotify.ID(playlistID))
	if err != nil {
		fmt.Println("cant get playlist")
	}
	fmt.Println(playlist.Name)
	tracks, err := spotifyClient.GetPlaylistTracks(spotify.ID(playlistID))
	if err != nil {
		log.Fatalf("error obteniendo tracks %v", err)
	}
	convertedTracks := make([]string, 0)
	fulltrackdata := make([]spotify.PlaylistTrack, 0)
	// spotifyClient.NextPage()
	for page := 1; ; page++ {

		if err != nil {
			log.Fatal(err)
		}
		for _, track := range tracks.Tracks {
			tpl := fmt.Sprintf("%s %s", track.Track.Name, track.Track.Album.Name)
			for _, artistName := range track.Track.Album.Artists {
				tpl = fmt.Sprintf("%s %s", tpl, artistName.Name)
			}
			// fmt.Println(tpl)
			// tpl := track.Track.Name
			// fmt.Println(tpl)/
			// fmt.Println(track.Track.Album.Name)
			// fmt.Println(track.Track.Album.Artists[0].Name)
			convertedTracks = append(convertedTracks, tpl)
			fulltrackdata = append(fulltrackdata, track)
		}
		err = spotifyClient.NextPage(tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
	}
	fmt.Println(convertedTracks)
	short_filename := fmt.Sprintf("binary_%s_data.bin", playlist.Name)
	trackfile, err := os.Create(short_filename)
	if err != nil {
		fmt.Printf("error creando el archivo de salida simple, %v", err)
	}
	defer func() {
		if err := trackfile.Close(); err != nil {
			panic("AT THE DISCO")
		}
	}()
	long_filename := fmt.Sprintf("binary_%s_data_long.bin", playlist.Name)
	fulltrackFIle, err := os.Create(long_filename)
	if err != nil {
		fmt.Printf("error creando el archivo de salida completo, %v", err)
	}
	defer func() {
		if err := fulltrackFIle.Close(); err != nil {
			panic("AT THE DISCO")
		}
	}()
	var b bytes.Buffer
	var bfull bytes.Buffer
	encoder := gob.NewEncoder(&b)
	fullEncoder := gob.NewEncoder(&bfull)
	if err := encoder.Encode(convertedTracks); err != nil {
		panic("cant't dump")
	}
	if err := fullEncoder.Encode(fulltrackdata); err != nil {
		panic("cant't dump")
	}
	trackfile.Write(b.Bytes())
	fulltrackFIle.Write(bfull.Bytes())
	// youtubeService, err := youtube.NewService(ctx, option.WithCredentialsFile(credentialsFilePath))
	// if err != nil {
	// 	fmt.Printf("%s", err)
	// }
	items := make([]youtube.PlaylistItem, 0)
	newPlaylistStub := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       playlist.Name,
			Description: "Convertida desde spotify",
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "public", // Set to "public" if you want it to be public
		},
	}
	// fmt.Println("ESTOY ACA")
	// results := make([]string)
	playlistInsertCall := service.Playlists.Insert([]string{"snippet, status"}, newPlaylistStub)
	newPlaylist, err := playlistInsertCall.Do()
	if err != nil {
		if e, ok := err.(*googleapi.Error); ok {
			if e.Code == 409 {
				fmt.Println("ya equiste bro")
			}
		} else {
			log.Fatalf("LAAACONN")
		}
	}

	for n, song := range convertedTracks {
		if n == 40 {
			break
		}
		searchCall := service.Search.List([]string{"id,snippet"}).Q(song)
		response, err := searchCall.Do()
		if err != nil {
			log.Fatalf("EN LA Bu %s", err)
			break
		}
		cntr := 0
		for _, item := range response.Items {
			if item.Id.Kind == "youtube#video" {
				// 			// fmt.Println(item.Snippet.Title)
				if cntr == 0 {
					insertItem := youtube.PlaylistItem{
						Snippet: &youtube.PlaylistItemSnippet{
							PlaylistId: newPlaylist.Id,
							ResourceId: item.Id,
						},
					}
					songInsertcall := service.PlaylistItems.Insert([]string{"snippet"}, &insertItem)
					fmt.Println(item)
					_, err := songInsertcall.Do()
					// items = append(items, newPlaylistItem)
					if err != nil {
						if e, ok := err.(*googleapi.Error); ok {
							if e.Code == 409 {
								fmt.Printf("No se pudo insertar %v", insertItem.Snippet.ResourceId)
							}
						}
						log.Fatalf("no se pudo insertar %v", err)
					}

					// results = append(results, item.Snippet.Title)
				}
				cntr += 1
			}

			// 		// fmt.Printf("%v%v", n, item.Id.)
		}
		if err != nil {
			log.Fatalf("ERROR SEARCH %v", err)
		}
	}
	// fmt.Println(results)
	// fmt.Println(response)
	// playlistInsertCall := service.Playlists.Insert([]string{"snippet,status"}, playlist)

	// createdPlaylist, err := playlistInsertCall.Do()
	// service.Playlists.
	// 	service.PlaylistItems.Insert("id")
	fmt.Println(items)
}

// func search_for_songs(song string) {

// }

package spotify_auth

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/zmb3/spotify"
)

func Authenticate() spotify.Client {
	fmt.Println("Iniciando auth en spotify")
	redirectURL := "http://localhost:8888"
	auth := spotify.NewAuthenticator(redirectURL, spotify.ScopeUserReadPrivate)
	auth.SetAuthInfo("e4c1abc2559b40378248b3e2b89dc8f1", "ffe46f9b6f174dd2a28f4d67139bb48d")
	responseW, err := webserver()
	if err != nil {
		log.Fatalf("Error obtenindo token, %v", err)
	}
	url := auth.AuthURL(redirectURL)
	// url = fmt.Sprintf("%s/callback", url)
	err = openURL(url)
	if err != nil {
		log.Fatalf("Unable to open authorization URL in web server: %v", err)
	} else {
		fmt.Println("Your browser has been opened to an authorization URL.",
			" This program will resume once authorization has been provided.\n")
		fmt.Println(redirectURL)
	}

	response := <-responseW
	fmt.Println(response.FormValue("code"))
	token, err := auth.Token(redirectURL, response)
	if err != nil {
		log.Fatalf("error obteniendo token  del canal, %v", err)
	}
	return auth.NewClient(token)
}

func webserver() (chan *http.Request, error) {
	fmt.Println("Iniciando ws")
	listener, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		fmt.Println("ERROR GENERANDO LISTENER")
		return nil, err
	}
	fmt.Println("response")
	codeCh := make(chan *http.Request)
	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// code := r.FormValue("code")
		codeCh <- r // send code to OAuth flow
		listener.Close()
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "Received code: %v \r\nYou can now safely close this browser window.", r)
	}))
	return codeCh, nil
}

func openURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:4001/").Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("Cannot open URL %s on this platform", url)
	}
	return err
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/PuerkitoBio/goquery"
	"github.com/olekukonko/tablewriter"
)

const baseUrl = "https://www.megacartoons.net/help-wanted/"

var (
	play                = flag.Int("p", -1, "play the wanted episode without any user interaction")
	list                = flag.Bool("l", false, "list episodes and quit")
	videoPlayer         = flag.String("vp", "mpv", "use another video player [default=mpv]")
	download            = flag.Int("d", -1, "download all episodes asynchronously but max [d] episodes at a time")
	listFavourites      = flag.Bool("lf", false, "list favourite episodes")
	addFavouriteEpisode = flag.Int("af", 0, "add an episode to your favourites")
	delFavouriteEpisode = flag.Int("rf", 0, "remove an episode from your favourites")
)

func getEpisodes() ([]string, []string) {
	resp, err := http.Get(baseUrl)
	if err != nil {
		fmt.Printf("Error while trying to connect to the website, error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error while trying to connect to the website, error code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error while loading the website content: %v\n", err)
		os.Exit(1)
	}

	episodesUrls := []string{}
	episodesTitles := []string{}

	doc.Find("a.btn.btn-sm.btn-default").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			episodesUrls = append(episodesUrls, href)
		}

		title, exists := s.Attr("title")
		if exists {
			episodesTitles = append(episodesTitles, title)
		}
	})

	return episodesUrls, episodesTitles
}

func listEpisodes(episodesTitles []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Episode", "Number"})
	table.SetAutoWrapText(false)

	for i, title := range episodesTitles {
		table.Append([]string{title, fmt.Sprintf("%d", i+1)})
	}

	table.Render()
}

var favouriteEpisodes = make(map[int]string)

func favouriteEpisode(episodeNumber int) {
	if episodeNumber == 0 || episodeNumber > 340 {
		fmt.Println("Enter a valid episode number")
		return
	}

	_, episodeTitles := getEpisodes()
	favouriteEpisodes[episodeNumber] = episodeTitles[episodeNumber-1]

	// Open the file in read-write mode
	file, err := os.OpenFile("favourites.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Decode existing data from the file
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&favouriteEpisodes); err != nil && err != io.EOF {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	// Move to the beginning of the file to overwrite its contents
	if _, err := file.Seek(0, 0); err != nil {
		fmt.Println("Error seeking file:", err)
		return
	}

	// Write the updated map back to the file
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(favouriteEpisodes); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
}

func listFavouriteEpisode() {
	file, err := os.Open("favourites.json")
	if err != nil {
		fmt.Println("Error opening favorites file", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&favouriteEpisodes)
	if err != nil {
		fmt.Println("Error decoding file", err)
	}
	fmt.Println("Your Favourite Episodes")
	for key, value := range favouriteEpisodes {
		fmt.Printf("Episode %d : %s \n", key, value)
	}

}

var favouriteEpisodesJson = make(map[int]string)

func unFavouriteEpisode(episodeNumber int) {

	file, err := os.OpenFile("favourites.json", os.O_RDWR|os.O_CREATE, 0644)
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&favouriteEpisodesJson); err != nil && err != io.EOF {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	delete(favouriteEpisodesJson, episodeNumber)
	if err := file.Truncate(0); err != nil {
		fmt.Println("Error truncating file:", err)
		return
	}
	if _, err := file.Seek(0, 0); err != nil {
		fmt.Println("Error seeking file:", err)
		return
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(favouriteEpisodesJson)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
	}
}

func extractVideo(source string) string {
	resp, err := http.Get(source)
	if err != nil {
		fmt.Printf("Error while trying to get the video source: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error while loading the video page: %v\n", err)
		os.Exit(1)
	}

	input := doc.Find("input[name='main_video_url']")
	videoSource, exists := input.Attr("value")
	if !exists {
		fmt.Println("Error: Could not find the video source.")
		os.Exit(1)
	}

	return videoSource
}

func playVideo(video, player string) {
	cmd := exec.Command(player, video)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error while playing the video: %v\n", err)
		os.Exit(1)
	}
}

func userInput(episodesUrls []string) int {
	fmt.Print("Which episode do you want to watch? ")
	var user int
	_, err := fmt.Scanf("%d", &user)
	if err != nil || (user < 1 || user > len(episodesUrls)) {
		fmt.Println("Invalid input, please try again.")
		return userInput(episodesUrls)
	}
	return user
}

func main() {
	flag.Parse()

	episodesUrls, episodesTitles := getEpisodes()
	if len(os.Args[1:]) == 0 {
		listEpisodes(episodesTitles)
		user := userInput(episodesUrls)
		video := extractVideo(episodesUrls[user-1])
		fmt.Printf("Playing '%s'...\n", episodesTitles[user-1])
		playVideo(video, *videoPlayer)
	} else {
		if *download > 0 {
			if err := downloadAllEpisodes(*download); err != nil {
				fmt.Printf("Error while download all episodes: %v\n", err)
			}
			return
		}

		if *play >= 1 {
			video := extractVideo(episodesUrls[*play-1])
			fmt.Printf("Playing '%s'...\n", episodesTitles[*play])
			playVideo(video, *videoPlayer)
		} else if *list {
			listEpisodes(episodesTitles)
		}
		if *listFavourites {
			listFavouriteEpisode()
			return
		}
		if *addFavouriteEpisode != 0 {
			favouriteEpisode(*addFavouriteEpisode)
		}
		if *delFavouriteEpisode != 0 {
			unFavouriteEpisode(*delFavouriteEpisode)
		}
	}
}

package main

import (
	"fmt"
	"github.com/google/go-github/github"
	"io"
	"log"
	"math"
	"os"
	"time"
)

const (
	REMAINING_THRESHOLD = 1
)

func main() {

	t := &github.UnauthenticatedRateLimitedTransport{
		ClientID:     "a4157b9e1e84e52658c9",
		ClientSecret: "0f7c2313ced247debae6009605380a95ee73ff6c",
	}
	client := github.NewClient(t.Client())

	fmt.Println("Repos that contain magento and PHP code.")

	page := 1
	maxPage := math.MaxInt32

	query := fmt.Sprintf("magento+language:php+page:10")

	opts := &github.SearchOptions{
		Sort:  "updated",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	filename := "/tmp/repos_locations"

	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}

	for page <= maxPage {
		opts.Page = page
		result, response, err := client.Search.Repositories(query, opts)
		Wait(response)

		if err != nil {
			log.Fatal("FindRepos:", err)
		}

		maxPage = response.LastPage

		msg := fmt.Sprintf("page: %v/%v, size: %v, total: %v",
			page, maxPage, len(result.Repositories), *result.Total)
		log.Println(msg)

		for _, repo := range result.Repositories {

			repo_name := *repo.FullName
			username := *repo.Owner.Login

			fmt.Println("repo: ", repo_name)
			fmt.Println("owner: ", username)

			user, response, err := client.Users.Get(username)
			Wait(response)

			if err != nil {
				fmt.Println(err)
			} else {

				if user.Location != nil {

					user_location := *user.Location

					n, err := io.WriteString(f, "\""+username+"\",\""+user_location+"\",\""+repo_name+"\"\n")
					if err != nil {
						fmt.Println(n, err)
					}

				} else {

					user_location := "not found"

					n, err := io.WriteString(f, "\""+username+"\",\""+user_location+"\",\""+repo_name+"\"\n")
					if err != nil {
						fmt.Println(n, err)
					}

				}

			}

			time.Sleep(time.Millisecond * 2500)

		}

		page++

	}

	f.Close()

}

func Wait(response *github.Response) {
	if response != nil && response.Remaining <= REMAINING_THRESHOLD {
		gap := time.Duration(response.Reset.Local().Unix() - time.Now().Unix())
		sleep := gap * time.Second
		if sleep < 0 {
			sleep = -sleep
		}

		time.Sleep(sleep)
	}
}

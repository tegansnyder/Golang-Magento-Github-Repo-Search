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
		ClientID:     "SOME_CLIENT_ID",
		ClientSecret: "SOME_CLIENT_SECRET",
	}
	client := github.NewClient(t.Client())

	fmt.Println("Repos that contain magento and PHP code.")

	// create a file to be used for geocoder
	filename := "/tmp/locations.txt"

	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}

	// slice the queries into batches to get around the API limit of 1000

	queries := []string{"\"2008-06-01 .. 2012-09-01\"", "\"2008-06-01 .. 2012-09-01\"", "\"2012-09-02 .. 2013-04-20\"", "\"2013-04-21 .. 2013-10-20\"", "\"2013-10-21 .. 2014-03-10\"", "\"2014-03-10 .. 2014-07-10\"", "\"2014-07-10 .. 2014-09-30\""}

	for _, q := range queries {

		query := fmt.Sprintf("magento language:PHP created:" + q)

		page := 1
		maxPage := math.MaxInt32

		opts := &github.SearchOptions{
			Sort:  "updated",
			Order: "desc",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
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
				created_at := repo.CreatedAt.String()

				fmt.Println("repo: ", repo_name)
				fmt.Println("owner: ", username)
				fmt.Println("created_at: ", created_at)

				user, response, err := client.Users.Get(username)
				Wait(response)

				if err != nil {
					fmt.Println(err)
				} else {

					if user.Location != nil {

						user_location := *user.Location

						n, err := io.WriteString(f, "\""+username+"\",\""+user_location+"\",\""+repo_name+"\",\""+created_at+"\"\n")
						if err != nil {
							fmt.Println(n, err)
						}

					} else {

						user_location := "not found"

						n, err := io.WriteString(f, "\""+username+"\",\""+user_location+"\",\""+repo_name+"\",\""+created_at+"\"\n")
						if err != nil {
							fmt.Println(n, err)
						}

					}

				}

				time.Sleep(time.Millisecond * 500)

			}

			page++

		}

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

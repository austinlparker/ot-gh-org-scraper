package main

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/shurcooL/githubv4"
	"github.com/subosito/gotenv"
	"golang.org/x/oauth2"
)

func main() {
	gotenv.Load()

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)

	type repo struct {
		Node struct {
			Name            string
			Description     string
			URL             string
			PrimaryLanguage struct {
				Name string
			}
			LicenseInfo struct {
				Name string
			}
		}
	}
	var query struct {
		Organization struct {
			Repositories struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Edges []repo
			} `graphql:"repositories(first: 50, after: $repoCursor)"`
		} `graphql:"organization(login: \"opentracing-contrib\")"`
	}

	variables := map[string]interface{}{
		"repoCursor": (*githubv4.String)(nil),
	}

	var allRepoNodes []repo

	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		allRepoNodes = append(allRepoNodes, query.Organization.Repositories.Edges...)
		if !query.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["repoCursor"] = githubv4.NewString(query.Organization.Repositories.PageInfo.EndCursor)
	}

	templatePath := []string{"contrib.tmpl"}

	for _, value := range allRepoNodes {
		t := template.Must(template.ParseFiles(templatePath...))
		f, _ := os.Create(fmt.Sprintf("./out/%s.md", value.Node.Name))
		err := t.Execute(f, value)
		if err != nil {
			panic(err)
		}
		f.Close()
	}
}

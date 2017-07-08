package main

import (
	"github.com/go-siris/siris"
	"github.com/go-siris/siris/_examples/best_practice/view/quicktemplate/templates"
	"github.com/go-siris/siris/context"
)

func main() {
	// Replace `"github.com/**/templates"` import with your own project's reference.
	// You're free to use any package-name and embed subpackages too if you wish.

	// Run `qtc` command that comes with Quicktemplate when you update
	// your `.qtpl` files and it will gogenerate .go buffer files for you.
	// The `qtc` command will automatically look into subfolders.

	// Just use `qtc && go run main.go` for quick re-runs.

	app := siris.New()

	app.Get("/", func(c context.Context) {
		isLoggedIn := true // Logic here to determine if a user's logged in.

		templates.WritePageTemplate(c, &templates.HomePage{
			BasePage: templates.BasePage{
				IsLoggedIn: isLoggedIn,
			},
			Context: c,
		})
	})

	app.Run(siris.Addr(":8080"))
}

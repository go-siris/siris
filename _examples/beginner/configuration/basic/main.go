package main

import (
	"github.com/go-siris/siris"
)

func main() {
	app := siris.New()

	// [...]

	// Good when you want to modify the whole configuration.
	app.Run(siris.Addr(":8080"), siris.WithConfiguration(siris.Configuration{ // default configuration:
		DisableBanner:                     false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		TimeFormat:                        "Mon, 02 Jan 2006 15:04:05 GMT",
		Charset:                           "UTF-8",
	}))

	// or before run:
	// app.Configure(siris.WithConfiguration(...))
	// app.Run(siris.Addr(":8080"))
}

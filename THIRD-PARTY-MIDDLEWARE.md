# Third Party Middleware

Siris has its own middleware form of `func(ctx context.Context)` but it's also compatible with all `net/http` middleware forms. See [here](https://github.com/go-siris/siris/tree/master/_examples/beginner/convert-handlers).

I'm sure that each of you have, already, found his own favorite list but here's a small list of third-party handlers:

| Middleware | Author | Description |
| -----------|--------|-------------|
| [tollbooth](https://github.com/didip/tollbooth) | [Didip Kerabat](https://github.com/didip) | Generic middleware to rate-limit HTTP requests. [Example](https://github.com/didip/tollbooth/pull/34) |
| [goth](https://github.com/markbates/goth) | [Mark Bates](https://github.com/markbates) | OAuth, OAuth2 authentication. [Example](https://github.com/go-siris/siris/tree/master/_examples/intermediate/oauth2) |
| [binding](https://github.com/mholt/binding) | [Matt Holt](https://github.com/mholt) | Data binding from HTTP requests into structs |
| [cloudwatch](https://github.com/cvillecsteele/negroni-cloudwatch) | [Colin Steele](https://github.com/cvillecsteele) | AWS cloudwatch metrics middleware |
| [csp](https://github.com/awakenetworks/csp) | [Awake Networks](https://github.com/awakenetworks) | [Content Security Policy](https://www.w3.org/TR/CSP2/) (CSP) support |
| [delay](https://github.com/jeffbmartinez/delay) | [Jeff Martinez](https://github.com/jeffbmartinez) | Add delays/latency to endpoints. Useful when testing effects of high latency |
| [New Relic Go Agent](https://github.com/yadvendar/negroni-newrelic-go-agent) | [Yadvendar Champawat](https://github.com/yadvendar) | Official [New Relic Go Agent](https://github.com/newrelic/go-agent) (currently in beta)  |
| [gorelic](https://github.com/jingweno/negroni-gorelic) | [Jingwen Owen Ou](https://github.com/jingweno) | New Relic agent for Go runtime |
| [JWT](https://github.com/go-siris/middleware/tree/master/jwt) | [Auth0](https://github.com/auth0) | Middleware checks for a JWT on the `Authorization` header on incoming requests and decodes it. [Example](https://github.com/go-siris/middleware/tree/master/jwt/_example)|
| [logrus](https://github.com/meatballhat/negroni-logrus) | [Dan Buch](https://github.com/meatballhat) | Logrus-based logger |
| [onthefly](https://github.com/xyproto/onthefly) | [Alexander Rødseth](https://github.com/xyproto) | Generate TinySVG, HTML and CSS on the fly |
| [permissions2](https://github.com/xyproto/permissions2) | [Alexander Rødseth](https://github.com/xyproto) | Cookies, users and permissions |
| [prometheus](https://github.com/zbindenren/negroni-prometheus) | [Rene Zbinden](https://github.com/zbindenren) | Easily create metrics endpoint for the [prometheus](http://prometheus.io) instrumentation tool |
| [render](https://github.com/unrolled/render) | [Cory Jacobsen](https://github.com/unrolled) | Render JSON, XML and HTML templates |
| [RestGate](https://github.com/pjebs/restgate) | [Prasanga Siripala](https://github.com/pjebs) | Secure authentication for REST API endpoints |
| [secure](https://github.com/go-siris/middleware/tree/master/secure) | [Cory Jacobsen](https://github.com/unrolled) | Middleware that implements a few quick security wins |
| [stats](https://github.com/thoas/stats) | [Florent Messa](https://github.com/thoas) | Store information about your web application (response time, etc.) |
| [VanGoH](https://github.com/auroratechnologies/vangoh) | [Taylor Wrobel](https://github.com/twrobel3) | Configurable [AWS-Style](http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html) HMAC authentication middleware |
| [xrequestid](https://github.com/pilu/xrequestid) | [Andrea Franz](https://github.com/pilu) | Middleware that assigns a random X-Request-Id header to each request |
| [digits](https://github.com/bamarni/digits) | [Bilal Amarni](https://github.com/bamarni) | Middleware that handles [Twitter Digits](https://get.digits.com/) authentication |

Feel free to put up a [PR](https://github.com/go-siris/middleware) your middleware!

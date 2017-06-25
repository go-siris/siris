### Reload on source code changes

```sh
$ go get -u github.com/kataras/rizla
$ cd $GOPATH/src/mywebapp
$ rizla main.go
```

## Performance optimization tips for multi-core systems

### enable Reuseport feature:
```GO
app.Run(siris.Addr(":8080"), siris.EnableReuseport)
// or before run:
app.Configure(siris.EnableReuseport)
```

- Enable the Reuseport feature like abdove
- Run a separate server instance per CPU core with GOMAXPROCS=1.
- Pin each server instance to a separate CPU core using [taskset](http://linux.die.net/man/1/taskset).
- Ensure the interrupts of multiqueue network card are evenly distributed between CPU cores. See [this article](https://blog.cloudflare.com/how-to-achieve-low-latency/) for details.

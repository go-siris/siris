package host

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-siris/siris/configuration"
	"github.com/go-siris/siris/core/errors"
	"github.com/go-siris/siris/core/nettools"
	"github.com/lucas-clemente/quic-go/h2quic"
	"golang.org/x/crypto/acme/autocert"
)

// Configurator provides an easy way to modify
// the Supervisor.
//
// Look the `Configure` func for more.
type Configurator func(su *Supervisor)

var (
	tlsNewTicketEvery = time.Hour * 6 // generate a new ticket for TLS PFS encryption every so often
	tlsNumTickets     = 5             // hold and consider that many tickets to decrypt TLS sessions
)

// Supervisor is the wrapper and the manager for a compatible server
// and it's relative actions, called Tasks.
//
// Interfaces are separated to return relative functionality to them.
type Supervisor struct {
	Server         *http.Server
	quicServer     *h2quic.Server
	config         *configuration.Configuration
	closedManually int32 // future use, accessed atomically (non-zero means we've called the Shutdown)
	manuallyTLS    bool  // we need that in order to determinate what to output on the console before the server begin.
	shouldWait     int32 // non-zero means that the host should wait for unblocking
	unblockChan    chan struct{}

	tlsGovChan chan struct{} // close to stop the TLS maintenance goroutine

	mu sync.Mutex

	onServe    []func(TaskHost)
	onErr      []func(error)
	onShutdown []func()
}

// New returns a new host supervisor
// based on a native net/http "srv".
//
// It contains all native net/http's Server methods.
// Plus you can add tasks on specific events.
// It has its own flow, which means that you can prevent
// to return and exit and restore the flow too.
func New(srv *http.Server, sirisConfig *configuration.Configuration) *Supervisor {
	return &Supervisor{
		Server:      srv,
		config:      sirisConfig,
		unblockChan: make(chan struct{}, 1),
	}
}

// Configure accepts one or more `Configurator`.
// With this function you can use simple functions
// that are spread across your app to modify
// the supervisor, these Configurators can be
// used on any Supervisor instance.
//
// Look `Configurator` too.
//
// Returns itself.
func (su *Supervisor) Configure(configurators ...Configurator) *Supervisor {
	for _, conf := range configurators {
		conf(su)
	}
	return su
}

// DeferFlow defers the flow of the exeuction,
// i.e: when server should return error and exit
// from app, a DeferFlow call inside a Task
// can wait for a `RestoreFlow` to exit or not exit if
// host's server is "fixed".
//
// See `RestoreFlow` too.
func (su *Supervisor) DeferFlow() {
	atomic.StoreInt32(&su.shouldWait, 1)
}

// RestoreFlow restores the flow of the execution,
// if called without a `DeferFlow` call before
// then it does nothing.
// See tests to understand how that can be useful on specific cases.
//
// See `DeferFlow` too.
func (su *Supervisor) RestoreFlow() {
	if su.isWaiting() {
		atomic.StoreInt32(&su.shouldWait, 0)
		su.mu.Lock()
		su.unblockChan <- struct{}{}
		su.mu.Unlock()
	}
}

func (su *Supervisor) isWaiting() bool {
	return atomic.LoadInt32(&su.shouldWait) != 0
}

func (su *Supervisor) newListener() (net.Listener, error) {
	// this will not work on "unix" as network
	// because UNIX doesn't supports the kind of
	// restarts we may want for the server.
	//
	// User still be able to call .Serve instead.
	l, err := nettools.TCPKeepAlive(su.Server.Addr, su.config.EnableReuseport)
	if err != nil {
		return nil, err
	}

	// here we can check for sure, without the need of the supervisor's `manuallyTLS` field.
	if nettools.IsTLS(su.Server) {
		// means tls
		tlsl := tls.NewListener(l, su.Server.TLSConfig)
		return tlsl, nil
	}

	return l, nil
}

// RegisterOnErrorHook registers a function to call when errors occured by the underline http server.
func (su *Supervisor) RegisterOnErrorHook(cb func(error)) {
	su.mu.Lock()
	su.onErr = append(su.onErr, cb)
	su.mu.Unlock()
}

func (su *Supervisor) notifyErr(err error) {
	// if err == http.ErrServerClosed {
	// 	su.notifyShutdown()
	// 	return
	// }

	su.mu.Lock()
	for _, f := range su.onErr {
		go f(err)
	}
	su.mu.Unlock()
}

// RegisterOnServeHook registers a function to call on
// Serve/ListenAndServe/ListenAndServeTLS/ListenAndServeAutoTLS.
func (su *Supervisor) RegisterOnServeHook(cb func(TaskHost)) {
	su.mu.Lock()
	su.onServe = append(su.onServe, cb)
	su.mu.Unlock()
}

func (su *Supervisor) notifyServe(host TaskHost) {
	su.mu.Lock()
	for _, f := range su.onServe {
		go f(host)
	}
	su.mu.Unlock()
}

// Remove all channels, do it with events
// or with channels but with a different channel on each task proc
// I don't know channels are not so safe, when go func and race risk..
// so better with callbacks....
func (su *Supervisor) supervise(blockFunc func() error) error {
	createdHost := createTaskHost(su)

	su.notifyServe(createdHost)

	tryStartInterruptNotifier()

	err := blockFunc()
	su.notifyErr(err)

	if su.isWaiting() {
	blockStatement:
		for {
			select {
			case <-su.unblockChan:
				break blockStatement
			}
		}
	}

	return err // start the server
}

// Serve accepts incoming connections on the Listener l, creating a
// new service goroutine for each. The service goroutines read requests and
// then call su.server.Handler to reply to them.
//
// For HTTP/2 support, server.TLSConfig should be initialized to the
// provided listener's TLS Config before calling Serve. If
// server.TLSConfig is non-nil and doesn't include the string "h2" in
// Config.NextProtos, HTTP/2 support is not enabled.
//
// Serve always returns a non-nil error. After Shutdown or Close, the
// returned error is http.ErrServerClosed.
func (su *Supervisor) Serve(l net.Listener) error {

	return su.supervise(func() error {
		hErr := make(chan error)
		qErr := make(chan error)
		go func() {
			hErr <- su.Server.Serve(l)
		}()

		if su.config.GetEnableQUICSupport() && su.quicServer != nil {
			// Open the listeners
			udpConn, err := su.ListenPacket()
			if err != nil {
				return err
			}
			go func() {
				qErr <- su.quicServer.Serve(udpConn)
			}()
		}

		select {
		case err := <-hErr:
			if su.config.GetEnableQUICSupport() && su.quicServer != nil {
				su.quicServer.Close()
			}
			return err
		case err := <-qErr:
			// Cannot close the HTTP server or wait for requests to complete properly :/
			return err
		}
	})
}

// ListenAndServe listens on the TCP network address addr
// and then calls Serve with handler to handle requests
// on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
func (su *Supervisor) ListenAndServe() error {
	l, err := su.newListener()
	if err != nil {
		return err
	}
	return su.Serve(l)
}

func setupHTTP2(cfg *tls.Config) {
	cfg.NextProtos = append(cfg.NextProtos, "h2") // HTTP2
}

// ListenAndServeTLS acts identically to ListenAndServe, except that it
// expects HTTPS connections. Additionally, files containing a certificate and
// matching private key for the server must be provided. If the certificate
// is signed by a certificate authority, the certFile should be the concatenation
// of the server's certificate, any intermediates, and the CA's certificate.
func (su *Supervisor) ListenAndServeTLS(certFile string, keyFile string) error {
	if certFile == "" || keyFile == "" {
		return errors.New("certFile or keyFile missing")
	}
	cfg := new(tls.Config)
	var err error
	cfg.Certificates = make([]tls.Certificate, 1)
	if cfg.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile); err != nil {
		return err
	}

	setupHTTP2(cfg)
	su.Server.TLSConfig = cfg
	su.manuallyTLS = true

	if su.config.GetEnableQUICSupport() {
		su.quicServer = &h2quic.Server{Server: su.Server}
		su.Server.Handler = su.wrapWithSvcHeaders(su.Server.Handler)
	}

	// Setup any goroutines governing over TLS settings
	su.tlsGovChan = make(chan struct{})
	timer := time.NewTicker(tlsNewTicketEvery)
	go runTLSTicketKeyRotation(su.Server.TLSConfig, timer, su.tlsGovChan)

	return su.ListenAndServe()
}

// ListenAndServeAutoTLS acts identically to ListenAndServe, except that it
// expects HTTPS connections. server's certificates are auto generated from LETSENCRYPT using
// the golang/x/net/autocert package.
func (su *Supervisor) ListenAndServeAutoTLS() error {
	autoTLSManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
	}

	cfg := new(tls.Config)
	cfg.GetCertificate = autoTLSManager.GetCertificate
	setupHTTP2(cfg)
	su.Server.TLSConfig = cfg
	su.manuallyTLS = true

	if su.config.GetEnableQUICSupport() {
		su.quicServer = &h2quic.Server{Server: su.Server}
		su.Server.Handler = su.wrapWithSvcHeaders(su.Server.Handler)
	}

	// Setup any goroutines governing over TLS settings
	su.tlsGovChan = make(chan struct{})
	timer := time.NewTicker(tlsNewTicketEvery)
	go runTLSTicketKeyRotation(su.Server.TLSConfig, timer, su.tlsGovChan)

	return su.ListenAndServe()
}

// RegisterOnShutdownHook registers a function to call on Shutdown.
// This can be used to gracefully shutdown connections that have
// undergone NPN/ALPN protocol upgrade or that have been hijacked.
// This function should start protocol-specific graceful shutdown,
// but should not wait for shutdown to complete.
func (su *Supervisor) RegisterOnShutdownHook(cb func()) {
	// when go1.9: replace the following lines with su.Server.RegisterOnShutdownHook(f)
	su.mu.Lock()
	su.onShutdown = append(su.onShutdown, cb)
	su.mu.Unlock()
}

func (su *Supervisor) notifyShutdown() {
	// when go1.9: remove the lines below
	su.mu.Lock()
	for _, f := range su.onShutdown {
		go f()
	}
	su.mu.Unlock()
	// end
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shut down.
// If the provided context expires before the shutdown is complete,
// then the context's error is returned.
//
// Shutdown does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of Shutdown should
// separately notify such long-lived connections of shutdown and wait
// for them to close, if desired.
func (su *Supervisor) Shutdown(ctx context.Context) error {
	atomic.AddInt32(&su.closedManually, 1) // future-use
	su.notifyShutdown()
	return su.Server.Shutdown(ctx)
}

var runTLSTicketKeyRotation = standaloneTLSTicketKeyRotation

var setSessionTicketKeysTestHook = func(keys [][32]byte) [][32]byte {
	return keys
}

// standaloneTLSTicketKeyRotation governs over the array of TLS ticket keys used to de/crypt TLS tickets.
// It periodically sets a new ticket key as the first one, used to encrypt (and decrypt),
// pushing any old ticket keys to the back, where they are considered for decryption only.
//
// Lack of entropy for the very first ticket key results in the feature being disabled (as does Go),
// later lack of entropy temporarily disables ticket key rotation.
// Old ticket keys are still phased out, though.
//
// Stops the timer when returning.
func standaloneTLSTicketKeyRotation(c *tls.Config, timer *time.Ticker, exitChan chan struct{}) {
	defer timer.Stop()
	// The entire page should be marked as sticky, but Go cannot do that
	// without resorting to syscall#Mlock. And, we don't have madvise (for NODUMP), too. ☹
	keys := make([][32]byte, 1, tlsNumTickets)

	rng := c.Rand
	if rng == nil {
		rng = rand.Reader
	}
	if _, err := io.ReadFull(rng, keys[0][:]); err != nil {
		c.SessionTicketsDisabled = true // bail if we don't have the entropy for the first one
		return
	}
	c.SessionTicketKey = keys[0] // SetSessionTicketKeys doesn't set a 'tls.keysAlreadSet'
	c.SetSessionTicketKeys(setSessionTicketKeysTestHook(keys))

	for {
		select {
		case _, isOpen := <-exitChan:
			if !isOpen {
				return
			}
		case <-timer.C:
			rng = c.Rand // could've changed since the start
			if rng == nil {
				rng = rand.Reader
			}
			var newTicketKey [32]byte
			_, err := io.ReadFull(rng, newTicketKey[:])

			if len(keys) < tlsNumTickets {
				keys = append(keys, keys[0]) // manipulates the internal length
			}
			for idx := len(keys) - 1; idx >= 1; idx-- {
				keys[idx] = keys[idx-1] // yes, this makes copies
			}

			if err == nil {
				keys[0] = newTicketKey
			}
			// pushes the last key out, doesn't matter that we don't have a new one
			c.SetSessionTicketKeys(setSessionTicketKeysTestHook(keys))
		}
	}
}

// ListenPacket creates udp connection for QUIC if it is enabled,
func (su *Supervisor) ListenPacket() (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", su.Server.Addr)
	if err != nil {
		return nil, err
	}
	return net.ListenUDP("udp", udpAddr)
}

func (su *Supervisor) wrapWithSvcHeaders(previousHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		su.quicServer.SetQuicHeaders(w.Header())
		previousHandler.ServeHTTP(w, r)
	}
}

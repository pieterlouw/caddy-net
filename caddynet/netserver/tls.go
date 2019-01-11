package netserver

import (
	"fmt"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddytls"
	"github.com/mholt/certmagic"
)

// activateTLS
func activateTLS(cctx caddy.Context) error {
	operatorPresent := !caddy.Started()

	// Follow steps stipulated in https://github.com/mholt/caddy/wiki/Writing-a-Plugin:-Server-Type#automatic-tls (indicated below by numbered comments)

	// 1. Prints a message to stdout, "Activating privacy features..." (if the operator is present; i.e. caddy.Started() == false) because the process can take a few seconds
	if !caddy.Quiet && operatorPresent {
		fmt.Print("Activating privacy features...")
	}

	ctx := cctx.(*netContext)

	// 2. Sets the Managed field to true on all configs that should be fully managed
	for _, cfg := range ctx.configs {
		if caddytls.QualifiesForManagedTLS(cfg) {
			cfg.TLS.Managed = true
		}
	}

	// 3. Calls ObtainCert() for each config (this method only obtains certificates if the config qualifies and has its Managed field set to true).
	// place certificates and keys on disk
	for _, c := range ctx.configs {
		err := c.TLS.Manager.ObtainCert(c.TLS.Hostname, operatorPresent)
		if err != nil {
			return err
		}

	}

	// 4. Configures the server struct to use the newly-obtained certificates by setting the Enabled field of the TLS config to true
	// and calling caddytls.CacheManagedCertificate() which actually loads the cert into memory for use
	for _, cfg := range ctx.configs {
		if cfg == nil || cfg.TLS == nil || !cfg.TLS.Managed {
			continue
		}
		cfg.TLS.Enabled = true
		if certmagic.HostQualifies(cfg.Hostname) {
			_, err := cfg.TLS.Manager.CacheManagedCertificate(cfg.Hostname)
			if err != nil {
				return err
			}
		}

		// 5. Calls caddytls.SetDefaultTLSParams() to make sure all the necessary fields have a value
		// Make sure any config values not explicitly set are set to default
		caddytls.SetDefaultTLSParams(cfg.TLS)

	}

	// 6. Calls caddytls.RenewManagedCertificates(true) to ensure that all certificates that were loaded into memory have been renewed if necessary
	// renew all relevant certificates that need renewal. this is important
	// to do right away so we guarantee that renewals aren't missed, and
	// also the user can respond to any potential errors that occur.

	// renew all relevant certificates that need renewal. this is important
	// to do right away so we guarantee that renewals aren't missed, and
	// also the user can respond to any potential errors that occur.
	// (skip if upgrading, because the parent process is likely already listening
	// on the ports we'd need to do ACME before we finish starting; parent process
	// already running renewal ticker, so renewal won't be missed anyway.)
	if !caddy.IsUpgrade() {
		ctx.instance.StorageMu.RLock()
		certCache, ok := ctx.instance.Storage[caddytls.CertCacheInstStorageKey].(*certmagic.Cache)
		ctx.instance.StorageMu.RUnlock()
		if ok && certCache != nil {
			if err := certCache.RenewManagedCertificates(operatorPresent); err != nil {
				return err
			}
		}
	}

	if !caddy.Quiet && operatorPresent {
		fmt.Println(" done.")
	}

	return nil
}

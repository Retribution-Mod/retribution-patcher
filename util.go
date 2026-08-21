package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

func fileNameWithoutExtension(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}

func exit() {
	logger.Debug("Cleaning up...")

	if _, e := os.Stat(directory); e == nil {
		if e := os.RemoveAll(directory); e != nil {
			logger.Errorf("Failed to clean up extracted directory: %v", e)
		}
	}

	if _, e := os.Stat(assets); e == nil {
		defer func() {
			if e := os.RemoveAll(assets); e != nil {
				logger.Errorf("Failed to clean up temporary assets directory: %v", e)
			}
		}()
	}

	logger.Info("Cleaned up.")

	os.Exit(0)
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}

	return
}

func loadInfo() {
	if info != nil {
		return
	}

	path := filepath.Join(directory, "Payload", "Discord.app", "Info.plist")
	file, err := os.Open(path)

	if err != nil {
		logger.Error("Couldn't find Info.plist. Is the provided zip an IPA file?")
		exit()
	}

	decoder := plist.NewDecoder(file)
	if err := decoder.Decode(&info); err != nil {
		logger.Error("Couldn't find Info.plist. Is the provided zip an IPA file?")
		exit()
	}
}

func saveInfo() {
	path := filepath.Join(directory, "Payload", "Discord.app", "Info.plist")
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0600)

	if err != nil {
		logger.Errorf("Failed to open Info.plist for saving: %v", err)
		exit()
	}

	logger.Debug("Saving Info.plist data...")
	encoder := plist.NewEncoder(file)
	err = encoder.Encode(info)

	if err != nil {
		logger.Errorf("Failed to save Info.plist. %v", err)
		exit()
	}

	logger.Infof("Saved Info.plist data.")
}

func isPrivateAddr(addr netip.Addr) bool {
	if addr.Is4In6() {
		addr = addr.Unmap()
	}
	return !addr.IsGlobalUnicast()
}

func isPrivateHost(host string) bool {
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		return isPrivateAddr(ip)
	}
	return false
}

func validateDownloadURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only http and https schemes are allowed")
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
	}
	host = strings.Trim(host, "[]")
	if isPrivateHost(host) {
		return errors.New("URL resolves to a private or loopback address")
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return err
	}
	for _, ip := range ips {
		if addr, ok := netip.AddrFromSlice(ip); ok {
			if isPrivateAddr(addr) {
				return errors.New("URL resolves to a private or loopback address")
			}
		}
	}

	return nil
}

func safeDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	if isPrivateHost(host) {
		return nil, errors.New("URL resolves to a private or loopback address")
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{}
	for _, ip := range ips {
		addr, ok := netip.AddrFromSlice(ip.IP)
		if !ok || isPrivateAddr(addr) {
			continue
		}
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(addr.String(), port))
		if err == nil {
			return conn, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	return nil, errors.New("URL resolves to a private or loopback address")
}

func download(rawURL string, path string) {
	if err := validateDownloadURL(rawURL); err != nil {
		logger.Errorf("Refusing to download from %s: %v", rawURL, err)
		exit()
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = safeDialContext
	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if err := validateDownloadURL(req.URL.String()); err != nil {
				return err
			}
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			return nil
		},
	}

	out, err := os.Create(path)

	if err != nil {
		logger.Errorf("Failed to pre-write file at %s.", path)
		exit()
	}

	res, err := client.Get(rawURL)

	if err != nil {
		logger.Errorf("Failed to download %s to %s %v", rawURL, path, err)
		exit()
	}

	defer res.Body.Close()
	defer out.Close()

	if res.StatusCode != http.StatusOK {
		logger.Errorf("Received bad status while downloading %s: %s", rawURL, res.Status)
		exit()
	}

	_, err = io.Copy(out, res.Body)

	if err == nil {
		logger.Infof("Successfully downloaded \"%s\" to \"%s\".", rawURL, path)
	} else {
		logger.Errorf("Failed to write \"%s\" to \"%s\": %v.", rawURL, path, err)
		exit()
	}
}

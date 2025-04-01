package fs

import (
	"archive/tar"
	"archive/zip"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/klauspost/compress/gzip"
	"github.com/nlnwa/warchaeology/v4/internal/ftpfs"
	"github.com/nlnwa/warchaeology/v4/internal/zipfs"
	whatwgUrl "github.com/nlnwa/whatwg-url/url"
	"github.com/spf13/afero"
	"github.com/spf13/afero/tarfs"
)

type fsOptions struct {
	ftpPoolSize int32
}

func WithFtpPoolSize(poolSize int32) func(*fsOptions) {
	return func(o *fsOptions) {
		o.ftpPoolSize = poolSize
	}
}

var ErrUnsupportedFilesystem = errors.New("unsupported filesystem")

// regex that matches a URL scheme
var schemeRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]+://`)

func ResolveFilesystem(fs afero.Fs, path string, options ...func(*fsOptions)) (afero.Fs, error) {
	opts := &fsOptions{}
	for _, option := range options {
		option(opts)
	}

	if path == "" {
		return fs, nil
	}

	var urlPath string

	if !schemeRegexp.MatchString(path) {
		switch filepath.Ext(path) {
		case ".tar":
			urlPath = "tar://" + path
		case ".tgz":
			urlPath = "tgz://" + path
		case ".wacz", ".zip":
			urlPath = "zip://" + path
		case ".gz":
			if filepath.Ext(strings.TrimSuffix(path, ".gz")) == ".tar" {
				urlPath = "tgz://" + path
				break
			}
			fallthrough
		default:
			return fs, nil
		}
	} else {
		urlPath = path
	}

	u, err := whatwgUrl.Parse(urlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path as URL: %w", err)
	}

	// ftp://user:password@host:port/path
	if u.Scheme() == "ftp" {
		hostPort := fmt.Sprintf("%s:%d", u.Host(), u.DecodedPort())
		return ftpfs.New(hostPort, u.Username(), u.Password(), opts.ftpPoolSize), nil
	}

	// tar://path/to/archive.tar
	if u.Scheme() == "tar" {
		name, err := url.PathUnescape(u.Hostname() + u.Pathname())
		if err != nil {
			return nil, fmt.Errorf("failed to unescape tar path: %w", err)
		}
		filepath, err := fs.Open(name)
		if err != nil {
			return nil, fmt.Errorf("failed to open tar file: %w", err)
		}
		tarReader := tar.NewReader(filepath)
		return tarfs.New(tarReader), nil
	}

	// tgz://path/to/archive.tar.gz
	if u.Scheme() == "tgz" {
		name, err := url.PathUnescape(u.Hostname() + u.Pathname())
		if err != nil {
			return nil, fmt.Errorf("failed to unescape tgz path: %w", err)
		}
		filepath, err := fs.Open(name)
		if err != nil {
			return nil, fmt.Errorf("failed to open tgz file: %w", err)
		}
		gzipReader, err := gzip.NewReader(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		tarReader := tar.NewReader(gzipReader)
		return tarfs.New(tarReader), nil
	}

	// zip://path/to/archive.zip
	if u.Scheme() == "zip" {
		name, err := url.PathUnescape(u.Hostname() + u.Pathname())
		if err != nil {
			return nil, fmt.Errorf("failed to unescape wacz path: %w", err)
		}
		filepath, err := fs.Open(name)
		if err != nil {
			return nil, fmt.Errorf("failed to open wacz file: %w", err)
		}
		fileInfo, err := filepath.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to stat wacz file: %w", err)
		}
		zipReader, err := zip.NewReader(filepath, fileInfo.Size())
		if err != nil {
			return nil, fmt.Errorf("failed to create zip reader: %w", err)
		}
		return zipfs.New(zipReader), nil
	}

	return nil, ErrUnsupportedFilesystem
}

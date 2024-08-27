package zipfs

// Because of https://github.com/spf13/afero/issues/317
// directories inside zip files are not traversed.
// This package is https://github.com/xtrafrancyz/afero/commit/0ec4cd15a07d33ba754360728ae424f5f4a15db1
// which fixes this.

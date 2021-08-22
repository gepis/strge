package internal

import (
	"archive/tar"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
)

type ZstdTOC struct {
	Version int                `json:"version"`
	Entries []ZstdFileMetadata `json:"entries"`
}

type ZstdFileMetadata struct {
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Linkname   string            `json:"linkName,omitempty"`
	Mode       int64             `json:"mode,omitempty"`
	Size       int64             `json:"size"`
	UID        int               `json:"uid"`
	GID        int               `json:"gid"`
	ModTime    time.Time         `json:"modtime"`
	AccessTime time.Time         `json:"accesstime"`
	ChangeTime time.Time         `json:"changetime"`
	Devmajor   int64             `json:"devMajor"`
	Devminor   int64             `json:"devMinor"`
	Xattrs     map[string]string `json:"xattrs,omitempty"`
	Digest     string            `json:"digest,omitempty"`
	Offset     int64             `json:"offset,omitempty"`
	EndOffset  int64             `json:"endOffset,omitempty"`

	// Currently chunking is not supported.
	ChunkSize   int64  `json:"chunkSize,omitempty"`
	ChunkOffset int64  `json:"chunkOffset,omitempty"`
	ChunkDigest string `json:"chunkDigest,omitempty"`
}

const (
	TypeReg     = "reg"
	TypeChunk   = "chunk"
	TypeLink    = "hardlink"
	TypeChar    = "char"
	TypeBlock   = "block"
	TypeDir     = "dir"
	TypeFifo    = "fifo"
	TypeSymlink = "symlink"
)

var TarTypes = map[byte]string{
	tar.TypeReg:     TypeReg,
	tar.TypeRegA:    TypeReg,
	tar.TypeLink:    TypeLink,
	tar.TypeChar:    TypeChar,
	tar.TypeBlock:   TypeBlock,
	tar.TypeDir:     TypeDir,
	tar.TypeFifo:    TypeFifo,
	tar.TypeSymlink: TypeSymlink,
}

func GetType(t byte) (string, error) {
	r, found := TarTypes[t]
	if !found {
		return "", fmt.Errorf("unknown tarball type: %v", t)
	}
	return r, nil
}
// Package protocol defines the wire format shared by sender and receiver.
//
// Version 2 runs over HTTPS (self-signed certificate pinned at pairing):
//
//	GET  /api/info                         public: name, id, protocol, fingerprint
//	POST /api/pair      {code,id,name}     → {token}
//	GET  /api/browse?path=                 folders on the receiver (paired only)
//	GET  /api/v2/offset?<file params>      → {offset, done} for resuming
//	POST /api/v2/upload?<file params>      raw file bytes from `offset`, SHA-256
//	                                       of the whole file in trailer X-Lt-Sha256
//
// File params: transfer (job id), path (slash-separated relative path),
// size (total bytes), offset, dest (optional existing folder on the receiver),
// dir=1 for an (empty) directory entry.
package protocol

import "lantransfer/internal/i18n"

const (
	Version = 2

	PathInfo   = "/api/info"
	PathPair   = "/api/pair"
	PathBrowse = "/api/browse"
	PathOffset = "/api/v2/offset"
	PathUpload = "/api/v2/upload"

	TrailerSHA256 = "X-Lt-Sha256"
)

// Info is served publicly so that discovery and the UI can show the peer.
type Info struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Port        int    `json:"port"`
	Protocol    int    `json:"protocol"`
	Fingerprint string `json:"fingerprint"`
}

type PairRequest struct {
	Code     string `json:"code"`
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

type PairResponse struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type OffsetResponse struct {
	Offset int64 `json:"offset"`
	Done   bool  `json:"done"`
}

type UploadResponse struct {
	SavedAs string `json:"saved_as"`
}

// ErrorResponse is returned with every non-2xx status. Error is English text
// (for older peers and logs); Key and Params (added in 2.1, optional) let the
// sender show the message in its own UI language.
type ErrorResponse struct {
	Error  string      `json:"error"`
	Code   string      `json:"code,omitempty"`
	Key    string      `json:"key,omitempty"`
	Params i18n.Params `json:"params,omitempty"`
}

// Machine-readable error codes.
const (
	CodeUnauthorized = "unauthorized"
	CodeNoSpace      = "no_space"
	CodeChecksum     = "checksum"
	CodeBadName      = "bad_name"
	CodeOffset       = "offset_mismatch"
	CodeTooLarge     = "too_large"
)

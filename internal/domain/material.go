package domain

// Material represents opaque content reference linked to a node
// The kernel does not interpret the content
type Material struct {
	ID        string                 `json:"material_id"`
	NodeID    string                 `json:"node_id"`
	ContentRef string                 `json:"content_ref"`
	MediaType string                 `json:"media_type"`
	ByteSize  int64                  `json:"byte_size"`
	Hash      *string                `json:"hash,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// Validate checks if the material is valid
func (m Material) Validate() error {
	if m.ID == "" {
		return ErrInvalidMaterialID
	}
	if m.NodeID == "" {
		return ErrInvalidNodeID
	}
	if m.ContentRef == "" {
		return ErrInvalidContentRef
	}
	if m.MediaType == "" {
		return ErrInvalidMediaType
	}
	if m.ByteSize < 0 {
		return ErrInvalidByteSize
	}
	return nil
}

package domain

import "errors"

// Domain errors
var (
	ErrInvalidNodeID            = errors.New("invalid node ID")
	ErrInvalidNodeTitle         = errors.New("invalid node title")
	ErrInvalidLinkID            = errors.New("invalid link ID")
	ErrInvalidFromNodeID        = errors.New("invalid from node ID")
	ErrInvalidToNodeID          = errors.New("invalid to node ID")
	ErrInvalidLinkType          = errors.New("invalid link type")
	ErrSelfLink                 = errors.New("cannot create self-link")
	ErrInvalidMaterialID         = errors.New("invalid material ID")
	ErrInvalidContentRef         = errors.New("invalid content ref")
	ErrInvalidMediaType          = errors.New("invalid media type")
	ErrInvalidByteSize           = errors.New("invalid byte size")
	ErrInvalidRoleAssignmentID   = errors.New("invalid role assignment ID")
	ErrInvalidRole               = errors.New("invalid role")
	ErrInvalidNamespaceID        = errors.New("invalid namespace ID")
	ErrInvalidNamespaceName      = errors.New("invalid namespace name")
	ErrInvalidIntentKind          = errors.New("invalid intent kind")
	ErrInvalidChangeKind          = errors.New("invalid change kind")
	ErrInvalidOperationID         = errors.New("invalid operation ID")
	ErrInvalidSeq                 = errors.New("invalid seq")
	ErrInvalidActorID             = errors.New("invalid actor ID")
	ErrInvalidPlanID               = errors.New("invalid plan ID")
	ErrInvalidPlanHash             = errors.New("invalid plan hash")
	ErrEmptyIntents               = errors.New("empty intents")
)

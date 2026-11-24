package invopop

import (
	"context"
	"errors"
	"net/url"
	"path"
)

const (
	commentsPath = "comments"
	resolvePath  = "resolve"
	reopenPath   = "reopen"
)

// Kinds of actor that may write or act on a comment.
const (
	SiloActorUser   = "user"
	SiloActorAPIKey = "api_key"
	SiloActorApp    = "app"
	SiloActorAgent  = "agent"
)

// SiloCommentsService provides access to the comments left on silo entries.
// An entry holds at most 500 comments, counting replies and deleted ones.
type SiloCommentsService service

// SiloComment is a note left on a silo entry. It applies to the whole entry,
// and may also point at fields of the entry's data and at one of its files.
type SiloComment struct {
	ID        string `json:"id" title:"ID" description:"Compound ID of the comment, made of the entry ID and a UUID."`
	EntryID   string `json:"entry_id" title:"Entry ID" description:"UUID of the entry the comment belongs to."`
	CreatedAt string `json:"created_at" title:"Created At" description:"When the comment was created."`
	UpdatedAt string `json:"updated_at" title:"Updated At" description:"When the comment was last changed, including resolutions."`

	ReplyTo string           `json:"reply_to,omitempty" title:"Reply To" description:"ID of the comment this one replies to. Replies are one level deep."`
	Version string           `json:"version,omitempty" title:"Version" description:"Version of the entry's data the comment was written against."`
	Paths   []string         `json:"paths,omitempty" title:"Paths" description:"Paths to fields of the entry's data, in the JSON Path form used for faults, rooted at the envelope." example:"$.doc.totals.payable"`
	File    *SiloCommentFile `json:"file,omitempty" title:"File" description:"File of the entry the comment refers to."`

	Body      string     `json:"body" title:"Body" description:"Text of the comment."`
	Author    *SiloActor `json:"author" title:"Author" description:"Who wrote the comment."`
	Important bool       `json:"important,omitempty" title:"Important" description:"When true, the comment should be resolved before the entry is changed or more jobs are run. A hint only, never enforced."`

	Resolved *SiloCommentAction `json:"resolved,omitempty" title:"Resolved" description:"When and by whom the comment was resolved. Absent while open."`
	Deleted  *SiloCommentAction `json:"deleted,omitempty" title:"Deleted" description:"When and by whom the comment was deleted."`

	Replies []*SiloComment `json:"replies,omitempty" title:"Replies" description:"Replies to the comment, oldest first. Only included in lists."`
}

// SiloActor identifies who performed an action on a comment.
type SiloActor struct {
	Kind string `json:"kind" title:"Kind" description:"Kind of actor: user, api_key, app or agent."`
	ID   string `json:"id" title:"ID" description:"ID of the user, API key, application or agent."`
}

// SiloCommentFile describes the file a comment refers to. A replaced file gets
// a new ID, so when this is no longer the ID of the entry's current file with
// the same name, the comment refers to an earlier version.
type SiloCommentFile struct {
	ID     string `json:"id" title:"ID" description:"UUID of the file."`
	Name   string `json:"name,omitempty" title:"Name" description:"Name of the file."`
	Hash   string `json:"hash,omitempty" title:"Hash" description:"SHA256 hash of the version of the file the comment refers to."`
	Page   int32  `json:"page,omitempty" title:"Page" description:"Page number starting at 1, for paged formats like PDF."`
	Anchor string `json:"anchor,omitempty" title:"Anchor" description:"Location inside the file, such as an XPath expression or line number."`
}

// SiloCommentFileRef points a comment at one of the entry's files.
type SiloCommentFileRef struct {
	ID     string `json:"id" title:"ID" description:"UUID of the file. Previous versions of a file may be referenced by their ID."`
	Page   int32  `json:"page,omitempty" title:"Page" description:"Page number starting at 1, for paged formats like PDF."`
	Anchor string `json:"anchor,omitempty" title:"Anchor" description:"Location inside the file, such as an XPath expression or line number."`
}

// SiloCommentAction records when and by whom a comment was resolved or deleted.
type SiloCommentAction struct {
	At string     `json:"at" title:"At" description:"When the action happened."`
	By *SiloActor `json:"by" title:"By" description:"Who performed the action."`
}

// SiloCommentCollection lists an entry's comments as threads, in the order
// they were created, each with its replies.
type SiloCommentCollection struct {
	List []*SiloComment `json:"list"`
}

// FindSiloComments filters the comments of an entry. Replies are filtered
// along with the comment they reply to, so threads are always returned whole.
type FindSiloComments struct {
	EntryID         string `json:"-"`
	IncludeResolved bool   `json:"-"` // include resolved threads
	IncludeDeleted  bool   `json:"-"` // include deleted comments and replies to them
	FileID          string `json:"-"` // only threads with a comment on this file
}

// CreateSiloComment adds a comment to an entry.
type CreateSiloComment struct {
	EntryID   string              `json:"-"`
	ReplyTo   string              `json:"reply_to,omitempty" title:"Reply To" description:"ID of the comment to reply to. Replies cannot themselves be replied to, marked important or resolved."`
	Body      string              `json:"body" title:"Body" description:"Text of the comment."`
	Paths     []string            `json:"paths,omitempty" title:"Paths" description:"Paths to fields of the entry's data, such as $.doc.totals.payable. Each must name a field of the document's GOBL type."`
	File      *SiloCommentFileRef `json:"file,omitempty" title:"File" description:"File of the entry the comment refers to."`
	Important bool                `json:"important,omitempty" title:"Important" description:"Flag the comment as one to resolve before the entry is changed or more jobs are run."`
	AgentID   string              `json:"agent_id,omitempty" title:"Agent ID" description:"ID of the agent acting through the application. Only accepted with application tokens."`
}

// UpdateSiloComment replaces the body, paths, file and important flag of a
// comment. Fields left empty are removed. Only the author may update a
// comment.
type UpdateSiloComment struct {
	EntryID   string              `json:"-"`
	ID        string              `json:"-"`
	Body      string              `json:"body" title:"Body" description:"New text of the comment."`
	Paths     []string            `json:"paths,omitempty" title:"Paths" description:"Paths to fields of the entry's data. Leave out to remove them."`
	File      *SiloCommentFileRef `json:"file,omitempty" title:"File" description:"File of the entry the comment refers to. Leave out to remove it."`
	Important bool                `json:"important,omitempty" title:"Important" description:"Flag the comment as one to resolve before the entry is changed or more jobs are run."`
	AgentID   string              `json:"agent_id,omitempty" title:"Agent ID" description:"ID of the agent acting through the application. Only accepted with application tokens."`
}

// ActOnSiloComment identifies a comment to resolve, reopen or delete.
type ActOnSiloComment struct {
	EntryID string `json:"-"`
	ID      string `json:"-"`
	AgentID string `json:"agent_id,omitempty" title:"Agent ID" description:"ID of the agent acting through the application. Only accepted with application tokens."`
}

// List provides the comments of an entry as threads. There is no paging, as
// an entry holds a limited number of comments.
func (s *SiloCommentsService) List(ctx context.Context, req *FindSiloComments) (*SiloCommentCollection, error) {
	if req == nil || req.EntryID == "" {
		return nil, errors.New(errMissingEntryID)
	}
	p := commentsBasePath(req.EntryID)
	query := make(url.Values)
	if req.IncludeResolved {
		query.Add("include_resolved", "true")
	}
	if req.IncludeDeleted {
		query.Add("include_deleted", "true")
	}
	if req.FileID != "" {
		query.Add("file_id", req.FileID)
	}
	if len(query) > 0 {
		p = p + "?" + query.Encode()
	}
	col := new(SiloCommentCollection)
	return col, s.client.get(ctx, p, col)
}

// Fetch loads a single comment, including deleted ones.
func (s *SiloCommentsService) Fetch(ctx context.Context, entryID, id string) (*SiloComment, error) {
	if err := checkCommentIDs(entryID, id); err != nil {
		return nil, err
	}
	m := new(SiloComment)
	return m, s.client.get(ctx, path.Join(commentsBasePath(entryID), id), m)
}

// Create adds a comment to an entry. Once the entry holds 500 comments, the
// API responds with a 422 error.
func (s *SiloCommentsService) Create(ctx context.Context, req *CreateSiloComment) (*SiloComment, error) {
	if req == nil || req.EntryID == "" {
		return nil, errors.New(errMissingEntryID)
	}
	if req.Body == "" {
		return nil, errors.New(errMissingBody)
	}
	m := new(SiloComment)
	return m, s.client.post(ctx, commentsBasePath(req.EntryID), req, m)
}

// Update replaces the content of a comment.
func (s *SiloCommentsService) Update(ctx context.Context, req *UpdateSiloComment) (*SiloComment, error) {
	if req == nil {
		return nil, errors.New(errMissingEntryID)
	}
	if err := checkCommentIDs(req.EntryID, req.ID); err != nil {
		return nil, err
	}
	if req.Body == "" {
		return nil, errors.New(errMissingBody)
	}
	m := new(SiloComment)
	return m, s.client.put(ctx, path.Join(commentsBasePath(req.EntryID), req.ID), req, m)
}

// Resolve marks a comment as resolved. Replies cannot be resolved.
func (s *SiloCommentsService) Resolve(ctx context.Context, req *ActOnSiloComment) (*SiloComment, error) {
	return s.act(ctx, req, resolvePath)
}

// Reopen clears a comment's resolution.
func (s *SiloCommentsService) Reopen(ctx context.Context, req *ActOnSiloComment) (*SiloComment, error) {
	return s.act(ctx, req, reopenPath)
}

// Delete marks a comment as deleted. Only the author may delete a comment.
// Deleted comments are kept, and only listed when asked for.
func (s *SiloCommentsService) Delete(ctx context.Context, req *ActOnSiloComment) (*SiloComment, error) {
	if req == nil {
		return nil, errors.New(errMissingEntryID)
	}
	if err := checkCommentIDs(req.EntryID, req.ID); err != nil {
		return nil, err
	}
	p := path.Join(commentsBasePath(req.EntryID), req.ID)
	if req.AgentID != "" {
		// DELETE has no body, so the agent travels in the query.
		p = p + "?" + url.Values{"agent_id": {req.AgentID}}.Encode()
	}
	m := new(SiloComment)
	return m, s.client.delete(ctx, p, m)
}

func (s *SiloCommentsService) act(ctx context.Context, req *ActOnSiloComment, action string) (*SiloComment, error) {
	if req == nil {
		return nil, errors.New(errMissingEntryID)
	}
	if err := checkCommentIDs(req.EntryID, req.ID); err != nil {
		return nil, err
	}
	m := new(SiloComment)
	return m, s.client.post(ctx, path.Join(commentsBasePath(req.EntryID), req.ID, action), req, m)
}

func commentsBasePath(entryID string) string {
	return path.Join(siloBasePath, entriesPath, entryID, commentsPath)
}

func checkCommentIDs(entryID, id string) error {
	if entryID == "" {
		return errors.New(errMissingEntryID)
	}
	if id == "" {
		return errors.New("missing comment ID")
	}
	return nil
}

const (
	errMissingEntryID = "missing entry ID"
	errMissingBody    = "missing body"
)

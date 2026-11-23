package invopop

import (
	"context"
	"encoding/json"
	"errors"
	"path"
)

const (
	catalogueSyncPath = "/app/catalogue/sync"
)

// What an operation does, which decides who may be given it.
const (
	// OperationEffectRead only reads.
	OperationEffectRead = "read"
	// OperationEffectWrite changes something inside Invopop.
	OperationEffectWrite = "write"
	// OperationEffectSend makes something leave Invopop, such as an email.
	OperationEffectSend = "send"
)

// Outcomes of syncing a part of an application's registration.
const (
	SyncedPartCreated   = "created"
	SyncedPartUpdated   = "updated"
	SyncedPartUnchanged = "unchanged"
)

// ErrMissingOperations is returned by CatalogueService.Sync before making a
// request without a list of operations. An empty list is not missing: it
// removes every operation.
var ErrMissingOperations = errors.New("missing operations")

// CatalogueService lets applications register what the rest of Invopop may
// use of them.
type CatalogueService service

// SyncCatalogue holds everything an application registers.
type SyncCatalogue struct {
	Operations []*OperationDefinition `json:"operations" title:"Operations" description:"Every operation of the application's API that Invopop may use as a workspace. Operations left out are removed, and can no longer be called."`
	// Skills are sent when the slice is not nil: an empty one removes every
	// skill, and nil leaves them as they are.
	Skills []*SkillDefinition `json:"skills,omitempty" title:"Skills" description:"Every skill the application gives agents. Skills left out are removed. Leave the field out altogether to leave the application's skills as they are."`
	// Panels are sent when the slice is not nil, as skills are.
	Panels []*PanelDefinition `json:"panels,omitempty" title:"Panels" description:"Every panel the application shows in Desk. Panels left out are removed. Leave the field out altogether to leave the application's panels as they are."`
}

// PanelDefinition is a static description of something the application
// shows on one of Desk's pages. The sources it reads and the calls it makes
// are operations the application registers.
type PanelDefinition struct {
	Key     string            `json:"key" title:"Key" description:"Identifies the panel within the application: lowercase letters, digits and dashes." example:"invoice-tab"`
	Title   string            `json:"title" title:"Title" description:"Name shown to users with the panel." example:"Engine"`
	Kind    string            `json:"kind,omitempty" title:"Kind" description:"How the panel is described. Only \"jsonui\", the default, for now."`
	Slots   []string          `json:"slots" title:"Slots" description:"Where the panel fits: \"entry.tab\", a tab of an entry's page, or \"entry.side\", a card in its sidebar."`
	Schemas []string          `json:"schemas" title:"Schemas" description:"GOBL schemas of the entries the panel applies to."`
	Sources map[string]string `json:"sources,omitempty" title:"Sources" description:"Read operations the panel shows, by the names it binds to them with, each as \"<app_id>:<key>\"."`
	UI      any               `json:"ui" title:"UI" description:"PopUI JSON UI panel, as anything that is written as its JSON."`
}

// SyncedPanel is a panel as Desk registered it.
type SyncedPanel struct {
	Key    string `json:"key" title:"Key" description:"Key of the panel."`
	Digest string `json:"digest" title:"Digest" description:"Identifies this version of the panel."`
}

// MarshalJSON keeps an empty list of skills or of panels, which removes
// them all, apart from no list, which leaves them alone.
func (s *SyncCatalogue) MarshalJSON() ([]byte, error) {
	type plain struct {
		Operations []*OperationDefinition `json:"operations"`
		Skills     *[]*SkillDefinition    `json:"skills,omitempty"`
		Panels     *[]*PanelDefinition    `json:"panels,omitempty"`
	}
	out := &plain{Operations: s.Operations}
	if s.Skills != nil {
		out.Skills = &s.Skills
	}
	if s.Panels != nil {
		out.Panels = &s.Panels
	}
	return json.Marshal(out)
}

// What Barn shows the model of a session's silo entry, in a skill's Gives.
const (
	SkillGivesDocument = "document"
	SkillGivesFiles    = "files"
)

// Statuses of a skill's results.
const (
	SkillResultOK = "ok"
	SkillResultKO = "ko"
)

// SkillDefinition is what an agent in Barn is told to do, what it is given
// and what it may use. Agents name it "<app_id>:<key>".
type SkillDefinition struct {
	Key          string                  `json:"key" title:"Key" description:"Identifies the skill within the application: lowercase letters, digits and dashes." example:"purchase-accounts"`
	Title        string                  `json:"title" title:"Title" description:"Name shown to people choosing skills for an agent." example:"Purchase accounts"`
	Description  string                  `json:"description,omitempty" title:"Description" description:"What an agent with the skill does, for the people choosing it."`
	Instructions string                  `json:"instructions" title:"Instructions" description:"What the model is told, sent as it is."`
	Gives        []string                `json:"gives,omitempty" title:"Gives" description:"What Barn shows the model of the session's silo entry: \"document\" and \"files\"."`
	Sources      map[string]*SkillSource `json:"sources,omitempty" title:"Sources" description:"Read operations Barn makes before the model is first called, by the name the model is given each under."`
	Tools        []string                `json:"tools,omitempty" title:"Tools" description:"Tools of Barn's own the agent may call."`
	Operations   []*SkillOperation       `json:"operations,omitempty" title:"Operations" description:"Operations of the catalogue the agent may call as tools."`
	Results      []*SkillResult          `json:"results" title:"Results" description:"The ways a session may end, which become the results of the workflow step."`
}

// SkillSource is a read operation Barn makes for a skill.
type SkillSource struct {
	Operation string            `json:"operation" title:"Operation" description:"Reference of a read operation, \"<app_id>:<key>\"." example:"engine:purchase-coding"`
	Params    map[string]string `json:"params,omitempty" title:"Parameters" description:"Fill the operation's path, as text that may hold \"{entry_id}\", \"{agent_id}\" and \"{session_id}\". A parameter named \"entry_id\" is the session's document unless it says otherwise."`
}

// SkillOperation is an operation a skill's agent may call as a tool.
type SkillOperation struct {
	Operation   string            `json:"operation" title:"Operation" description:"Reference of the operation, \"<app_id>:<key>\"." example:"engine:set-purchase-lines"`
	Name        string            `json:"name,omitempty" title:"Name" description:"What the model calls the tool. Without one it is the operation's key, with underscores." example:"set_accounts"`
	Description string            `json:"description,omitempty" title:"Description" description:"Replaces the operation's own description, to say when this agent is to use it."`
	Params      map[string]string `json:"params,omitempty" title:"Parameters" description:"Fill parameters of the path, which the model then cannot set. Besides the variables of a source they may hold \"{sources.<name>.<field>}\", a field of what a source returns."`
	Fixed       map[string]any    `json:"fixed,omitempty" title:"Fixed" description:"Fields of the body, or of a read's query, that the skill decides. The model is not shown them and cannot set them."`
}

// SkillResult is one way a session of a skill may end.
type SkillResult struct {
	Code        string `json:"code" title:"Code" description:"Code that identifies the result." example:"assigned"`
	Status      string `json:"status" title:"Status" description:"Either \"ok\" or \"ko\"." example:"ok"`
	Description string `json:"description,omitempty" title:"Description" description:"What the result means, for the model and for people."`
}

// SyncedSkills is the outcome for the skills of a registration.
type SyncedSkills struct {
	Status string   `json:"status" title:"Status" description:"Either \"created\", \"updated\" or \"unchanged\"."`
	Digest string   `json:"digest" title:"Digest" description:"Identifies this version of the skills."`
	Keys   []string `json:"keys,omitempty" title:"Keys" description:"The skills as agents name them."`
}

// OperationDefinition is one request to the application's own API that
// Invopop may make as a workspace, for a panel in Desk, an agent or a script.
// Outside the application it is named "<app_id>:<key>".
type OperationDefinition struct {
	Key          string            `json:"key" title:"Key" description:"Identifies the operation within the application: lowercase letters, digits and dashes." example:"set-purchase-lines"`
	Title        string            `json:"title" title:"Title" description:"Name shown to people choosing the operation." example:"Set a purchase's accounts"`
	Description  string            `json:"description,omitempty" title:"Description" description:"What the operation does and when to use it, written for an agent or the author of a script."`
	Effect       string            `json:"effect" title:"Effect" description:"What the operation does: \"read\", \"write\" for a change inside Invopop, or \"send\" for anything that leaves Invopop." example:"write"`
	Method       string            `json:"method,omitempty" title:"Method" description:"GET for a read, which is the default. One of POST, PUT, PATCH or DELETE otherwise." example:"PATCH"`
	Path         string            `json:"path" title:"Path" description:"Where the application's API takes the request, relative to its API URL. Parameters are in braces, each a whole segment." example:"/purchases/{id}/lines"`
	Params       []*OperationParam `json:"params,omitempty" title:"Parameters" description:"Descriptions of the parameters in the path. The names are read from the path."`
	QuerySchema  json.RawMessage   `json:"query_schema,omitempty" title:"Query Schema" description:"JSON Schema of the query a read accepts. It may only \"$ref\" itself."`
	BodySchema   json.RawMessage   `json:"body_schema,omitempty" title:"Body Schema" description:"JSON Schema of the body a write or send takes. It may only \"$ref\" itself."`
	ResultSchema json.RawMessage   `json:"result_schema,omitempty" title:"Result Schema" description:"JSON Schema of what the operation returns. It may only \"$ref\" itself."`
	Example      json.RawMessage   `json:"example,omitempty" title:"Example" description:"Something the operation could return, which must fit the result schema."`
	Scope        string            `json:"scope" title:"Scope" description:"Scope a person needs to use the operation, or to grant it to an agent or a script." example:"admin:silo:entries"`
	Schemas      []string          `json:"schemas,omitempty" title:"Schemas" description:"GOBL schemas the operation applies to, when it is about one kind of document." example:"[\"bill/invoice\"]"`
}

// OperationParam describes a parameter in an operation's path.
type OperationParam struct {
	Name        string `json:"name" title:"Name" description:"Name of the parameter, as it is in the path." example:"id"`
	Description string `json:"description,omitempty" title:"Description" description:"What the parameter is." example:"The purchase's ID."`
}

// SyncCatalogueResult reports what was done with each part of a registration.
type SyncCatalogueResult struct {
	AppID      string         `json:"app_id" title:"Application ID" description:"ID of the application the registration belongs to."`
	Operations *SyncedPart    `json:"operations" title:"Operations" description:"What was done with the operations."`
	Skills     *SyncedSkills  `json:"skills,omitempty" title:"Skills" description:"What was done with the skills, when the request had any."`
	Panels     []*SyncedPanel `json:"panels,omitempty" title:"Panels" description:"Each panel in the request, as Desk registered it, when the request had any."`
}

// SyncedPart is the outcome for one part of a registration.
type SyncedPart struct {
	Status string `json:"status" title:"Status" description:"Either \"created\", \"updated\" or \"unchanged\"."`
	Digest string `json:"digest" title:"Digest" description:"Identifies this version of the part: it changes when anything in it does."`
}

// Catalogue returns the service for what an application registers for the
// rest of Invopop to use.
func (s *AccessService) Catalogue() *CatalogueService {
	return (*CatalogueService)(s)
}

// Sync sends everything the application registers, such as on every
// startup. The operations replace those registered before: all of them are
// validated before anything is written, and those left out are removed and
// can no longer be called. A list equal to what is registered writes
// nothing, so syncing often is cheap.
//
// The client must be configured with the application's OAuth client ID and
// secret, and the application must have the "enrolled" scope. They are always
// sent, even when the client also carries an auth token, as the API refuses
// tokens for this endpoint.
func (s *CatalogueService) Sync(ctx context.Context, req *SyncCatalogue) (*SyncCatalogueResult, error) {
	c := s.client
	if c.clientID == "" || c.clientSecret == "" {
		return nil, ErrMissingCredentials
	}
	if req == nil || req.Operations == nil {
		return nil, ErrMissingOperations
	}

	p := path.Join(accessBasePath, catalogueSyncPath)
	out := new(SyncCatalogueResult)
	re := new(ResponseError)
	res, err := c.conn.R().
		SetContext(ctx).
		// A token would replace the Basic header, so drop any the client has.
		SetAuthToken("").
		SetBasicAuth(c.clientID, c.clientSecret).
		SetBody(req).
		SetError(re).
		SetResult(out).
		Post(p)
	if err != nil {
		return nil, err
	}
	if err := re.handle(res); err != nil {
		return nil, err
	}
	return out, nil
}

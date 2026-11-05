package invopop

import (
	"context"
	"encoding/json"
	"errors"
	"path"
)

const (
	actionsSyncPath = "/app/actions/sync"
)

// Action result statuses.
const (
	ActionResultOK = "OK"
	ActionResultKO = "KO"
)

// Seat renewal options for an action. An empty value is the same as
// SeatRenewalNone.
const (
	SeatRenewalNone      = "none"
	SeatRenewalMonthly   = "monthly"
	SeatRenewalQuarterly = "quarterly"
	SeatRenewalYearly    = "yearly"
	SeatRenewalUnroll    = "unroll"
)

// Outcomes of syncing an action definition.
const (
	SyncedActionCreated   = "created"
	SyncedActionUpdated   = "updated"
	SyncedActionUnchanged = "unchanged"
)

// Errors returned by ActionsService.Sync before making a request.
var (
	ErrMissingCredentials = errors.New("missing OAuth client credentials: client_id and client_secret")
	ErrMissingActions     = errors.New("missing actions")
)

// ActionsService lets applications keep the actions they provide up to date.
type ActionsService service

// SyncActions holds an application's complete list of action definitions.
type SyncActions struct {
	Actions []*ActionDefinition `json:"actions" title:"Actions" description:"Every action the application provides. Actions left out are not changed or removed, but are reported as undeclared."`
}

// ActionDefinition holds every field of an action that the application owns.
// All of them are set on each sync, so a field left empty is cleared.
type ActionDefinition struct {
	ID           string          `json:"id" title:"ID" description:"Action ID, which must start with the application's ID: \"<app_id>[.<channel>].<action>[.<qualifier>]\"." example:"verifactu.send"`
	Name         string          `json:"name" title:"Name" description:"Name shown to users when choosing the action." example:"Send to VERI*FACTU"`
	Description  string          `json:"description,omitempty" title:"Description" description:"Short description of what the action does."`
	IconURL      string          `json:"icon_url,omitempty" title:"Icon URL" description:"URL of the action's icon."`
	ConfigURL    string          `json:"config_url,omitempty" title:"Config URL" description:"Page the console loads in an iframe to configure workflow steps that use the action."`
	ConfigUI     json.RawMessage `json:"config_ui,omitempty" title:"Config UI" description:"Experimental. PopUI JSON UI panel the console renders to configure steps, instead of loading the config URL. Its \"config\" data is the step's configuration."`
	ConfigSchema json.RawMessage `json:"config_schema,omitempty" title:"Config Schema" description:"Experimental. JSON Schema a step's configuration must match. It may only \"$ref\" itself."`
	InfoURL      string          `json:"info_url,omitempty" title:"Info URL" description:"URL of a Markdown document explaining how to use the action."`
	Schemas      []string        `json:"schemas,omitempty" title:"Schemas" description:"GOBL schemas of the documents the action accepts." example:"[\"bill/invoice\"]"`
	Countries    []string        `json:"countries,omitempty" title:"Countries" description:"Countries the action applies to." example:"[\"ES\"]"`
	Results      []*ActionResult `json:"results,omitempty" title:"Results" description:"Results a task of the action may complete with."`
	Pops         int32           `json:"pops,omitempty" title:"Pops" description:"What each use of the action costs, in pops, from 0 to 1000." example:"10"`
	SeatRenewal  string          `json:"seat_renewal,omitempty" title:"Seat Renewal" description:"Whether the action allocates a billing seat, keyed by the application and channel, and how often its pops charge renews: none, monthly, quarterly or yearly. \"unroll\" releases the seat instead." example:"monthly"`
}

// ActionResult is one of the results a task of an action can complete with.
type ActionResult struct {
	Status      string `json:"status" title:"Status" description:"Either \"OK\" or \"KO\"." example:"OK"`
	Code        string `json:"code,omitempty" title:"Code" description:"Code that identifies the result." example:"sent"`
	Description string `json:"description,omitempty" title:"Description" description:"Description of the result."`
	Default     bool   `json:"default,omitempty" title:"Default" description:"Whether this is the default result for its status."`
}

// SyncActionsResult reports what the sync did with each action definition.
type SyncActionsResult struct {
	AppID      string          `json:"app_id" title:"Application ID" description:"ID of the application the actions belong to."`
	Actions    []*SyncedAction `json:"actions" title:"Actions" description:"Outcome of each action definition in the request."`
	Undeclared []string        `json:"undeclared,omitempty" title:"Undeclared" description:"IDs of the application's existing actions that weren't in the request. They are left as they are."`
}

// SyncedAction is the outcome of syncing one action definition.
type SyncedAction struct {
	ID     string `json:"id" title:"ID" description:"ID of the action."`
	Status string `json:"status" title:"Status" description:"Either \"created\", \"updated\" or \"unchanged\"."`
	Rev    string `json:"rev" title:"Revision" description:"The action's revision after the sync."`
}

// Actions returns the service for an application's own actions.
func (s *AccessService) Actions() *ActionsService {
	return (*ActionsService)(s)
}

// Sync sends the application's complete list of action definitions, such as
// on every startup. Access validates them all before writing anything, then
// creates the new ones and updates those that differ; unchanged actions
// aren't written, so syncing often is cheap. Actions left out of the list are
// not removed, but are reported as undeclared.
//
// The application owns every field in a definition, including its pricing,
// and the definitions sent always win over manual changes. A field left empty
// is cleared.
//
// The client must be configured with the application's OAuth client ID and
// secret, and the application must have the "enrolled" scope. They are always
// sent, even when the client also carries an auth token, as the API refuses
// tokens for this endpoint.
func (s *ActionsService) Sync(ctx context.Context, req *SyncActions) (*SyncActionsResult, error) {
	c := s.client
	if c.clientID == "" || c.clientSecret == "" {
		return nil, ErrMissingCredentials
	}
	if req == nil || req.Actions == nil {
		return nil, ErrMissingActions
	}

	p := path.Join(accessBasePath, actionsSyncPath)
	out := new(SyncActionsResult)
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

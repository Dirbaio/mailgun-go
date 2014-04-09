package mailgun

import (
	"github.com/mbanzon/simplehttp"
	"strconv"
	"encoding/json"
//	"fmt"
)

// A mailing list may have one of three membership modes.
// ReadOnly specifies that nobody, including subscribers,
// may send messages to the mailing list.
// Messages distributed on such lists come from list administrator accounts only.
// Members specifies that only those who subscribe to the mailing list may send messages.
// Everyone specifies that anyone and everyone may both read and submit messages
// to the mailing list, including non-subscribers.
const (
	ReadOnly = "readonly"
	Members = "members"
	Everyone = "everyone"
)

// Mailing list members have an attribute that determines if they've subscribed to the mailing list or not.
// This attribute may be used to filter the results returned by GetSubscribers().
// All, Subscribed, and Unsubscribed provides a convenient and readable syntax for specifying the scope of the search.
var (
	All *bool = nil
	Subscribed *bool = &yes
	Unsubscribed *bool = &no
)

// yes and no are variables which provide us the ability to take their addresses.
// Subscribed and Unsubscribed are pointers to these booleans.
//
// We use a pointer to boolean as a kind of trinary data type:
// if nil, the relevant data type remains unspecified.
// Otherwise, its value is either true or false.
var (
	yes bool = true
	no bool = false
)

// A List structure provides information for a mailing list.
//
// AccessLevel may be one of ReadOnly, Members, or Everyone.
type List struct {
	Address      string `json:"address",omitempty"`
	Name         string `json:"name",omitempty"`
	Description  string `json:"description",omitempty"`
	AccessLevel  string `json:"access_level",omitempty"`
	CreatedAt    string `json:"created_at",omitempty"`
	MembersCount int    `json:"members_count",omitempty"`
}

// A Subscriber structure represents a member of the mailing list.
// The Vars field can represent any JSON-encodable data.
type Subscriber struct {
	Address    string `json:"address,omitempty"`
	Name       string `json:"name,omitempty"`
	Subscribed *bool `json:"subscribed,omitempty"`
	Vars       map[string]interface{} `json:"vars,omitempty"`
}

// GetLists returns the specified set of mailing lists administered by your account.
func (mg *mailgunImpl) GetLists(limit, skip int, filter string) (int, []List, error) {
	r := simplehttp.NewHTTPRequest(generatePublicApiUrl(listsEndpoint))
	r.SetBasicAuth(basicAuthUser, mg.ApiKey())
	p := simplehttp.NewUrlEncodedPayload()
	if limit != DefaultLimit {
		p.AddValue("limit", strconv.Itoa(limit))
	}
	if skip != DefaultSkip {
		p.AddValue("skip", strconv.Itoa(skip))
	}
	if filter != "" {
		p.AddValue("address", filter)
	}
	var envelope struct {
		Items []List `json:"items"`
		TotalCount int `json:"total_count"`
	}
	response, err := r.MakeRequest("GET", p)
	if err != nil {
		return -1, nil, err
	}
	err = response.ParseFromJSON(&envelope)
	return envelope.TotalCount, envelope.Items, err
}

// CreateList creates a new mailing list under your Mailgun account.
// You need specify only the Address and Name members of the prototype;
// Description, and AccessLevel are optional.
// If unspecified, Description remains blank,
// while AccessLevel defaults to Everyone.
func (mg *mailgunImpl) CreateList(prototype List) (List, error) {
	r := simplehttp.NewHTTPRequest(generatePublicApiUrl(listsEndpoint))
	r.SetBasicAuth(basicAuthUser, mg.ApiKey())
	p := simplehttp.NewUrlEncodedPayload()
	if prototype.Address != "" {
		p.AddValue("address", prototype.Address)
	}
	if prototype.Name != "" {
		p.AddValue("name", prototype.Name)
	}
	if prototype.Description != "" {
		p.AddValue("description", prototype.Description)
	}
	if prototype.AccessLevel != "" {
		p.AddValue("access_level", prototype.AccessLevel)
	}
	response, err := r.MakePostRequest(p)
	if err != nil {
		return List{}, err
	}
	var l List
	err = response.ParseFromJSON(&l)
	return l, err
}

// DeleteList removes all current members of the list, then removes the list itself.
// Attempts to send e-mail to the list will fail subsequent to this call.
func (mg *mailgunImpl) DeleteList(addr string) error {
	r := simplehttp.NewHTTPRequest(generatePublicApiUrl(listsEndpoint) + "/" + addr)
	r.SetBasicAuth(basicAuthUser, mg.ApiKey())
	_, err := r.MakeDeleteRequest()
	return err
}

// GetListByAddress allows your application to recover the complete List structure
// representing a mailing list, so long as you have its e-mail address.
func (mg *mailgunImpl) GetListByAddress(addr string) (List, error) {
	r := simplehttp.NewHTTPRequest(generatePublicApiUrl(listsEndpoint) + "/" + addr)
	r.SetBasicAuth(basicAuthUser, mg.ApiKey())
	response, err := r.MakeGetRequest()
	var envelope struct {
		List `json:"list"`
	}
	err = response.ParseFromJSON(&envelope)
	return envelope.List, err
}

// UpdateList allows you to change various attributes of a list.
// Address, Name, Description, and AccessLevel are all optional;
// only those fields which are set in the prototype will change.
//
// Be careful!  If changing the address of a mailing list,
// e-mail sent to the old address will not succeed.
// Make sure you account for the change accordingly.
func (mg *mailgunImpl) UpdateList(addr string, prototype List) (List, error) {
	r := simplehttp.NewHTTPRequest(generatePublicApiUrl(listsEndpoint) + "/" + addr)
	r.SetBasicAuth(basicAuthUser, mg.ApiKey())
	p := simplehttp.NewUrlEncodedPayload()
	if prototype.Address != "" {
		p.AddValue("address", prototype.Address)
	}
	if prototype.Name != "" {
		p.AddValue("name", prototype.Name)
	}
	if prototype.Description != "" {
		p.AddValue("description", prototype.Description)
	}
	if prototype.AccessLevel != "" {
		p.AddValue("access_level", prototype.AccessLevel)
	}
	var l List
	response, err := r.MakePutRequest(p)
	if err != nil {
		return l, err
	}
	err = response.ParseFromJSON(&l)
	return l, err
}

// GetSubscribers returns the list of members belonging to the indicated mailing list.
// The s parameter can be set to one of three settings to help narrow the returned data set:
// All indicates that you want both subscribers and unsubscribed members alike, while
// Subscribed and Unsubscribed indicate you want only those eponymous subsets.
func (mg *mailgunImpl) GetSubscribers(limit, skip int, s *bool, addr string) (int, []Subscriber, error) {
	r := simplehttp.NewHTTPRequest(generateSubscriberApiUrl(listsEndpoint, addr))
	r.SetBasicAuth(basicAuthUser, mg.ApiKey())
	p := simplehttp.NewUrlEncodedPayload()
	if limit != DefaultLimit {
		p.AddValue("limit", strconv.Itoa(limit))
	}
	if skip != DefaultSkip {
		p.AddValue("skip", strconv.Itoa(skip))
	}
	if s != nil {
		p.AddValue("subscribed", yesNo(*s))
	}
	var envelope struct {
		TotalCount int `json:"total_count"`
		Items []Subscriber `json:"items"`
	}
	response, err := r.MakeRequest("GET", p)
	if err != nil {
		return -1, nil, err
	}
	err = response.ParseFromJSON(&envelope)
	return envelope.TotalCount, envelope.Items, err
}

// GetSubscriberByAddress returns a complete Subscriber structure for a member of a mailing list,
// given only their subscription e-mail address.
func (mg *mailgunImpl) GetSubscriberByAddress(s, l string) (Subscriber, error) {
	r := simplehttp.NewHTTPRequest(generateSubscriberApiUrl(listsEndpoint, l) + "/" + s)
	r.SetBasicAuth(basicAuthUser, mg.ApiKey())
	response, err := r.MakeGetRequest()
	if err != nil {
		return Subscriber{}, err
	}
	var envelope struct {
		Member Subscriber `json:"member"`
	}
	err = response.ParseFromJSON(&envelope)
	return envelope.Member, err
}

// CreateSubscriber registers a new member of the indicated mailing list.
// If merge is set to true, then the registration may update an existing subscriber's settings.
// Otherwise, an error will occur if you attempt to add a member with a duplicate e-mail address.
func (mg *mailgunImpl) CreateSubscriber(merge bool, addr string, prototype Subscriber) error {
	vs, err := json.Marshal(prototype.Vars)
	if err != nil {
		return err
	}

	r := simplehttp.NewHTTPRequest(generateSubscriberApiUrl(listsEndpoint, addr))
	r.SetBasicAuth(basicAuthUser, mg.ApiKey())
	p := simplehttp.NewFormDataPayload()
	p.AddValue("upsert", yesNo(merge))
	p.AddValue("address", prototype.Address)
	p.AddValue("name", prototype.Name)
	p.AddValue("vars", string(vs))
	if prototype.Subscribed != nil {
		p.AddValue("subscribed", yesNo(*prototype.Subscribed))
	}
	_, err = r.MakePostRequest(p)
	return err
}

// UpdateSubscriber lets you change certain details about the indicated mailing list member.
// Address, Name, Vars, and Subscribed fields may be changed.
func (mg *mailgunImpl) UpdateSubscriber(s, l string, prototype Subscriber) (Subscriber, error) {
	r := simplehttp.NewHTTPRequest(generateSubscriberApiUrl(listsEndpoint, l) + "/" + s)
	r.SetBasicAuth(basicAuthUser, mg.ApiKey())
	p := simplehttp.NewFormDataPayload()
	if prototype.Address != "" {
		p.AddValue("address", prototype.Address)
	}
	if prototype.Name != "" {
		p.AddValue("name", prototype.Name)
	}
	if prototype.Vars != nil {
		vs, err := json.Marshal(prototype.Vars)
		if err != nil {
			return Subscriber{}, err
		}
		p.AddValue("vars", string(vs))
	}
	if prototype.Subscribed != nil {
		p.AddValue("subscribed", yesNo(*prototype.Subscribed))
	}
	response, err := r.MakePutRequest(p)
	if err != nil {
		return Subscriber{}, err
	}
	var envelope struct {
		Member Subscriber `json:"member"`
	}
	err = response.ParseFromJSON(&envelope)
	return envelope.Member, err
}
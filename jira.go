package gojira

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Jira struct {
	BaseUrl      string
	ApiPath      string
	ActivityPath string
	Client       *http.Client
	Auth         *Auth
}

type Auth struct {
	Login    string
	Password string
}

type Pagination struct {
	Total      int
	StartAt    int
	MaxResults int
	Page       int
	PageCount  int
	Pages      []int
}

func (p *Pagination) Compute() {
	p.PageCount = int(math.Ceil(float64(p.Total) / float64(p.MaxResults)))
	p.Page = int(math.Ceil(float64(p.StartAt) / float64(p.MaxResults)))

	p.Pages = make([]int, p.PageCount)
	for i := range p.Pages {
		p.Pages[i] = i
	}
}

type Issue struct {
	Id        string
	Key       string
	Self      string
	Expand    string
	Fields    *IssueFields
	CreatedAt time.Time
}

type IssueList struct {
	Expand     string
	StartAt    int
	MaxResults int
	Total      int
	Issues     []*Issue
	Pagination *Pagination
}

type IssueFields struct {
	IssueType   *IssueType
	Summary     string
	Description string
	Reporter    *User
	Assignee    *User
	Project     *JiraProject
	Created     string
}

type IssueType struct {
	Self        string
	Id          string
	Description string
	IconUrl     string
	Name        string
	Subtask     bool
}

type JiraProject struct {
	Self       string
	Id         string
	Key        string
	Name       string
	AvatarUrls map[string]string
}

type ActivityItem struct {
	Title    string    `xml:"title"json:"title"`
	Id       string    `xml:"id"json:"id"`
	Link     []Link    `xml:"link"json:"link"`
	Updated  time.Time `xml:"updated"json:"updated"`
	Author   Person    `xml:"author"json:"author"`
	Summary  Text      `xml:"summary"json:"summary"`
	Category Category  `xml:"category"json:"category"`
}

type ActivityFeed struct {
	XMLName xml.Name        `xml:"http://www.w3.org/2005/Atom feed"json:"xml_name"`
	Title   string          `xml:"title"json:"title"`
	Id      string          `xml:"id"json:"id"`
	Link    []Link          `xml:"link"json:"link"`
	Updated time.Time       `xml:"updated,attr"json:"updated"`
	Author  Person          `xml:"author"json:"author"`
	Entries []*ActivityItem `xml:"entry"json:"entries"`
}

type Category struct {
	Term string `xml:"term,attr"json:"term"`
}

type Link struct {
	Rel  string `xml:"rel,attr,omitempty"json:"rel"`
	Href string `xml:"href,attr"json:"href"`
}

type Person struct {
	Name     string `xml:"name"json:"name"`
	URI      string `xml:"uri"json:"uri"`
	Email    string `xml:"email"json:"email"`
	InnerXML string `xml:",innerxml"json:"inner_xml"`
}

type Text struct {
	Type string `xml:"type,attr,omitempty"json:"type"`
	Body string `xml:",chardata"json:"body"`
}

func NewJira(baseUrl string, apiPath string, activityPath string, auth *Auth) *Jira {

	client := &http.Client{}

	return &Jira{
		BaseUrl:      baseUrl,
		ApiPath:      apiPath,
		ActivityPath: activityPath,
		Client:       client,
		Auth:         auth,
	}
}

const (
	dateLayout = "2006-01-02T15:04:05.000-0700"
)

func (j *Jira) getRequest(uri string) ([]byte, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	return j.execRequest(req)
}

func (j *Jira) postJson(uri string, body *bytes.Buffer) ([]byte, error) {
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return j.execRequest(req)
}

func (j *Jira) execRequest(req *http.Request) ([]byte, error) {
	req.SetBasicAuth(j.Auth.Login, j.Auth.Password)
	resp, err := j.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func (j *Jira) UserActivity(user string) (ActivityFeed, error) {
	uri := j.BaseUrl + j.ActivityPath + "?streams=" + url.QueryEscape("user IS "+user)

	return j.Activity(uri)
}

func (j *Jira) Activity(uri string) (a ActivityFeed, err error) {

	contents, err := j.getRequest(uri)
	if err != nil {
		return
	}
	var activity ActivityFeed
	err = xml.Unmarshal(contents, &activity)
	if err != nil {
		return
	}
	return activity, nil
}

// search issues assigned to given user
func (j *Jira) IssuesAssignedTo(user string, maxResults int, startAt int) (i IssueList, err error) {

	uri := j.BaseUrl + j.ApiPath + "/search?jql=assignee=\"" + url.QueryEscape(user) + "\"&startAt=" + strconv.Itoa(startAt) + "&maxResults=" + strconv.Itoa(maxResults)
	contents, err := j.getRequest(uri)
	if err != nil {
		return
	}
	var issues IssueList
	err = json.Unmarshal(contents, &issues)
	if err != nil {
		return
	}

	for _, issue := range issues.Issues {
		t, _ := time.Parse(dateLayout, issue.Fields.Created)
		issue.CreatedAt = t
	}

	pagination := Pagination{
		Total:      issues.Total,
		StartAt:    issues.StartAt,
		MaxResults: issues.MaxResults,
	}
	pagination.Compute()

	issues.Pagination = &pagination

	return issues, nil
}

// search an issue by its id
func (j *Jira) Issue(id string) (i Issue, err error) {

	uri := j.BaseUrl + j.ApiPath + "/issue/" + id
	contents, err := j.getRequest(uri)
	if err != nil {
		return
	}

	var issue Issue
	err = json.Unmarshal(contents, &issue)
	if err != nil {
		return
	}

	return issue, nil
}

func (j *Jira) AddComment(issue *Issue, comment string) error {
	var cMap = make(map[string]string)
	cMap["body"] = comment
	cJson, err := json.Marshal(cMap)
	if err != nil {
		return err
	}
	uri := j.BaseUrl + j.ApiPath + "/issue/" + issue.Key + "/comment"
	body := bytes.NewBuffer(cJson)
	_, err = j.postJson(uri, body)
	if err != nil {
		return err
	}
	return nil
}

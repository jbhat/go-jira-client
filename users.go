package gojira

import (
	"encoding/json"
)

const (
	user_uri        = "/user"
	user_search_uri = "/user/search"
	// http://example.com:8080/jira/rest/api/2/user/assignable/multiProjectSearch [GET]
	// http://example.com:8080/jira/rest/api/2/user/assignable/search [GET]
	// http://example.com:8080/jira/rest/api/2/user/avatar [POST, PUT]
	// http://example.com:8080/jira/rest/api/2/user/avatar/temporary [POST, POST]
	// http://example.com:8080/jira/rest/api/2/user/avatar/{id} [DELETE]
	// http://example.com:8080/jira/rest/api/2/user/avatars [GET]
	// http://example.com:8080/jira/rest/api/2/user/picker [GET]
	// http://example.com:8080/jira/rest/api/2/user/viewissue/search [GET]
)

type User struct {
	Self         string            `json:"self"`
	Name         string            `json:"name"`
	EmailAddress string            `json:"emailAddress"`
	DisplayName  string            `json:"displayName"`
	Active       bool              `json:"active"`
	TimeZone     string            `json:"timeZone"`
	AvatarUrls   map[string]string `json:"avatarUrls"`
	Expand       string            `json:"expand"`
	// "groups": {
	//     "size": 3,
	//     "items": [
	//         {
	//             "name": "jira-user",
	//             "self": "http://www.example.com/jira/rest/api/2/group?groupname=jira-user"
	//         },
	//         {
	//             "name": "jira-admin",
	//             "self": "http://www.example.com/jira/rest/api/2/group?groupname=jira-admin"
	//         },
	//         {
	//             "name": "important",
	//             "self": "http://www.example.com/jira/rest/api/2/group?groupname=important"
	//         }
	//     ]
	// }
}

/*
Returns a user. This resource cannot be accessed anonymously.

    GET http://example.com:8080/jira/rest/api/2/user?username=USERNAME

Parameters

    username string The username

Usage

	user, err := jira.User("username")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("%+v\n", user)
*/
func (j *Jira) User(username string) (u *User, err error) {
	uri := j.BaseUrl + j.ApiPath + user_uri + "?username=" + username
	contents, err := j.getRequest(uri)
	if err != nil {
		return
	}

	user := new(User)
	err = json.Unmarshal(contents, &user)
	if err != nil {
		return
	}

	return user, nil
}

/*
Returns a list of users that match the search string. This resource cannot be accessed anonymously.

	GET http://example.com:8080/jira/rest/api/2/user/search

Parameters

	username        string  A query string used to search username, name or e-mail address
	startAt         int     The index of the first user to return (0-based)
	maxResults      int     The maximum number of users to return (defaults to 50).
				   	        The maximum allowed value is 1000.
				   	        If you specify a value that is higher than this number,
				   	        your search results will be truncated.
	includeActive   boolean If true, then active users are included in the results (default true)
	includeInactive boolean If true, then inactive users are included in the results (default false)

*/
func (j *Jira) SearchUser(username string, startAt int, maxResults int, includeActive bool, includeInactive bool) (c []byte, err error) {
	uri := j.BaseUrl + j.ApiPath + user_uri + "?username=" + username
	contents, err := j.getRequest(uri)
	if err != nil {
		return nil, err
	}
	// @todo
	return contents, nil
}

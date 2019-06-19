package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"
)

// код писать тут

type UserFromXML struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

type UsersData struct {
	Users []UserFromXML `xml:"row"`
}

type Filter struct {
	Limit      int
	Offset     int
	Query      string
	OrderField string
	OrderBy    int
}

func getUsers(filter Filter) (users []User) {
	file, err := os.Open("dataset.xml")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	v := new(UsersData)
	err = xml.Unmarshal(b, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	for _, u := range v.Users {
		user := User{
			Id:     u.ID,
			Name:   u.FirstName + " " + u.LastName,
			Age:    u.Age,
			About:  u.About,
			Gender: u.Gender,
		}
		if filter.Query == user.Name || filter.Query == user.About {
			users = append(users, user)
		} else if filter.Query == "" {
			users = append(users, user)
		}
	}
	if filter.OrderField == "Id" {
		sort.Slice(users, func(i, j int) (out bool) {
			if filter.OrderBy == OrderByAsc {
				out = users[i].Id > users[j].Id
			} else {
				out = users[i].Id < users[j].Id
			}
			return
		})
	} else if filter.OrderField == "Age" {
		sort.Slice(users, func(i, j int) (out bool) {
			if filter.OrderBy == OrderByAsc {
				out = users[i].Age > users[j].Age
			} else {
				out = users[i].Age < users[j].Age
			}
			return
		})
	} else {
		sort.Slice(users, func(i, j int) (out bool) {
			if filter.OrderBy == OrderByAsc {
				out = users[i].Name > users[j].Name
			} else {
				out = users[i].Name < users[j].Name
			}
			return
		})
	}
	limit := filter.Limit
	if len(users)-filter.Offset-filter.Limit < 0 {
		limit = len(users) - filter.Offset
	}
	users = users[filter.Offset : filter.Offset+limit]
	return users
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	var filter Filter
	var err error
	filter.Limit, err = strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "Parametr \"limit\" is bad"}`)
	}
	filter.Offset, err = strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "Parametr \"offset\" is bad"}`)
	}
	filter.Query = r.FormValue("query")
	filter.OrderField = r.FormValue("order_field")
	if filter.OrderField != "Id" && filter.OrderField != "Age" &&
		filter.OrderField != "Name" && filter.OrderField != "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "Parametr \"order_field\" is bad"}`)
	}
	filter.OrderBy, err = strconv.Atoi(r.FormValue("order_by"))
	if err != nil || filter.OrderBy < -1 || filter.OrderBy > 1 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "Parametr \"order_by\" is bad"}`)
	}
	users := getUsers(filter)
	json, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(json))
}

func TestSearchServer(t *testing.T) {
	cases := []SearchRequest{
		SearchRequest{
			Offset: 0,
			Limit:  0,
		},
		SearchRequest{
			Limit: -1,
		},
		SearchRequest{
			Limit: 30,
		},
		SearchRequest{
			Offset: -1,
		},
		SearchRequest{
			Limit: 20,
			Query: "Owen Lynn",
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	for _, item := range cases {
		c := &SearchClient{
			AccessToken: "aaa",
			URL:         ts.URL,
		}
		_, err := c.FindUsers(item)

		if err != nil {
			fmt.Println(err)
		}
	}
}

func TestErrTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerErrTimeout))
	defer ts.Close()
	c := &SearchClient{
		URL: ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Error: Timeout")
	}

}

func SearchServerErrTimeout(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 2)
	w.WriteHeader(http.StatusOK)
}

func TestErrUnknown(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerErrUnknown))
	defer ts.Close()
	c := &SearchClient{
		URL: "",
	}
	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Error: Unknown")
	}

}

func SearchServerErrUnknown(w http.ResponseWriter, r *http.Request) {
}

func TestErrStatusUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerErrStatusUnauthorized))
	defer ts.Close()
	c := &SearchClient{
		URL: ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Error: StatusUnauthorized")
	}

}

func SearchServerErrStatusUnauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
}

func TestErrStatusInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerErrStatusInternalServerError))
	defer ts.Close()
	c := &SearchClient{
		URL: ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Error: StatusInternalServerError")
	}

}

func SearchServerErrStatusInternalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TestErrStatusBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerErrStatusBadRequest))
	defer ts.Close()
	c := &SearchClient{
		URL: ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Error: StatusBadRequest")
	}
}

func SearchServerErrStatusBadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func TestErrStatusBadRequestUnpackJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerErrStatusBadRequestUnpackJSON))
	defer ts.Close()
	c := &SearchClient{
		URL: ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Error: StatusBadRequestUnpackJSON")
	}
}

func SearchServerErrStatusBadRequestUnpackJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, `{}`)
}

func TestErrStatusBadRequestErrorBadOrderField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerErrStatusBadRequestErrorBadOrderField))
	defer ts.Close()
	c := &SearchClient{
		URL: ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Error: StatusBadRequestErrorBadOrderField")
	}
}

func SearchServerErrStatusBadRequestErrorBadOrderField(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, `{"Error":"ErrorBadOrderField"}`)
}

func TestErrUnpackJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerErrUnpackJSON))
	defer ts.Close()
	c := &SearchClient{
		URL: ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Error: UnpackJSON")
	}
}

func SearchServerErrUnpackJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `aa}`)
}

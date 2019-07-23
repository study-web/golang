package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"	
)

func (h *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	field := ""
	s := ProfileParams{}
	field = r.FormValue("login")
	s.Login = field
	if s.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"login must me not empty\"}")
		return
	}
	var ctx context.Context
	varUser, err := h.Profile(ctx, s)
	if err != nil {
		switch err.(type) {
		case ApiError:
			//fmt.Printf("\nAAAAAAA %T\n", v)
			err := err.(ApiError)
			w.WriteHeader(err.HTTPStatus)
			io.WriteString(w, "{\"error\": \""+err.Error()+"\"}")
			return
		default:
			//fmt.Printf("\nBBBBBBBBB %T\n", v)
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "{\"error\": \"bad user\"}")
			return
		}
	}
	answer, err := json.Marshal(map[string]interface{}{"error": "", "response": varUser})
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Write(answer)
}

func (h *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	field := ""
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "{\"error\": \"unauthorized\"}")
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		io.WriteString(w, "{\"error\": \"bad method\"}")
		return
	}			
	s := CreateParams{}
	field = r.FormValue("login")
	s.Login = field
	if s.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"login must me not empty\"}")
		return
	}
	if len(s.Login) < 10 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"login len must be >= 10\"}")
		return
	}
	field = r.FormValue("full_name")
	s.Name = field
	field = r.FormValue("status")
	if field == "" {
		field = "user"
	}
	s.Status = field
	if s.Status != "user" && s.Status != "moderator" && s.Status != "admin" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"status must be one of [user, moderator, admin]\"}")
		return
	}
	field = r.FormValue("age")
	intField, err := strconv.Atoi(field)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"age must be int\"}")
		return		
	}
	s.Age = intField
	if s.Age < 0 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"age must be >= 0\"}")
		return
	}
	if s.Age > 128 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"age must be <= 128\"}")
		return
	}
	var ctx context.Context
	varNewUser, err := h.Create(ctx, s)
	if err != nil {
		switch err.(type) {
		case ApiError:
			//fmt.Printf("\nAAAAAAA %T\n", v)
			err := err.(ApiError)
			w.WriteHeader(err.HTTPStatus)
			io.WriteString(w, "{\"error\": \""+err.Error()+"\"}")
			return
		default:
			//fmt.Printf("\nBBBBBBBBB %T\n", v)
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "{\"error\": \"bad user\"}")
			return
		}
	}
	answer, err := json.Marshal(map[string]interface{}{"error": "", "response": varNewUser})
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Write(answer)
}

func (h *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	field := ""
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "{\"error\": \"unauthorized\"}")
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		io.WriteString(w, "{\"error\": \"bad method\"}")
		return
	}			
	s := OtherCreateParams{}
	field = r.FormValue("username")
	s.Username = field
	if s.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"username must me not empty\"}")
		return
	}
	if len(s.Username) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"username len must be >= 3\"}")
		return
	}
	field = r.FormValue("account_name")
	s.Name = field
	field = r.FormValue("class")
	if field == "" {
		field = "warrior"
	}
	s.Class = field
	if s.Class != "warrior" && s.Class != "sorcerer" && s.Class != "rouge" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"class must be one of [warrior, sorcerer, rouge]\"}")
		return
	}
	field = r.FormValue("level")
	intField, err := strconv.Atoi(field)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"level must be int\"}")
		return		
	}
	s.Level = intField
	if s.Level < 1 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"level must be >= 1\"}")
		return
	}
	if s.Level > 50 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "{\"error\": \"level must be <= 50\"}")
		return
	}
	var ctx context.Context
	varOtherUser, err := h.Create(ctx, s)
	if err != nil {
		switch err.(type) {
		case ApiError:
			//fmt.Printf("\nAAAAAAA %T\n", v)
			err := err.(ApiError)
			w.WriteHeader(err.HTTPStatus)
			io.WriteString(w, "{\"error\": \""+err.Error()+"\"}")
			return
		default:
			//fmt.Printf("\nBBBBBBBBB %T\n", v)
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "{\"error\": \"bad user\"}")
			return
		}
	}
	answer, err := json.Marshal(map[string]interface{}{"error": "", "response": varOtherUser})
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Write(answer)
}

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { 
	case "/user/profile":
		h.handlerProfile(w, r)
	case "/user/create":
		h.handlerCreate(w, r)
    default:
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "{\"error\": \"unknown method\"}")
    }
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { 
	case "/user/create":
		h.handlerCreate(w, r)
    default:
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "{\"error\": \"unknown method\"}")
    }
}

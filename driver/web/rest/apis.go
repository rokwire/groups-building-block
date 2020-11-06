package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"groups/core"
	"groups/core/model"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/go-playground/validator.v9"
)

//ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app *core.Application
}

//Version gives the service version
// @Description Gives the service version.
// @ID Version
// @Produce plain
// @Success 200 {string} v1.1.0
// @Router /version [get]
func (h ApisHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.app.Services.GetVersion()))
}

//GetGroupCategories gives all group categories
// @Description Gives all group categories.
// @ID GetGroupCategories
// @Accept  json
// @Success 200 {array} string
// @Security APIKeyAuth
// @Router /api/group-categories [get]
func (h *ApisHandler) GetGroupCategories(w http.ResponseWriter, r *http.Request) {
	groupCategories, err := h.app.Services.GetGroupCategories()
	if err != nil {
		log.Println("Error on getting the group categories")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(groupCategories) == 0 {
		groupCategories = make([]string, 0)
	}

	data, err := json.Marshal(groupCategories)
	if err != nil {
		log.Println("Error on marshal the group categories")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type createGroupRequest struct {
	Title           string   `json:"title" validate:"required"`
	Description     *string  `json:"description"`
	Category        string   `json:"category" validate:"required"`
	Tags            []string `json:"tags"`
	Privacy         string   `json:"privacy" validate:"required,oneof=public private"`
	CreatorName     string   `json:"creator_name"`
	CreatorEmail    string   `json:"creator_email"`
	CreatorPhotoURL string   `json:"creator_photo_url"`
} //@name createGroupRequest

//CreateGroup creates a group
// @Description Creates a group. The user must be part of urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire groups access. Title must be a unique. Category must be one of the categories list. Privacy can be public or private
// @ID CreateGroup
// @Accept json
// @Produce json
// @Param data body createGroupRequest true "body data"
// @Success 200 {object} createResponse
// @Security AppUserAuth
// @Router /api/groups [post]
func (h *ApisHandler) CreateGroup(current *model.User, w http.ResponseWriter, r *http.Request) {
	if !current.IsMemberOfGroup("urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire groups access") {
		log.Printf("%s is not allowed to create a group", current.Email)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal create a group - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createGroupRequest
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create group data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//validate
	validate := validator.New()
	err = validate.Struct(requestData)
	if err != nil {
		log.Printf("Error on validating create group data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	title := requestData.Title
	description := requestData.Description
	category := requestData.Category
	tags := requestData.Tags
	privacy := requestData.Privacy
	creatorName := requestData.CreatorName
	creatorEmail := requestData.CreatorEmail
	creatorPhotoURL := requestData.CreatorPhotoURL

	insertedID, err := h.app.Services.CreateGroup(*current, title, description, category, tags, privacy,
		creatorName, creatorEmail, creatorPhotoURL)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(createResponse{InsertedID: *insertedID})
	if err != nil {
		log.Println("Error on marshal create group response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type Resp struct {
	List []model.Group
}

func (p Resp) MarshalJSON() ([]byte, error) {
	/*millis, err := time.Parse(time.RFC3339, p.CreatedAt)
	if err != nil {
		return nil, err
	}

	res := ProfileJSON{
		Name:      p.Name,
		Address:   p.Address,
		CreatedAt: millis.UnixNano() / 1000000,
	} */
	/*
		item := p.List[0]

		type Alias model.Group
		al := item

		zz := &struct {
			LastSeen int64 `json:"lastSeen"`
			//Alias
			model.Group Alias
		}{
			LastSeen: 1234,
			Alias:    al,
		}
		return json.Marshal(zz) */
	/*
			result := make([]model.Group)
		for _, current := p.List {

		} */

	/*	res := ProfileJSON{
		Name:      p.Name,
		Address:   p.Address,
		CreatedAt: millis.UnixNano() / 1000000,
	} */

	buffer := bytes.NewBufferString("[")
	length := len(p.List)
	count := 0
	for _, value := range p.List {

		/*v := reflect.ValueOf(value)
		t := v.Type()
		sf := make([]reflect.StructField, 0)
		for i := 0; i < t.NumField(); i++ {
			structField := t.Field(i)
			fmt.Printf("tag:%s\tname:%s\n", structField.Tag, structField.Name)

			if structField.Name == "ID" {
				//	sf[i].Tag = `json:"name"`

				sf = append(sf, structField)
			}
		}
		newType := reflect.StructOf(sf)
		newValue := v.Convert(newType)
		nnv := newValue.Interface()
		log.Println(nnv)

		jsonValue := "12345" */

		groupWrapper := GroupWrapperResp{value}
		jsonValue, err := json.Marshal(groupWrapper)
		if err != nil {
			return nil, err
		}
		buffer.WriteString(fmt.Sprintf("%s", string(jsonValue)))
		count++
		if count < length {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("]")
	return buffer.Bytes(), nil

	//	return json.Marshal(p.List)
}

type GroupWrapperResp struct {
	Group model.Group
}

func (gr GroupWrapperResp) MarshalJSON() ([]byte, error) {
	return json.Marshal(gr.Group)
}

//GetGroups gets groups
func (h *ApisHandler) GetGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.app.Services.GetGroups(nil)
	if err != nil {
		log.Printf("error getting groups - %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//TODO nil
	//	group := groups[0]

	/*
		type alias struct {
			ID       string `json:"id"`
			Category string `json:"category"` //one of the enums categories list
			Title    string `json:"title"`
		}
		var a alias = alias(group)
		//	return json.Marshal(&a)
	*/
	////////////////////////

	//in := &MyData{One: 1, Two: "second"}

	var inInterface []map[string]interface{}
	inrec, _ := json.Marshal(groups)
	json.Unmarshal(inrec, &inInterface)

	// iterate through inrecs
	for _, groupMap := range inInterface {
		//fmt.Println("KV Pair: ", i, groupMap)
		members := groupMap["members"].([]interface{})

		log.Println(members)
		//delete(map(members), "user")
		//	if field == "members" {
		//
		//	}
	}

	////////////////////

	/*value := reflect.ValueOf(group)
	t := value.Type()
	sf := make([]reflect.StructField, 0)
	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		fmt.Printf("tag:%s\tname:%s\n", structField.Tag, structField.Name)

		if structField.Name == "ID" {
			//	sf[i].Tag = `json:"name"`

			sf = append(sf, structField)
		}
	}
	newType := reflect.StructOf(sf)
	newValue := value.Convert(newType)
	nnv := newValue.Interface() */
	//json.Marshal(newValue.Interface())

	resp := Resp{List: groups}
	//resp := groups

	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("Error on marshal the groups items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//JustMixed test TODO
func (h *ApisHandler) JustMixed(current *model.User, w http.ResponseWriter, r *http.Request) {
	log.Println("JustMixed")
}

//NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application) *ApisHandler {
	return &ApisHandler{app: app}
}

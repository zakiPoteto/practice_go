package testdata

import "encoding/json"

type TaskInput struct {
	Title  string `json:"title,omitempty"`
	Status string `json:"status,omitempty"`
}

var DefaultTaskInputs = []TaskInput{
	{Title: "doing homework", Status: "done"},
	{Title: "go to shopping", Status: "todo"},
	{Title: "cleaning", Status: "todo"},
	{Title: "attend meeting", Status: "todo"},
	{Title: "go to work", Status: "done"},
}

func TaskJSONBody(input TaskInput) ([]byte, error) {
	return json.Marshal(input)
}

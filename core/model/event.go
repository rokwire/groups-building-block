package model

import "time"

//Event represents event entity
type Event struct {
	EventID     string    `json:"event_id"`
	Group       Group     `json:"group"`
	DateCreated time.Time `json:"date_created"`
	Comments    []Comment `json:"comments"`
} // @name Event

//Comment represents comment entity
type Comment struct {
	Member      Member    `json:"member"`
	Text        string    `json:"text"`
	DateCreated time.Time `json:"date_created"`
} // @name Comment

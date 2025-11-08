package domain

import "time"

type Nationality struct {
	ID   int32
	Name string
	Code *string
}

type FamilyMember struct {
	ID         int32
	CustomerID int32
	Relation   string
	Name       string
	Dob        time.Time
}

type Customer struct {
	ID            int32
	NationalityID int32
	Name          string
	Dob           time.Time
	PhoneNum      string
	Email         string
	Family        []FamilyMember
}

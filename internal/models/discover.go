package models

type UserDetailsES struct {
	UserID   string         `json:"UserID"`
	Name     string         `json:"name"`
	Gender   string         `json:"gender"`
	Age      int            `json:"age"`
	Location UserLocationES `json:"location"`
}

type UserLocationES struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type DiscoverFilters struct {
	UserID      string
	Gender      string `json:"gender,omitempty"`
	MaxAge      int    `json:"maxAge,omitempty"`
	MinAge      int    `json:"minAge,omitempty"`
	MaxLocation int    `json:"maxLocation,omitempty"`
}

type DiscoverReturn struct {
	UsersES []UserDetailsES `json:"users"`
}

package application

type CreateTenantInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type UpdateTenantInput struct {
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Version int64  `json:"version"`
}

type CreateSiteInput struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Timezone string `json:"timezone"`
}

type UpdateSiteInput struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Timezone string `json:"timezone"`
	Version  int64  `json:"version"`
}

type CreateResult struct {
	ID      string `json:"id"`
	Created bool   `json:"created"`
	Version int64  `json:"version"`
}

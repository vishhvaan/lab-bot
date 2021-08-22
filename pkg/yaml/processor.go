package yaml

type Secrets struct {
	AppTolken string
	BotTolken string
}

type Members struct {
	FirstName string
	LastName string
	Subgroup string
	UserID string
	Roles []string
}
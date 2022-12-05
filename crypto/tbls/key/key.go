package key

type PriShare struct {
	Index    int    `json:"Index"`
	PriBytes []byte `json:"PriBytes"`
}

type PubShare struct {
	Index    int    `json:"Index"`
	PubBytes []byte `json:"PubBytes"`
}

package v1

type EntryResponse struct {
	Key, Value string
}

type CountResponse struct {
	Size int
}

type PutEntryRequest struct {
	Value string
}

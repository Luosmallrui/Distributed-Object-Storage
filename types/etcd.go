package types

type KvStorage struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Storage struct {
	StorageNode []string `json:"storage_node"`
	ManageNode  []string `json:"manage_node"`
}

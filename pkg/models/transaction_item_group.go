package models

// TransactionItemGroup represents transaction item group data stored in database
type TransactionItemGroup struct {
	ItemGroupId     int64  `xorm:"PK"`
	Uid             int64  `xorm:"INDEX(IDX_item_group_uid_deleted_order) NOT NULL"`
	Deleted         bool   `xorm:"INDEX(IDX_item_group_uid_deleted_order) NOT NULL"`
	Name            string `xorm:"VARCHAR(64) NOT NULL"`
	DisplayOrder    int32  `xorm:"INDEX(IDX_item_group_uid_deleted_order) NOT NULL"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

// TransactionItemGroupGetRequest represents all parameters of transaction item group getting request
type TransactionItemGroupGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// TransactionItemGroupCreateRequest represents all parameters of transaction item group creation request
type TransactionItemGroupCreateRequest struct {
	Name string `json:"name" binding:"required,notBlank,max=64"`
}

// TransactionItemGroupModifyRequest represents all parameters of transaction item group modification request
type TransactionItemGroupModifyRequest struct {
	Id   int64  `json:"id,string" binding:"required,min=1"`
	Name string `json:"name" binding:"required,notBlank,max=64"`
}

// TransactionItemGroupMoveRequest represents all parameters of transaction item group moving request
type TransactionItemGroupMoveRequest struct {
	NewDisplayOrders []*TransactionItemGroupNewDisplayOrderRequest `json:"newDisplayOrders" binding:"required,min=1"`
}

// TransactionItemGroupNewDisplayOrderRequest represents a data pair of id and display order
type TransactionItemGroupNewDisplayOrderRequest struct {
	Id           int64 `json:"id,string" binding:"required,min=1"`
	DisplayOrder int32 `json:"displayOrder"`
}

// TransactionItemGroupDeleteRequest represents all parameters of transaction item group deleting request
type TransactionItemGroupDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// TransactionItemGroupInfoResponse represents a view-object of transaction item group
type TransactionItemGroupInfoResponse struct {
	Id           int64  `json:"id,string"`
	Name         string `json:"name"`
	DisplayOrder int32  `json:"displayOrder"`
}

// ToTransactionItemGroupInfoResponse returns a view-object according to database model
func (t *TransactionItemGroup) ToTransactionItemGroupInfoResponse() *TransactionItemGroupInfoResponse {
	return &TransactionItemGroupInfoResponse{
		Id:           t.ItemGroupId,
		Name:         t.Name,
		DisplayOrder: t.DisplayOrder,
	}
}

// TransactionItemGroupInfoResponseSlice represents the slice data structure of TransactionItemGroupInfoResponse
type TransactionItemGroupInfoResponseSlice []*TransactionItemGroupInfoResponse

// Len returns the count of items
func (s TransactionItemGroupInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s TransactionItemGroupInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s TransactionItemGroupInfoResponseSlice) Less(i, j int) bool {
	return s[i].DisplayOrder < s[j].DisplayOrder
}

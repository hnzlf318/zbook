package models

// TransactionItem represents transaction item data stored in database
type TransactionItem struct {
	ItemId          int64  `xorm:"PK"`
	Uid             int64  `xorm:"INDEX(IDX_item_uid_deleted_group_order) NOT NULL"`
	Deleted         bool   `xorm:"INDEX(IDX_item_uid_deleted_group_order) NOT NULL"`
	ItemGroupId     int64  `xorm:"INDEX(IDX_item_uid_deleted_group_order) NOT NULL DEFAULT 0"`
	Name            string `xorm:"VARCHAR(64) NOT NULL"`
	DisplayOrder    int32  `xorm:"INDEX(IDX_item_uid_deleted_group_order) NOT NULL"`
	Hidden          bool   `xorm:"NOT NULL"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

// TransactionItemGetRequest represents all parameters of transaction item getting request
type TransactionItemGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// TransactionItemCreateRequest represents all parameters of transaction item creation request
type TransactionItemCreateRequest struct {
	GroupId int64  `json:"groupId,string"`
	Name    string `json:"name" binding:"required,notBlank,max=64"`
}

// TransactionItemCreateBatchRequest represents all parameters of transaction item batch creation request
type TransactionItemCreateBatchRequest struct {
	Items      []*TransactionItemCreateRequest `json:"items" binding:"required"`
	GroupId    int64                           `json:"groupId,string"`
	SkipExists bool                            `json:"skipExists"`
}

// TransactionItemModifyRequest represents all parameters of transaction item modification request
type TransactionItemModifyRequest struct {
	Id      int64  `json:"id,string" binding:"required,min=1"`
	GroupId int64  `json:"groupId,string"`
	Name    string `json:"name" binding:"required,notBlank,max=64"`
}

// TransactionItemHideRequest represents all parameters of transaction item hiding request
type TransactionItemHideRequest struct {
	Id     int64 `json:"id,string" binding:"required,min=1"`
	Hidden bool  `json:"hidden"`
}

// TransactionItemMoveRequest represents all parameters of transaction item moving request
type TransactionItemMoveRequest struct {
	NewDisplayOrders []*TransactionItemNewDisplayOrderRequest `json:"newDisplayOrders" binding:"required,min=1"`
}

// TransactionItemNewDisplayOrderRequest represents a data pair of id and display order
type TransactionItemNewDisplayOrderRequest struct {
	Id           int64 `json:"id,string" binding:"required,min=1"`
	DisplayOrder int32 `json:"displayOrder"`
}

// TransactionItemDeleteRequest represents all parameters of transaction item deleting request
type TransactionItemDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// TransactionItemInfoResponse represents a view-object of transaction item
type TransactionItemInfoResponse struct {
	Id           int64  `json:"id,string"`
	Name         string `json:"name"`
	ItemGroupId  int64  `json:"groupId,string"`
	DisplayOrder int32  `json:"displayOrder"`
	Hidden       bool   `json:"hidden"`
}

// FillFromOtherItem fills all the fields in this current item from other transaction item
func (t *TransactionItem) FillFromOtherItem(item *TransactionItem) {
	t.ItemId = item.ItemId
	t.Uid = item.Uid
	t.Deleted = item.Deleted
	t.Name = item.Name
	t.ItemGroupId = item.ItemGroupId
	t.DisplayOrder = item.DisplayOrder
	t.Hidden = item.Hidden
	t.CreatedUnixTime = item.CreatedUnixTime
	t.UpdatedUnixTime = item.UpdatedUnixTime
	t.DeletedUnixTime = item.DeletedUnixTime
}

// ToTransactionItemInfoResponse returns a view-object according to database model
func (t *TransactionItem) ToTransactionItemInfoResponse() *TransactionItemInfoResponse {
	return &TransactionItemInfoResponse{
		Id:           t.ItemId,
		Name:         t.Name,
		ItemGroupId:  t.ItemGroupId,
		DisplayOrder: t.DisplayOrder,
		Hidden:       t.Hidden,
	}
}

// TransactionItemInfoResponseSlice represents the slice data structure of TransactionItemInfoResponse
type TransactionItemInfoResponseSlice []*TransactionItemInfoResponse

// Len returns the count of items
func (s TransactionItemInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s TransactionItemInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s TransactionItemInfoResponseSlice) Less(i, j int) bool {
	if s[i].ItemGroupId != s[j].ItemGroupId {
		return s[i].ItemGroupId < s[j].ItemGroupId
	}

	return s[i].DisplayOrder < s[j].DisplayOrder
}

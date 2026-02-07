package models

// TransactionItemIndex represents transaction and transaction item relation stored in database
type TransactionItemIndex struct {
	ItemIndexId      int64 `xorm:"PK"`
	Uid              int64 `xorm:"INDEX(IDX_transaction_item_index_uid_deleted_item_id_transaction_id) INDEX(IDX_transaction_item_index_uid_deleted_transaction_time_item_id) INDEX(IDX_transaction_item_index_uid_deleted_transaction_id)"`
	Deleted          bool  `xorm:"INDEX(IDX_transaction_item_index_uid_deleted_item_id_transaction_id) INDEX(IDX_transaction_item_index_uid_deleted_transaction_time_item_id) INDEX(IDX_transaction_item_index_uid_deleted_transaction_id) NOT NULL"`
	TransactionTime  int64 `xorm:"INDEX(IDX_transaction_item_index_uid_deleted_transaction_time_item_id) NOT NULL"`
	ItemId           int64 `xorm:"INDEX(IDX_transaction_item_index_uid_deleted_item_id_transaction_id) INDEX(IDX_transaction_item_index_uid_deleted_transaction_time_item_id)"`
	TransactionId    int64 `xorm:"INDEX(IDX_transaction_item_index_uid_deleted_item_id_transaction_id) INDEX(IDX_transaction_item_index_uid_deleted_transaction_id)"`
	CreatedUnixTime  int64
	UpdatedUnixTime  int64
	DeletedUnixTime  int64
}

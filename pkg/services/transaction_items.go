package services

import (
	"strings"
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// TransactionItemService represents transaction item service
type TransactionItemService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction item service singleton instance
var (
	TransactionItems = &TransactionItemService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetTotalItemCountByUid returns total item count of user
func (s *TransactionItemService) GetTotalItemCountByUid(c core.Context, uid int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	count, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Count(&models.TransactionItem{})

	return count, err
}

// GetAllItemsByUid returns all transaction item models of user
func (s *TransactionItemService) GetAllItemsByUid(c core.Context, uid int64) ([]*models.TransactionItem, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var items []*models.TransactionItem
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Find(&items)

	return items, err
}

// GetItemByItemId returns a transaction item model according to transaction item id
func (s *TransactionItemService) GetItemByItemId(c core.Context, uid int64, itemId int64) (*models.TransactionItem, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if itemId <= 0 {
		return nil, errs.ErrTransactionItemIdInvalid
	}

	item := &models.TransactionItem{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(itemId).Where("uid=? AND deleted=?", uid, false).Get(item)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrTransactionItemNotFound
	}

	return item, nil
}

// GetItemsByItemIds returns transaction item models according to transaction item ids
func (s *TransactionItemService) GetItemsByItemIds(c core.Context, uid int64, itemIds []int64) (map[int64]*models.TransactionItem, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if itemIds == nil {
		return nil, errs.ErrTransactionItemIdInvalid
	}

	var items []*models.TransactionItem
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("item_id", itemIds).Find(&items)

	if err != nil {
		return nil, err
	}

	itemMap := s.GetItemMapByList(items)
	return itemMap, err
}

// GetMaxDisplayOrder returns the max display order
func (s *TransactionItemService) GetMaxDisplayOrder(c core.Context, uid int64, itemGroupId int64) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	item := &models.TransactionItem{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "display_order").Where("uid=? AND deleted=? AND item_group_id=?", uid, false, itemGroupId).OrderBy("display_order desc").Limit(1).Get(item)

	if err != nil {
		return 0, err
	}

	if has {
		return item.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// GetAllItemIdsOfTransactions returns transaction item ids for given transactions
func (s *TransactionItemService) GetAllItemIdsOfTransactions(c core.Context, uid int64, transactionIds []int64) (map[int64][]int64, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var itemIndexes []*models.TransactionItemIndex
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("transaction_id", transactionIds).OrderBy("transaction_id asc, item_index_id asc").Find(&itemIndexes)

	allTransactionItemIds := s.GetGroupedTransactionItemIds(itemIndexes)

	return allTransactionItemIds, err
}

// GetGroupedTransactionItemIds returns a map of transaction item ids grouped by transaction id
func (s *TransactionItemService) GetGroupedTransactionItemIds(itemIndexes []*models.TransactionItemIndex) map[int64][]int64 {
	allTransactionItemIds := make(map[int64][]int64)

	for i := 0; i < len(itemIndexes); i++ {
		itemIndex := itemIndexes[i]

		var transactionItemIds []int64

		if _, exists := allTransactionItemIds[itemIndex.TransactionId]; exists {
			transactionItemIds = allTransactionItemIds[itemIndex.TransactionId]
		}

		transactionItemIds = append(transactionItemIds, itemIndex.ItemId)
		allTransactionItemIds[itemIndex.TransactionId] = transactionItemIds
	}

	for _, itemIds := range allTransactionItemIds {
		utils.Int64Sort(itemIds)
	}

	return allTransactionItemIds
}

// CreateItem saves a new transaction item model to database
func (s *TransactionItemService) CreateItem(c core.Context, item *models.TransactionItem) error {
	if item.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	exists, err := s.ExistsItemName(c, item.Uid, item.Name)

	if err != nil {
		return err
	} else if exists {
		return errs.ErrTransactionItemNameAlreadyExists
	}

	item.ItemId = s.GenerateUuid(uuid.UUID_TYPE_ITEM)

	if item.ItemId < 1 {
		return errs.ErrSystemIsBusy
	}

	item.Deleted = false
	item.CreatedUnixTime = time.Now().Unix()
	item.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(item.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(item)
		return err
	})
}

// CreateItems saves a few transaction item models to database
func (s *TransactionItemService) CreateItems(c core.Context, uid int64, items []*models.TransactionItem, skipExists bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	allItemNames := make([]string, len(items))

	for i := 0; i < len(items); i++ {
		allItemNames[i] = items[i].Name
	}

	var existItems []*models.TransactionItem
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("name", allItemNames).Find(&existItems)

	if err != nil {
		return err
	} else if !skipExists && len(existItems) > 0 {
		return errs.ErrTransactionItemNameAlreadyExists
	}

	existsNameItemMap := make(map[string]*models.TransactionItem, len(existItems))

	for i := 0; i < len(existItems); i++ {
		item := existItems[i]
		existsNameItemMap[item.Name] = item
	}

	newItems := make([]*models.TransactionItem, 0, len(items)-len(existItems))

	for i := 0; i < len(items); i++ {
		item := items[i]
		existsItem, exists := existsNameItemMap[item.Name]

		if exists {
			item.FillFromOtherItem(existsItem)
			continue
		}

		newItems = append(newItems, item)
	}

	itemUuids := s.GenerateUuids(uuid.UUID_TYPE_ITEM, uint16(len(newItems)))

	if len(itemUuids) < len(newItems) {
		return errs.ErrSystemIsBusy
	}

	for i := 0; i < len(newItems); i++ {
		item := newItems[i]
		item.ItemId = itemUuids[i]
		item.Deleted = false
		item.CreatedUnixTime = time.Now().Unix()
		item.UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(newItems); i++ {
			item := newItems[i]
			_, err := sess.Insert(item)

			if err != nil {
				return err
			}
		}

		return nil
	})
}

// ModifyItem saves an existed transaction item model to database
func (s *TransactionItemService) ModifyItem(c core.Context, item *models.TransactionItem, itemNameChanged bool) error {
	if item.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if itemNameChanged {
		exists, err := s.ExistsItemName(c, item.Uid, item.Name)

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionItemNameAlreadyExists
		}
	}

	item.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(item.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(item.ItemId).Cols("name", "item_group_id", "display_order", "updated_unix_time").Where("uid=? AND deleted=?", item.Uid, false).Update(item)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionItemNotFound
		}

		return err
	})
}

// HideItem updates hidden field of given transaction items
func (s *TransactionItemService) HideItem(c core.Context, uid int64, ids []int64, hidden bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionItem{
		Hidden:          hidden,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("hidden", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).In("item_id", ids).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionItemNotFound
		}

		return err
	})
}

// ModifyItemDisplayOrders updates display order of given transaction items
func (s *TransactionItemService) ModifyItemDisplayOrders(c core.Context, uid int64, items []*models.TransactionItem) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(items); i++ {
		items[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(items); i++ {
			item := items[i]
			updatedRows, err := sess.ID(item.ItemId).Cols("display_order", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(item)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrTransactionItemNotFound
			}
		}

		return nil
	})
}

// DeleteItem deletes an existed transaction item from database
func (s *TransactionItemService) DeleteItem(c core.Context, uid int64, itemId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionItem{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		exists, err := sess.Cols("uid", "item_id").Where("uid=? AND deleted=? AND item_id=?", uid, false, itemId).Limit(1).Exist(&models.TransactionItemIndex{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionItemInUseCannotBeDeleted
		}

		deletedRows, err := sess.ID(itemId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTransactionItemNotFound
		}

		return err
	})
}

// ExistsItemName returns whether the given item name exists
func (s *TransactionItemService) ExistsItemName(c core.Context, uid int64, name string) (bool, error) {
	if name == "" {
		return false, errs.ErrTransactionItemNameIsEmpty
	}

	return s.UserDataDB(uid).NewSession(c).Cols("name").Where("uid=? AND deleted=? AND name=?", uid, false, name).Exist(&models.TransactionItem{})
}

// GetItemMapByList returns a transaction item map by a list
func (s *TransactionItemService) GetItemMapByList(items []*models.TransactionItem) map[int64]*models.TransactionItem {
	itemMap := make(map[int64]*models.TransactionItem)

	for i := 0; i < len(items); i++ {
		item := items[i]
		itemMap[item.ItemId] = item
	}
	return itemMap
}

// DeleteAllItemIndexes soft-deletes all transaction item indexes for user
func (s *TransactionItemService) DeleteAllItemIndexes(c core.Context, uid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionItemIndex{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	_, err := s.UserDataDB(uid).NewSession(c).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

	return err
}

// DeleteAllItems soft-deletes all transaction items for user
func (s *TransactionItemService) DeleteAllItems(c core.Context, uid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionItem{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	_, err := s.UserDataDB(uid).NewSession(c).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

	return err
}

// GetItemIds converts a comma-separated string of item ids into a slice of int64
func (s *TransactionItemService) GetItemIds(itemIds string) ([]int64, error) {
	if itemIds == "" || itemIds == "0" {
		return nil, nil
	}

	requestItemIds, err := utils.StringArrayToInt64Array(strings.Split(itemIds, ","))

	if err != nil {
		return nil, errs.Or(err, errs.ErrTransactionItemIdInvalid)
	}

	return requestItemIds, nil
}

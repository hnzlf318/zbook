package services

import (
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// TransactionItemGroupService represents transaction item group service
type TransactionItemGroupService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction item group service singleton instance
var (
	TransactionItemGroups = &TransactionItemGroupService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetAllItemGroupsByUid returns all transaction item group models of user
func (s *TransactionItemGroupService) GetAllItemGroupsByUid(c core.Context, uid int64) ([]*models.TransactionItemGroup, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var itemGroups []*models.TransactionItemGroup
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Find(&itemGroups)

	return itemGroups, err
}

// GetItemGroupByItemGroupId returns a transaction item group model according to transaction item group id
func (s *TransactionItemGroupService) GetItemGroupByItemGroupId(c core.Context, uid int64, itemGroupId int64) (*models.TransactionItemGroup, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if itemGroupId <= 0 {
		return nil, errs.ErrTransactionItemGroupIdInvalid
	}

	itemGroup := &models.TransactionItemGroup{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(itemGroupId).Where("uid=? AND deleted=?", uid, false).Get(itemGroup)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrTransactionItemGroupNotFound
	}

	return itemGroup, nil
}

// GetMaxDisplayOrder returns the max display order
func (s *TransactionItemGroupService) GetMaxDisplayOrder(c core.Context, uid int64) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	itemGroup := &models.TransactionItemGroup{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "display_order").Where("uid=? AND deleted=?", uid, false).OrderBy("display_order desc").Limit(1).Get(itemGroup)

	if err != nil {
		return 0, err
	}

	if has {
		return itemGroup.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// CreateItemGroup saves a new transaction item group model to database
func (s *TransactionItemGroupService) CreateItemGroup(c core.Context, itemGroup *models.TransactionItemGroup) error {
	if itemGroup.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	itemGroup.ItemGroupId = s.GenerateUuid(uuid.UUID_TYPE_ITEM_GROUP)

	if itemGroup.ItemGroupId < 1 {
		return errs.ErrSystemIsBusy
	}

	itemGroup.Deleted = false
	itemGroup.CreatedUnixTime = time.Now().Unix()
	itemGroup.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(itemGroup.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(itemGroup)
		return err
	})
}

// ModifyItemGroup saves an existed transaction item group model to database
func (s *TransactionItemGroupService) ModifyItemGroup(c core.Context, itemGroup *models.TransactionItemGroup) error {
	if itemGroup.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	itemGroup.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(itemGroup.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(itemGroup.ItemGroupId).Cols("name", "updated_unix_time").Where("uid=? AND deleted=?", itemGroup.Uid, false).Update(itemGroup)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionItemGroupNotFound
		}

		return err
	})
}

// ModifyItemGroupDisplayOrders updates display order of given transaction item groups
func (s *TransactionItemGroupService) ModifyItemGroupDisplayOrders(c core.Context, uid int64, itemGroups []*models.TransactionItemGroup) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(itemGroups); i++ {
		itemGroups[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(itemGroups); i++ {
			itemGroup := itemGroups[i]
			updatedRows, err := sess.ID(itemGroup.ItemGroupId).Cols("display_order", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(itemGroup)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrTransactionItemGroupNotFound
			}
		}

		return nil
	})
}

// DeleteItemGroup deletes an existed transaction item group from database
func (s *TransactionItemGroupService) DeleteItemGroup(c core.Context, uid int64, itemGroupId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionItemGroup{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		exists, err := sess.Cols("uid", "deleted").Where("uid=? AND deleted=? AND item_group_id=?", uid, false, itemGroupId).Limit(1).Exist(&models.TransactionItem{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionItemGroupInUseCannotBeDeleted
		}

		deletedRows, err := sess.ID(itemGroupId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTransactionItemGroupNotFound
		}

		return err
	})
}

// DeleteAllItemGroups soft-deletes all transaction item groups for user
func (s *TransactionItemGroupService) DeleteAllItemGroups(c core.Context, uid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TransactionItemGroup{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		exists, err := sess.Cols("uid", "deleted").Where("uid=? AND deleted=? AND item_group_id>?", uid, false, 0).Limit(1).Exist(&models.TransactionItem{})

		if err != nil {
			return err
		} else if exists {
			return errs.ErrTransactionItemGroupInUseCannotBeDeleted
		}

		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		return err
	})
}

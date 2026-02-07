package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// TransactionItemsApi represents transaction item api
type TransactionItemsApi struct {
	items      *services.TransactionItemService
	itemGroups *services.TransactionItemGroupService
}

// Initialize a transaction item api singleton instance
var (
	TransactionItems = &TransactionItemsApi{
		items:      services.TransactionItems,
		itemGroups: services.TransactionItemGroups,
	}
)

// ItemListHandler returns transaction item list of current user
func (a *TransactionItemsApi) ItemListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	items, err := a.items.GetAllItemsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemListHandler] failed to get items for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	itemResps := make(models.TransactionItemInfoResponseSlice, len(items))

	for i := 0; i < len(items); i++ {
		itemResps[i] = items[i].ToTransactionItemInfoResponse()
	}

	sort.Sort(itemResps)

	return itemResps, nil
}

// ItemGetHandler returns one specific transaction item of current user
func (a *TransactionItemsApi) ItemGetHandler(c *core.WebContext) (any, *errs.Error) {
	var itemGetReq models.TransactionItemGetRequest
	err := c.ShouldBindQuery(&itemGetReq)

	if err != nil {
		log.Warnf(c, "[transaction_items.ItemGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	item, err := a.items.GetItemByItemId(c, uid, itemGetReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemGetHandler] failed to get item \"id:%d\" for user \"uid:%d\", because %s", itemGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	itemResp := item.ToTransactionItemInfoResponse()

	return itemResp, nil
}

// ItemCreateHandler saves a new transaction item by request parameters for current user
func (a *TransactionItemsApi) ItemCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var itemCreateReq models.TransactionItemCreateRequest
	err := c.ShouldBindJSON(&itemCreateReq)

	if err != nil {
		log.Warnf(c, "[transaction_items.ItemCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	if itemCreateReq.GroupId > 0 {
		itemGroup, err := a.itemGroups.GetItemGroupByItemGroupId(c, uid, itemCreateReq.GroupId)

		if err != nil {
			log.Errorf(c, "[transaction_items.ItemCreateHandler] failed to get item group \"id:%d\" for user \"uid:%d\", because %s", itemCreateReq.GroupId, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		if itemGroup == nil {
			log.Warnf(c, "[transaction_items.ItemCreateHandler] the item group \"id:%d\" does not exist for user \"uid:%d\"", itemCreateReq.GroupId, uid)
			return nil, errs.ErrTransactionItemGroupNotFound
		}
	}

	maxOrderId, err := a.items.GetMaxDisplayOrder(c, uid, itemCreateReq.GroupId)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemCreateHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	item := a.createNewItemModel(uid, &itemCreateReq, maxOrderId+1)

	err = a.items.CreateItem(c, item)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemCreateHandler] failed to create item \"id:%d\" for user \"uid:%d\", because %s", item.ItemId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_items.ItemCreateHandler] user \"uid:%d\" has created a new item \"id:%d\" successfully", uid, item.ItemId)

	itemResp := item.ToTransactionItemInfoResponse()

	return itemResp, nil
}

// ItemCreateBatchHandler saves some new transaction items by request parameters for current user
func (a *TransactionItemsApi) ItemCreateBatchHandler(c *core.WebContext) (any, *errs.Error) {
	var itemCreateBatchReq models.TransactionItemCreateBatchRequest
	err := c.ShouldBindJSON(&itemCreateBatchReq)

	if err != nil {
		log.Warnf(c, "[transaction_items.ItemCreateBatchHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	for i := 0; i < len(itemCreateBatchReq.Items); i++ {
		if itemCreateBatchReq.Items[i].GroupId != itemCreateBatchReq.GroupId {
			log.Warnf(c, "[transaction_items.ItemCreateBatchHandler] the group id \"%d\" of item#%d is inconsistent with the batch group id \"%d\"", itemCreateBatchReq.Items[i].GroupId, i, itemCreateBatchReq.GroupId)
			return nil, errs.ErrTransactionItemGroupIdInvalid
		}
	}

	uid := c.GetCurrentUid()

	if itemCreateBatchReq.GroupId > 0 {
		itemGroup, err := a.itemGroups.GetItemGroupByItemGroupId(c, uid, itemCreateBatchReq.GroupId)

		if err != nil {
			log.Errorf(c, "[transaction_items.ItemCreateBatchHandler] failed to get item group \"id:%d\" for user \"uid:%d\", because %s", itemCreateBatchReq.GroupId, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		if itemGroup == nil {
			log.Warnf(c, "[transaction_items.ItemCreateBatchHandler] the item group \"id:%d\" does not exist for user \"uid:%d\"", itemCreateBatchReq.GroupId, uid)
			return nil, errs.ErrTransactionItemGroupNotFound
		}
	}

	maxOrderId, err := a.items.GetMaxDisplayOrder(c, uid, itemCreateBatchReq.GroupId)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemCreateBatchHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	items := a.createNewItemModels(uid, &itemCreateBatchReq, maxOrderId+1)

	err = a.items.CreateItems(c, uid, items, itemCreateBatchReq.SkipExists)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemCreateBatchHandler] failed to create items for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_items.ItemCreateBatchHandler] user \"uid:%d\" has created items successfully", uid)

	itemResps := make(models.TransactionItemInfoResponseSlice, len(items))

	for i := 0; i < len(items); i++ {
		itemResps[i] = items[i].ToTransactionItemInfoResponse()
	}

	sort.Sort(itemResps)

	return itemResps, nil
}

// ItemModifyHandler saves an existed transaction item by request parameters for current user
func (a *TransactionItemsApi) ItemModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var itemModifyReq models.TransactionItemModifyRequest
	err := c.ShouldBindJSON(&itemModifyReq)

	if err != nil {
		log.Warnf(c, "[transaction_items.ItemModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	item, err := a.items.GetItemByItemId(c, uid, itemModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemModifyHandler] failed to get item \"id:%d\" for user \"uid:%d\", because %s", itemModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if itemModifyReq.GroupId != item.ItemGroupId && itemModifyReq.GroupId > 0 {
		itemGroup, err := a.itemGroups.GetItemGroupByItemGroupId(c, uid, itemModifyReq.GroupId)

		if err != nil {
			log.Errorf(c, "[transaction_items.ItemModifyHandler] failed to get item group \"id:%d\" for user \"uid:%d\", because %s", itemModifyReq.GroupId, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		if itemGroup == nil {
			log.Warnf(c, "[transaction_items.ItemModifyHandler] the item group \"id:%d\" does not exist for user \"uid:%d\"", itemModifyReq.GroupId, uid)
			return nil, errs.ErrTransactionItemGroupNotFound
		}
	}

	newItem := &models.TransactionItem{
		ItemId:       item.ItemId,
		Uid:          uid,
		Name:         itemModifyReq.Name,
		ItemGroupId:  itemModifyReq.GroupId,
		DisplayOrder: item.DisplayOrder,
	}

	itemNameChanged := newItem.Name != item.Name

	if !itemNameChanged && newItem.ItemGroupId == item.ItemGroupId {
		return nil, errs.ErrNothingWillBeUpdated
	}

	if newItem.ItemGroupId != item.ItemGroupId {
		maxOrderId, err := a.items.GetMaxDisplayOrder(c, uid, newItem.ItemGroupId)

		if err != nil {
			log.Errorf(c, "[transaction_items.ItemModifyHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		newItem.DisplayOrder = maxOrderId + 1
	}

	err = a.items.ModifyItem(c, newItem, itemNameChanged)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemModifyHandler] failed to update item \"id:%d\" for user \"uid:%d\", because %s", itemModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_items.ItemModifyHandler] user \"uid:%d\" has updated item \"id:%d\" successfully", uid, itemModifyReq.Id)

	item.Name = newItem.Name
	item.ItemGroupId = newItem.ItemGroupId
	item.DisplayOrder = newItem.DisplayOrder
	itemResp := item.ToTransactionItemInfoResponse()

	return itemResp, nil
}

// ItemHideHandler hides a transaction item by request parameters for current user
func (a *TransactionItemsApi) ItemHideHandler(c *core.WebContext) (any, *errs.Error) {
	var itemHideReq models.TransactionItemHideRequest
	err := c.ShouldBindJSON(&itemHideReq)

	if err != nil {
		log.Warnf(c, "[transaction_items.ItemHideHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.items.HideItem(c, uid, []int64{itemHideReq.Id}, itemHideReq.Hidden)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemHideHandler] failed to hide item \"id:%d\" for user \"uid:%d\", because %s", itemHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_items.ItemHideHandler] user \"uid:%d\" has hidden item \"id:%d\"", uid, itemHideReq.Id)
	return true, nil
}

// ItemMoveHandler moves display order of existed transaction items by request parameters for current user
func (a *TransactionItemsApi) ItemMoveHandler(c *core.WebContext) (any, *errs.Error) {
	var itemMoveReq models.TransactionItemMoveRequest
	err := c.ShouldBindJSON(&itemMoveReq)

	if err != nil {
		log.Warnf(c, "[transaction_items.ItemMoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	items := make([]*models.TransactionItem, len(itemMoveReq.NewDisplayOrders))

	for i := 0; i < len(itemMoveReq.NewDisplayOrders); i++ {
		newDisplayOrder := itemMoveReq.NewDisplayOrders[i]
		item := &models.TransactionItem{
			Uid:          uid,
			ItemId:       newDisplayOrder.Id,
			DisplayOrder: newDisplayOrder.DisplayOrder,
		}

		items[i] = item
	}

	err = a.items.ModifyItemDisplayOrders(c, uid, items)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemMoveHandler] failed to move items for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_items.ItemMoveHandler] user \"uid:%d\" has moved items", uid)
	return true, nil
}

// ItemDeleteHandler deletes an existed transaction item by request parameters for current user
func (a *TransactionItemsApi) ItemDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var itemDeleteReq models.TransactionItemDeleteRequest
	err := c.ShouldBindJSON(&itemDeleteReq)

	if err != nil {
		log.Warnf(c, "[transaction_items.ItemDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.items.DeleteItem(c, uid, itemDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_items.ItemDeleteHandler] failed to delete item \"id:%d\" for user \"uid:%d\", because %s", itemDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_items.ItemDeleteHandler] user \"uid:%d\" has deleted item \"id:%d\"", uid, itemDeleteReq.Id)
	return true, nil
}

func (a *TransactionItemsApi) createNewItemModel(uid int64, itemCreateReq *models.TransactionItemCreateRequest, order int32) *models.TransactionItem {
	return &models.TransactionItem{
		Uid:          uid,
		Name:         itemCreateReq.Name,
		ItemGroupId:  itemCreateReq.GroupId,
		DisplayOrder: order,
	}
}

func (a *TransactionItemsApi) createNewItemModels(uid int64, itemCreateBatchReq *models.TransactionItemCreateBatchRequest, order int32) []*models.TransactionItem {
	items := make([]*models.TransactionItem, len(itemCreateBatchReq.Items))

	for i := 0; i < len(itemCreateBatchReq.Items); i++ {
		itemCreateReq := itemCreateBatchReq.Items[i]
		item := a.createNewItemModel(uid, itemCreateReq, order+int32(i))
		item.ItemGroupId = itemCreateBatchReq.GroupId
		items[i] = item
	}

	return items
}

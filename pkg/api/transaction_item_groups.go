package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// TransactionItemGroupsApi represents transaction item group api
type TransactionItemGroupsApi struct {
	itemGroups *services.TransactionItemGroupService
}

// Initialize a transaction item group api singleton instance
var (
	TransactionItemGroups = &TransactionItemGroupsApi{
		itemGroups: services.TransactionItemGroups,
	}
)

// ItemGroupListHandler returns transaction item group list of current user
func (a *TransactionItemGroupsApi) ItemGroupListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	itemGroups, err := a.itemGroups.GetAllItemGroupsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[transaction_item_groups.ItemGroupListHandler] failed to get item groups for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	itemGroupResps := make(models.TransactionItemGroupInfoResponseSlice, len(itemGroups))

	for i := 0; i < len(itemGroups); i++ {
		itemGroupResps[i] = itemGroups[i].ToTransactionItemGroupInfoResponse()
	}

	sort.Sort(itemGroupResps)

	return itemGroupResps, nil
}

// ItemGroupGetHandler returns one specific transaction item group of current user
func (a *TransactionItemGroupsApi) ItemGroupGetHandler(c *core.WebContext) (any, *errs.Error) {
	var itemGroupGetReq models.TransactionItemGroupGetRequest
	err := c.ShouldBindQuery(&itemGroupGetReq)

	if err != nil {
		log.Warnf(c, "[transaction_item_groups.ItemGroupGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	itemGroup, err := a.itemGroups.GetItemGroupByItemGroupId(c, uid, itemGroupGetReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_item_groups.ItemGroupGetHandler] failed to get item group \"id:%d\" for user \"uid:%d\", because %s", itemGroupGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	itemGroupResp := itemGroup.ToTransactionItemGroupInfoResponse()

	return itemGroupResp, nil
}

// ItemGroupCreateHandler saves a new transaction item group by request parameters for current user
func (a *TransactionItemGroupsApi) ItemGroupCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var itemGroupCreateReq models.TransactionItemGroupCreateRequest
	err := c.ShouldBindJSON(&itemGroupCreateReq)

	if err != nil {
		log.Warnf(c, "[transaction_item_groups.ItemGroupCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	maxOrderId, err := a.itemGroups.GetMaxDisplayOrder(c, uid)

	if err != nil {
		log.Errorf(c, "[transaction_item_groups.ItemGroupCreateHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	itemGroup := a.createNewItemGroupModel(uid, &itemGroupCreateReq, maxOrderId+1)

	err = a.itemGroups.CreateItemGroup(c, itemGroup)

	if err != nil {
		log.Errorf(c, "[transaction_item_groups.ItemGroupCreateHandler] failed to create item group \"id:%d\" for user \"uid:%d\", because %s", itemGroup.ItemGroupId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_item_groups.ItemGroupCreateHandler] user \"uid:%d\" has created a new item group \"id:%d\" successfully", uid, itemGroup.ItemGroupId)

	itemGroupResp := itemGroup.ToTransactionItemGroupInfoResponse()

	return itemGroupResp, nil
}

// ItemGroupModifyHandler saves an existed transaction item group by request parameters for current user
func (a *TransactionItemGroupsApi) ItemGroupModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var itemGroupModifyReq models.TransactionItemGroupModifyRequest
	err := c.ShouldBindJSON(&itemGroupModifyReq)

	if err != nil {
		log.Warnf(c, "[transaction_item_groups.ItemGroupModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	itemGroup, err := a.itemGroups.GetItemGroupByItemGroupId(c, uid, itemGroupModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_item_groups.ItemGroupModifyHandler] failed to get item group \"id:%d\" for user \"uid:%d\", because %s", itemGroupModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	newItemGroup := &models.TransactionItemGroup{
		ItemGroupId: itemGroup.ItemGroupId,
		Uid:         uid,
		Name:        itemGroupModifyReq.Name,
	}

	if newItemGroup.Name == itemGroup.Name {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.itemGroups.ModifyItemGroup(c, newItemGroup)

	if err != nil {
		log.Errorf(c, "[transaction_item_groups.ItemGroupModifyHandler] failed to update item group \"id:%d\" for user \"uid:%d\", because %s", itemGroupModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_item_groups.ItemGroupModifyHandler] user \"uid:%d\" has updated item group \"id:%d\" successfully", uid, itemGroupModifyReq.Id)

	itemGroup.Name = newItemGroup.Name
	itemGroupResp := itemGroup.ToTransactionItemGroupInfoResponse()

	return itemGroupResp, nil
}

// ItemGroupMoveHandler moves display order of existed transaction item groups by request parameters for current user
func (a *TransactionItemGroupsApi) ItemGroupMoveHandler(c *core.WebContext) (any, *errs.Error) {
	var itemGroupMoveReq models.TransactionItemGroupMoveRequest
	err := c.ShouldBindJSON(&itemGroupMoveReq)

	if err != nil {
		log.Warnf(c, "[transaction_item_groups.ItemGroupMoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	itemGroups := make([]*models.TransactionItemGroup, len(itemGroupMoveReq.NewDisplayOrders))

	for i := 0; i < len(itemGroupMoveReq.NewDisplayOrders); i++ {
		newDisplayOrder := itemGroupMoveReq.NewDisplayOrders[i]
		itemGroup := &models.TransactionItemGroup{
			Uid:          uid,
			ItemGroupId:  newDisplayOrder.Id,
			DisplayOrder: newDisplayOrder.DisplayOrder,
		}

		itemGroups[i] = itemGroup
	}

	err = a.itemGroups.ModifyItemGroupDisplayOrders(c, uid, itemGroups)

	if err != nil {
		log.Errorf(c, "[transaction_item_groups.ItemGroupMoveHandler] failed to move item groups for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_item_groups.ItemGroupMoveHandler] user \"uid:%d\" has moved item groups", uid)
	return true, nil
}

// ItemGroupDeleteHandler deletes an existed transaction item group by request parameters for current user
func (a *TransactionItemGroupsApi) ItemGroupDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var itemGroupDeleteReq models.TransactionItemGroupDeleteRequest
	err := c.ShouldBindJSON(&itemGroupDeleteReq)

	if err != nil {
		log.Warnf(c, "[transaction_item_groups.ItemGroupDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.itemGroups.DeleteItemGroup(c, uid, itemGroupDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_item_groups.ItemGroupDeleteHandler] failed to delete item group \"id:%d\" for user \"uid:%d\", because %s", itemGroupDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_item_groups.ItemGroupDeleteHandler] user \"uid:%d\" has deleted item group \"id:%d\"", uid, itemGroupDeleteReq.Id)
	return true, nil
}

func (a *TransactionItemGroupsApi) createNewItemGroupModel(uid int64, itemGroupCreateReq *models.TransactionItemGroupCreateRequest, order int32) *models.TransactionItemGroup {
	return &models.TransactionItemGroup{
		Uid:          uid,
		Name:         itemGroupCreateReq.Name,
		DisplayOrder: order,
	}
}

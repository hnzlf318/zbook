import { ref, computed } from 'vue';
import { defineStore } from 'pinia';

import { type BeforeResolveFunction, itemAndIndex, values } from '@/core/base.ts';

import {
    type TransactionItemGroupInfoResponse,
    type TransactionItemGroupNewDisplayOrderRequest,
    TransactionItemGroup
} from '@/models/transaction_item_group.ts';

import {
    type TransactionItemCreateBatchRequest,
    type TransactionItemInfoResponse,
    type TransactionItemNewDisplayOrderRequest,
    TransactionItem
} from '@/models/transaction_item.ts';

import { isEquals } from '@/lib/common.ts';

import logger from '@/lib/logger.ts';
import services, { type ApiResponsePromise } from '@/lib/services.ts';

export const useTransactionItemsStore = defineStore('transactionItems', () => {
    const allTransactionItemGroups = ref<TransactionItemGroup[]>([]);
    const allTransactionItemGroupsMap = ref<Record<string, TransactionItemGroup>>({});
    const allTransactionItems = ref<TransactionItem[]>([]);
    const allTransactionItemsMap = ref<Record<string, TransactionItem>>({});
    const allTransactionItemsByGroupMap = ref<Record<string, TransactionItem[]>>({});
    const transactionItemGroupListStateInvalid = ref<boolean>(true);
    const transactionItemListStateInvalid = ref<boolean>(true);

    const allAvailableItemsCount = computed<number>(() => allTransactionItems.value.length);

    function loadTransactionItemGroupList(itemGroups: TransactionItemGroup[]): void {
        allTransactionItemGroups.value = itemGroups;
        allTransactionItemGroupsMap.value = {};

        for (const itemGroup of itemGroups) {
            allTransactionItemGroupsMap.value[itemGroup.id] = itemGroup;
        }
    }

    function loadTransactionItemList(items: TransactionItem[]): void {
        allTransactionItems.value = items;
        allTransactionItemsMap.value = {};
        allTransactionItemsByGroupMap.value = {};

        for (const item of items) {
            allTransactionItemsMap.value[item.id] = item;
        }

        for (const item of items) {
            let itemsInGroup = allTransactionItemsByGroupMap.value[item.groupId];

            if (!itemsInGroup) {
                itemsInGroup = [];
                allTransactionItemsByGroupMap.value[item.groupId] = itemsInGroup;
            }

            itemsInGroup.push(item);
        }
    }

    function addItemGroupToTransactionItemGroupList(itemGroup: TransactionItemGroup): void {
        allTransactionItemGroups.value.push(itemGroup);
        allTransactionItemGroupsMap.value[itemGroup.id] = itemGroup;
    }

    function addItemToTransactionItemList(item: TransactionItem): void {
        allTransactionItems.value.push(item);
        allTransactionItemsMap.value[item.id] = item;

        let itemsInGroup = allTransactionItemsByGroupMap.value[item.groupId];

        if (!itemsInGroup) {
            itemsInGroup = [];
            allTransactionItemsByGroupMap.value[item.groupId] = itemsInGroup;
        }

        itemsInGroup.push(item);
    }

    function updateItemGroupInTransactionItemGroupList(currentItemGroup: TransactionItemGroup): void {
        for (const [transactionItemGroup, index] of itemAndIndex(allTransactionItemGroups.value)) {
            if (transactionItemGroup.id === currentItemGroup.id) {
                allTransactionItemGroups.value.splice(index, 1, currentItemGroup);
                break;
            }
        }

        allTransactionItemGroupsMap.value[currentItemGroup.id] = currentItemGroup;
    }

    function updateItemInTransactionItemList(currentItem: TransactionItem, oldItemGroupId?: string): void {
        for (const [transactionItem, index] of itemAndIndex(allTransactionItems.value)) {
            if (transactionItem.id === currentItem.id) {
                if (oldItemGroupId && oldItemGroupId !== currentItem.groupId) {
                    allTransactionItems.value.splice(index, 1);
                } else {
                    allTransactionItems.value.splice(index, 1, currentItem);
                }
                break;
            }
        }

        if (oldItemGroupId && oldItemGroupId !== currentItem.groupId) {
            let insertIndex = allTransactionItems.value.length;

            for (const [item, index] of itemAndIndex(allTransactionItems.value)) {
                if (item.groupId === currentItem.groupId) {
                    insertIndex = index;
                    break;
                }
            }

            allTransactionItems.value.splice(insertIndex, 0, currentItem);
        }

        allTransactionItemsMap.value[currentItem.id] = currentItem;

        for (const items of values(allTransactionItemsByGroupMap.value)) {
            for (const [transactionItem, index] of itemAndIndex(items)) {
                if (transactionItem.id === currentItem.id) {
                    if (oldItemGroupId && oldItemGroupId !== currentItem.groupId) {
                        items.splice(index, 1);
                    } else {
                        items.splice(index, 1, currentItem);
                    }
                    break;
                }
            }
        }

        if (oldItemGroupId && oldItemGroupId !== currentItem.groupId) {
            let newGroupItems = allTransactionItemsByGroupMap.value[currentItem.groupId];

            if (!newGroupItems) {
                newGroupItems = [];
                allTransactionItemsByGroupMap.value[currentItem.groupId] = newGroupItems;
            }

            newGroupItems.push(currentItem);
        }
    }

    function updateItemGroupDisplayOrderInTransactionItemList({ from, to }: { from: number, to: number }): void {
        allTransactionItemGroups.value.splice(to, 0, allTransactionItemGroups.value.splice(from, 1)[0] as TransactionItemGroup);
    }

    function updateItemDisplayOrderInTransactionItemList({ groupId, from, to }: { groupId: string, from: number, to: number }): void {
        const itemsInGroup = allTransactionItemsByGroupMap.value[groupId];

        if (!itemsInGroup) {
            return;
        }

        const fromItem = itemsInGroup[from];
        if (!fromItem) return;
        const toItem = itemsInGroup[to];
        if (!toItem) return;

        itemsInGroup.splice(to, 0, itemsInGroup.splice(from, 1)[0] as TransactionItem);

        let mainListFromIndex = -1;
        let mainListToIndex = -1;
        for (const [item, index] of itemAndIndex(allTransactionItems.value)) {
            if (item.id === fromItem.id) mainListFromIndex = index;
            if (item.id === toItem.id) mainListToIndex = index;
        }
        if (mainListFromIndex === -1 || mainListToIndex === -1) return;
        allTransactionItems.value.splice(mainListToIndex, 0, allTransactionItems.value.splice(mainListFromIndex, 1)[0] as TransactionItem);
    }

    function updateItemVisibilityInTransactionItemList({ item, hidden }: { item: TransactionItem, hidden: boolean }): void {
        if (allTransactionItemsMap.value[item.id]) {
            allTransactionItemsMap.value[item.id]!.hidden = hidden;
        }
    }

    function removeItemGroupFromTransactionItemGroupList(currentItemGroup: TransactionItemGroup): void {
        for (const [transactionItemGroup, index] of itemAndIndex(allTransactionItemGroups.value)) {
            if (transactionItemGroup.id === currentItemGroup.id) {
                allTransactionItemGroups.value.splice(index, 1);
                break;
            }
        }
        if (allTransactionItemGroupsMap.value[currentItemGroup.id]) {
            delete allTransactionItemGroupsMap.value[currentItemGroup.id];
        }
    }

    function removeItemFromTransactionItemList(currentItem: TransactionItem): void {
        for (const [transactionItem, index] of itemAndIndex(allTransactionItems.value)) {
            if (transactionItem.id === currentItem.id) {
                allTransactionItems.value.splice(index, 1);
                break;
            }
        }
        if (allTransactionItemsMap.value[currentItem.id]) {
            delete allTransactionItemsMap.value[currentItem.id];
        }
        for (const items of values(allTransactionItemsByGroupMap.value)) {
            for (const [transactionItem, index] of itemAndIndex(items)) {
                if (transactionItem.id === currentItem.id) {
                    items.splice(index, 1);
                    break;
                }
            }
        }
    }

    function updateTransactionItemGroupListInvalidState(invalidState: boolean): void {
        transactionItemGroupListStateInvalid.value = invalidState;
    }

    function updateTransactionItemListInvalidState(invalidState: boolean): void {
        transactionItemListStateInvalid.value = invalidState;
    }

    function resetTransactionItems(): void {
        allTransactionItemGroups.value = [];
        allTransactionItemGroupsMap.value = {};
        allTransactionItems.value = [];
        allTransactionItemsMap.value = {};
        allTransactionItemsByGroupMap.value = {};
        transactionItemGroupListStateInvalid.value = true;
        transactionItemListStateInvalid.value = true;
    }

    function loadAllItemGroups({ force }: { force?: boolean }): Promise<TransactionItemGroup[]> {
        if (!force && !transactionItemGroupListStateInvalid.value) {
            return new Promise((resolve) => resolve(allTransactionItemGroups.value));
        }
        return new Promise((resolve, reject) => {
            services.getAllTransactionItemGroups().then(response => {
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve item group list' });
                    return;
                }
                if (transactionItemGroupListStateInvalid.value) {
                    updateTransactionItemGroupListInvalidState(false);
                }
                const transactionItemGroups = TransactionItemGroup.ofMulti(data.result);
                loadTransactionItemGroupList(transactionItemGroups);
                resolve(transactionItemGroups);
            }).catch(error => {
                logger.error('failed to load item group list', error);
                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve item group list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function loadAllItems({ force }: { force?: boolean }): Promise<TransactionItem[]> {
        if (!force && !transactionItemGroupListStateInvalid.value && !transactionItemListStateInvalid.value) {
            return new Promise((resolve) => resolve(allTransactionItems.value));
        }
        return new Promise((resolve, reject) => {
            services.getAllTransactionItemGroups().then(response => {
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve item list' });
                    return;
                }
                if (transactionItemGroupListStateInvalid.value) {
                    updateTransactionItemGroupListInvalidState(false);
                }
                const transactionItemGroups = TransactionItemGroup.ofMulti(data.result);
                loadTransactionItemGroupList(transactionItemGroups);
                return services.getAllTransactionItems();
            }).then(response => {
                if (!response) {
                    reject({ message: 'Unable to retrieve item list' });
                    return;
                }
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve item list' });
                    return;
                }
                if (transactionItemListStateInvalid.value) {
                    updateTransactionItemListInvalidState(false);
                }
                const transactionItems = TransactionItem.ofMulti(data.result);
                if (force && data.result && isEquals(allTransactionItems.value, transactionItems)) {
                    reject({ message: 'Item list is up to date', isUpToDate: true });
                    return;
                }
                loadTransactionItemList(transactionItems);
                resolve(transactionItems);
            }).catch(error => {
                if (force) {
                    logger.error('failed to force load item list', error);
                } else {
                    logger.error('failed to load item list', error);
                }
                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve item list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveItemGroup({ itemGroup }: { itemGroup: TransactionItemGroup }): Promise<TransactionItemGroup> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<TransactionItemGroupInfoResponse>;
            if (!itemGroup.id) {
                promise = services.addTransactionItemGroup(itemGroup.toCreateRequest());
            } else {
                promise = services.modifyTransactionItemGroup(itemGroup.toModifyRequest());
            }
            promise.then(response => {
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    if (!itemGroup.id) {
                        reject({ message: 'Unable to add item group' });
                    } else {
                        reject({ message: 'Unable to save item group' });
                    }
                    return;
                }
                const transactionItemGroup = TransactionItemGroup.of(data.result);
                if (!itemGroup.id) {
                    addItemGroupToTransactionItemGroupList(transactionItemGroup);
                } else {
                    updateItemGroupInTransactionItemGroupList(transactionItemGroup);
                }
                resolve(transactionItemGroup);
            }).catch(error => {
                logger.error('failed to save item group', error);
                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!itemGroup.id) {
                        reject({ message: 'Unable to add item group' });
                    } else {
                        reject({ message: 'Unable to save item group' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function changeItemGroupDisplayOrder({ itemGroupId, from, to }: { itemGroupId: string, from: number, to: number }): Promise<void> {
        return new Promise((resolve, reject) => {
            let currentItemGroup: TransactionItemGroup | null = null;
            for (const g of allTransactionItemGroups.value) {
                if (g.id === itemGroupId) {
                    currentItemGroup = g;
                    break;
                }
            }
            if (!currentItemGroup || !allTransactionItemGroups.value[to]) {
                reject({ message: 'Unable to move item group' });
                return;
            }
            if (!transactionItemGroupListStateInvalid.value) {
                updateTransactionItemGroupListInvalidState(true);
            }
            updateItemGroupDisplayOrderInTransactionItemList({ from, to });
            resolve();
        });
    }

    function updateItemGroupDisplayOrders(): Promise<boolean> {
        const newDisplayOrders: TransactionItemGroupNewDisplayOrderRequest[] = [];
        for (const [g, index] of itemAndIndex(allTransactionItemGroups.value)) {
            newDisplayOrders.push({ id: g.id, displayOrder: index + 1 });
        }
        return new Promise((resolve, reject) => {
            services.moveTransactionItemGroup({ newDisplayOrders }).then(response => {
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move item group' });
                    return;
                }
                if (transactionItemGroupListStateInvalid.value) {
                    updateTransactionItemGroupListInvalidState(false);
                }
                resolve(data.result);
            }).catch(error => {
                logger.error('failed to save item groups display order', error);
                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move item group' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteItemGroup({ itemGroup, beforeResolve }: { itemGroup: TransactionItemGroup, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteTransactionItemGroup({ id: itemGroup.id }).then(response => {
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this item group' });
                    return;
                }
                if (beforeResolve) {
                    beforeResolve(() => removeItemGroupFromTransactionItemGroupList(itemGroup));
                } else {
                    removeItemGroupFromTransactionItemGroupList(itemGroup);
                }
                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete item group', error);
                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this item group' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveItem({ item, beforeResolve }: { item: TransactionItem, beforeResolve?: BeforeResolveFunction }): Promise<TransactionItem> {
        const oldItemGroupId = allTransactionItemsMap.value[item.id]?.groupId;
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<TransactionItemInfoResponse>;
            if (!item.id) {
                promise = services.addTransactionItem(item.toCreateRequest());
            } else {
                promise = services.modifyTransactionItem(item.toModifyRequest());
            }
            promise.then(response => {
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    if (!item.id) {
                        reject({ message: 'Unable to add item' });
                    } else {
                        reject({ message: 'Unable to save item' });
                    }
                    return;
                }
                const transactionItem = TransactionItem.of(data.result);
                if (beforeResolve) {
                    beforeResolve(() => {
                        if (!item.id) {
                            addItemToTransactionItemList(transactionItem);
                        } else {
                            updateItemInTransactionItemList(transactionItem, oldItemGroupId);
                        }
                    });
                } else {
                    if (!item.id) {
                        addItemToTransactionItemList(transactionItem);
                    } else {
                        updateItemInTransactionItemList(transactionItem, oldItemGroupId);
                    }
                }
                resolve(transactionItem);
            }).catch(error => {
                logger.error('failed to save item', error);
                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!item.id) {
                        reject({ message: 'Unable to add item' });
                    } else {
                        reject({ message: 'Unable to save item' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function addItems(req: TransactionItemCreateBatchRequest): Promise<TransactionItem[]> {
        return new Promise((resolve, reject) => {
            services.addTransactionItemBatch(req).then(response => {
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to add item' });
                    return;
                }
                if (!transactionItemListStateInvalid.value) {
                    updateTransactionItemListInvalidState(true);
                }
                resolve(TransactionItem.ofMulti(data.result));
            }).catch(error => {
                logger.error('failed to add items', error);
                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to add item' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function changeItemDisplayOrder({ itemId, from, to }: { itemId: string, from: number, to: number }): Promise<void> {
        return new Promise((resolve, reject) => {
            const currentItem = allTransactionItemsMap.value[itemId];
            if (!currentItem || !allTransactionItemsByGroupMap.value[currentItem.groupId]?.[to]) {
                reject({ message: 'Unable to move item' });
                return;
            }
            if (!transactionItemListStateInvalid.value) {
                updateTransactionItemListInvalidState(true);
            }
            updateItemDisplayOrderInTransactionItemList({ groupId: currentItem.groupId, from, to });
            resolve();
        });
    }

    function updateItemDisplayOrders(groupId: string): Promise<boolean> {
        const itemsInGroup = allTransactionItemsByGroupMap.value[groupId];
        if (!itemsInGroup) {
            return Promise.reject('Unable to move item');
        }
        const newDisplayOrders: TransactionItemNewDisplayOrderRequest[] = [];
        for (const [item, index] of itemAndIndex(itemsInGroup)) {
            newDisplayOrders.push({ id: item.id, displayOrder: index + 1 });
        }
        return new Promise((resolve, reject) => {
            services.moveTransactionItem({ newDisplayOrders }).then(response => {
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move item' });
                    return;
                }
                if (transactionItemListStateInvalid.value) {
                    updateTransactionItemListInvalidState(false);
                }
                resolve(data.result);
            }).catch(error => {
                logger.error('failed to save items display order', error);
                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move item' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function hideItem({ item, hidden }: { item: TransactionItem, hidden: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.hideTransactionItem({ id: item.id, hidden }).then(response => {
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    reject({ message: hidden ? 'Unable to hide this item' : 'Unable to unhide this item' });
                    return;
                }
                updateItemVisibilityInTransactionItemList({ item, hidden });
                resolve(data.result);
            }).catch(error => {
                logger.error('failed to change item visibility', error);
                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: hidden ? 'Unable to hide this item' : 'Unable to unhide this item' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteItem({ item, beforeResolve }: { item: TransactionItem, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteTransactionItem({ id: item.id }).then(response => {
                const data = response.data;
                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this item' });
                    return;
                }
                if (beforeResolve) {
                    beforeResolve(() => removeItemFromTransactionItemList(item));
                } else {
                    removeItemFromTransactionItemList(item);
                }
                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete item', error);
                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this item' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return {
        allTransactionItemGroups,
        allTransactionItemGroupsMap,
        allTransactionItemsMap,
        allTransactionItemsByGroupMap,
        transactionItemGroupListStateInvalid,
        transactionItemListStateInvalid,
        allAvailableItemsCount,
        updateTransactionItemListInvalidState,
        resetTransactionItems,
        loadAllItemGroups,
        loadAllItems,
        saveItemGroup,
        changeItemGroupDisplayOrder,
        updateItemGroupDisplayOrders,
        deleteItemGroup,
        saveItem,
        addItems,
        changeItemDisplayOrder,
        updateItemDisplayOrders,
        hideItem,
        deleteItem
    };
});

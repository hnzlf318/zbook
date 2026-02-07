<template>
    <f7-page :ptr="!sortable && !hasEditingItem" @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')" v-if="!sortable"></f7-nav-left>
            <f7-nav-left v-else-if="sortable">
                <f7-link icon-f7="xmark" :class="{ 'disabled': displayOrderSaving }" @click="cancelSort"></f7-link>
            </f7-nav-left>
            <f7-nav-title>
                <f7-link popover-open=".item-group-popover-menu" :class="{ 'disabled': loading || sortable || displayOrderModified || hasEditingItem }">
                    <span style="color: var(--f7-text-color)">{{ displayItemGroupName }}</span>
                    <f7-icon class="page-title-bar-icon" color="gray" style="opacity: 0.5" f7="chevron_down_circle_fill"></f7-icon>
                </f7-link>
            </f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading }">
                <f7-link icon-f7="ellipsis" :class="{ 'disabled': hasEditingItem || sortable }" @click="showMoreActionSheet = true"></f7-link>
                <f7-link icon-f7="plus" :class="{ 'disabled': hasEditingItem }" v-if="!sortable" @click="add"></f7-link>
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': displayOrderSaving || !displayOrderModified || hasEditingItem }" @click="saveSortResult" v-else-if="sortable"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-popover class="item-group-popover-menu"
                    @popover:open="scrollPopoverToSelectedItem">
            <f7-list dividers>
                <f7-list-item link="#" no-chevron popover-close
                              :title="itemGroup.name"
                              :class="{ 'list-item-selected': activeItemGroupId === itemGroup.id }"
                              :key="itemGroup.id"
                              v-for="itemGroup in allItemGroupsWithDefault"
                              @click="switchItemGroup(itemGroup.id)">
                    <template #after>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" v-if="activeItemGroupId === itemGroup.id"></f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <f7-list strong inset dividers class="item-item-list margin-top skeleton-text" v-if="loading">
            <f7-list-item :key="itemIdx" v-for="itemIdx in [ 1, 2, 3 ]">
                <template #media>
                    <f7-icon class="transaction-item-icon" f7="list_bullet"></f7-icon>
                </template>
                <template #title>
                    <div class="display-flex">
                        <div class="transaction-item-list-item-content list-item-valign-middle padding-inline-start-half">Item Name</div>
                    </div>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="item-item-list margin-top" v-if="!loading && noAvailableItem && !newItem">
            <f7-list-item :title="tt('No available item')"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers sortable class="item-item-list margin-top"
                 :sortable-enabled="sortable" @sortable:sort="onSort"
                 v-if="!loading">
            <f7-list-item swipeout
                          :class="{ 'actual-first-child': item.id === firstShowingId, 'actual-last-child': item.id === lastShowingId && !newItem, 'editing-list-item': editingItem.id === item.id }"
                          :id="getItemDomId(item)"
                          :key="item.id"
                          v-for="item in items"
                          v-show="showHidden || !item.hidden"
                          @taphold="setSortable()">
                <template #media>
                    <f7-icon class="transaction-item-icon" f7="list_bullet">
                        <f7-badge color="gray" class="right-bottom-icon" v-if="item.hidden">
                            <f7-icon f7="eye_slash_fill"></f7-icon>
                        </f7-badge>
                    </f7-icon>
                </template>
                <template #title>
                    <div class="display-flex">
                        <div class="transaction-item-list-item-content list-item-valign-middle padding-inline-start-half"
                             v-if="editingItem.id !== item.id">
                            {{ item.name }}
                        </div>
                        <f7-input class="list-title-input padding-inline-start-half"
                                  type="text"
                                  :placeholder="tt('Item Title')"
                                  v-else-if="editingItem.id === item.id"
                                  v-model:value="editingItem.name"
                                  @keyup.enter="save(editingItem)">
                        </f7-input>
                    </div>
                </template>
                <template #after>
                    <f7-button :class="{ 'no-padding': true, 'disabled': !isItemModified(item) }"
                               raised fill
                               icon-f7="checkmark_alt"
                               color="blue"
                               v-if="editingItem.id === item.id"
                               @click="save(editingItem)">
                    </f7-button>
                    <f7-button class="no-padding margin-inline-start-half"
                               raised fill
                               icon-f7="xmark"
                               color="gray"
                               v-if="editingItem.id === item.id"
                               @click="cancelSave(editingItem)">
                    </f7-button>
                </template>
                <f7-swipeout-actions :left="textDirection === TextDirection.LTR"
                                     :right="textDirection === TextDirection.RTL"
                                     v-if="sortable && editingItem.id !== item.id">
                    <f7-swipeout-button :color="item.hidden ? 'blue' : 'gray'" class="padding-horizontal"
                                        overswipe close @click="hide(item, !item.hidden)">
                        <f7-icon :f7="item.hidden ? 'eye' : 'eye_slash'"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
                <f7-swipeout-actions :left="textDirection === TextDirection.RTL"
                                     :right="textDirection === TextDirection.LTR"
                                     v-if="!sortable && editingItem.id !== item.id">
                    <f7-swipeout-button color="primary" close :class="{ 'disabled': allItemGroupsWithDefault.length < 2 }" :text="tt('Move')" @click="moveItemToGroup(item)"></f7-swipeout-button>
                    <f7-swipeout-button color="orange" close :text="tt('Edit')" @click="edit(item)"></f7-swipeout-button>
                    <f7-swipeout-button color="red" class="padding-horizontal" @click="remove(item, false)">
                        <f7-icon f7="trash"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
            </f7-list-item>

            <f7-list-item class="editing-list-item" v-if="newItem">
                <template #media>
                    <f7-icon class="transaction-item-icon" f7="list_bullet"></f7-icon>
                </template>
                <template #title>
                    <div class="display-flex">
                        <f7-input class="list-title-input padding-inline-start-half"
                                  type="text"
                                  :placeholder="tt('Item Title')"
                                  v-model:value="newItem.name"
                                  @keyup.enter="save(newItem)">
                        </f7-input>
                    </div>
                </template>
                <template #after>
                    <f7-button :class="{ 'no-padding': true, 'disabled': !isItemModified(newItem) }"
                               raised fill
                               icon-f7="checkmark_alt"
                               color="blue"
                               @click="save(newItem)">
                    </f7-button>
                    <f7-button class="no-padding margin-inline-start-half"
                               raised fill
                               icon-f7="xmark"
                               color="gray"
                               @click="cancelSave(newItem)">
                    </f7-button>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-popup push :close-on-escape="false" :opened="showMoveItemPopup"
                  @popup:closed="showMoveItemPopup = false">
            <f7-page>
                <f7-navbar>
                    <f7-nav-left>
                        <f7-link popup-close icon-f7="xmark"></f7-link>
                    </f7-nav-left>
                    <f7-nav-title :title="tt('Move to...')"></f7-nav-title>
                    <f7-nav-right>
                        <f7-link icon-f7="checkmark_alt"
                                 :class="{ 'disabled': !itemToMove || !moveToItemGroupId }"
                                 @click="moveItemToGroup(itemToMove, moveToItemGroupId)"></f7-link>
                    </f7-nav-right>
                </f7-navbar>

                <f7-list strong inset dividers class="margin-top" v-if="!loading && allItemGroupsWithDefault.length < 2">
                    <f7-list-item :title="tt('No available item group')"></f7-list-item>
                </f7-list>

                <f7-list strong inset dividers class="margin-vertical" v-if="allItemGroupsWithDefault.length >= 2">
                    <template :key="itemGroup.id" v-for="itemGroup in allItemGroupsWithDefault">
                        <f7-list-item checkbox
                                      :title="itemGroup.name"
                                      :value="itemGroup.id"
                                      :checked="moveToItemGroupId === itemGroup.id"
                                      :key="itemGroup.id"
                                      @change="updateItemGroupSelected"
                                      v-if="itemToMove?.groupId !== itemGroup.id"></f7-list-item>
                    </template>
                </f7-list>
            </f7-page>
        </f7-popup>

        <f7-actions close-by-outside-click close-on-escape :opened="showMoreActionSheet" @actions:closed="showMoreActionSheet = false">
            <f7-actions-group>
                <f7-actions-button @click="addItemGroup">{{ tt('Add Item Group') }}</f7-actions-button>
                <f7-actions-button @click="renameItemGroup"
                                   v-if="activeItemGroupId && activeItemGroupId !== DEFAULT_ITEM_GROUP_ID">{{ tt('Rename Item Group') }}</f7-actions-button>
                <f7-actions-button color="red" :class="{ 'disabled': items && items.length > 0 }"
                                   @click="removeItemGroup"
                                   v-if="activeItemGroupId && activeItemGroupId !== DEFAULT_ITEM_GROUP_ID">{{ tt('Delete Item Group') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="allItemGroupsWithDefault.length >= 2">
                <f7-actions-button @click="changeItemGroupDisplayOrder">{{ tt('Change Group Display Order') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button @click="setSortable()">{{ tt('Sort') }}</f7-actions-button>
                <f7-actions-button v-if="!showHidden" @click="showHidden = true">{{ tt('Show Hidden Transaction Items') }}</f7-actions-button>
                <f7-actions-button v-if="showHidden" @click="showHidden = false">{{ tt('Hide Hidden Transaction Items') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteActionSheet" @actions:closed="showDeleteActionSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this item?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="remove(itemToDelete, true)">{{ tt('Delete') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { type Framework7Dom, useI18nUIComponents, showLoading, hideLoading, onSwipeoutDeleted } from '@/lib/ui/mobile.ts';
import { useItemListPageBase } from '@/views/base/items/ItemListPageBase.ts';

import { useTransactionItemsStore } from '@/stores/transactionItem.ts';

import { TextDirection } from '@/core/text.ts';
import { DEFAULT_ITEM_GROUP_ID } from '@/consts/item.ts';

import { TransactionItemGroup } from '@/models/transaction_item_group.ts';
import { TransactionItem } from '@/models/transaction_item.ts';

import { scrollToSelectedItem } from '@/lib/ui/common.ts';
import { getFirstShowingId, getLastShowingId } from '@/lib/item.ts';

const props = defineProps<{
    f7router: Router.Router;
}>();

const { tt, getCurrentLanguageTextDirection } = useI18n();
const { showAlert, showConfirm, showPrompt, showToast, routeBackOnError } = useI18nUIComponents();

const {
    activeItemGroupId,
    newItem,
    editingItem,
    loading,
    showHidden,
    displayOrderModified,
    allItemGroupsWithDefault,
    items,
    noAvailableItem,
    hasEditingItem,
    isItemModified,
    switchItemGroup,
    add,
    edit
} = useItemListPageBase();

const transactionItemsStore = useTransactionItemsStore();

const loadingError = ref<unknown | null>(null);
const sortable = ref<boolean>(false);
const moveToItemGroupId = ref<string | undefined>(undefined);
const itemToMove = ref<TransactionItem | null>(null);
const itemToDelete = ref<TransactionItem | null>(null);
const showMoveItemPopup = ref<boolean>(false);
const showMoreActionSheet = ref<boolean>(false);
const showDeleteActionSheet = ref<boolean>(false);
const displayOrderSaving = ref<boolean>(false);

const textDirection = computed<TextDirection>(() => getCurrentLanguageTextDirection());
const firstShowingId = computed<string | null>(() => getFirstShowingId(items.value, showHidden.value));
const lastShowingId = computed<string | null>(() => getLastShowingId(items.value, showHidden.value));

const displayItemGroupName = computed<string>(() => {
    const itemGroup = transactionItemsStore.allTransactionItemGroupsMap[activeItemGroupId.value];
    return itemGroup ? itemGroup.name : tt('Default Group');
});

function getItemDomId(item: TransactionItem): string {
    return 'item_' + item.id;
}

function parseItemIdFromDomId(domId: string): string | null {
    if (!domId || domId.indexOf('item_') !== 0) {
        return null;
    }

    return domId.substring(5);
}

function init(): void {
    loading.value = true;

    transactionItemsStore.loadAllItems({
        force: false
    }).then(() => {
        loading.value = false;
    }).catch(error => {
        if (error.processed) {
            loading.value = false;
        } else {
            loadingError.value = error;
            showToast(error.message || error);
        }
    });
}

function reload(done?: () => void): void {
    if (sortable.value || hasEditingItem.value) {
        done?.();
        return;
    }

    const force = !!done;

    transactionItemsStore.loadAllItems({
        force: force
    }).then(() => {
        done?.();

        if (force) {
            showToast('Item list has been updated');
        }
    }).catch(error => {
        done?.();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function save(item: TransactionItem): void {
    showLoading();

    transactionItemsStore.saveItem({
        item: item
    }).then(() => {
        hideLoading();

        if (item.id) {
            editingItem.value.id = '';
            editingItem.value.name = '';
        } else {
            newItem.value = null;
        }
    }).catch(error => {
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function cancelSave(item: TransactionItem): void {
    if (item.id) {
        editingItem.value.id = '';
        editingItem.value.name = '';
    } else {
        newItem.value = null;
    }
}

function hide(item: TransactionItem, hidden: boolean): void {
    showLoading();

    transactionItemsStore.hideItem({
        item: item,
        hidden: hidden
    }).then(() => {
        hideLoading();
    }).catch(error => {
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function moveItemToGroup(item: TransactionItem | null, targetItemGroupId?: string): void {
    if (!item) {
        showAlert('An error occurred');
        return;
    }

    if (!targetItemGroupId) {
        moveToItemGroupId.value = undefined;
        itemToMove.value = item;
        showMoveItemPopup.value = true;
        return;
    }

    showMoveItemPopup.value = false;
    itemToMove.value = null;
    moveToItemGroupId.value = undefined;
    showLoading();

    const newItemObj = item.clone();
    newItemObj.groupId = targetItemGroupId;

    transactionItemsStore.saveItem({
        item: newItemObj,
        beforeResolve: (done) => {
            onSwipeoutDeleted(getItemDomId(item), done);
        }
    }).then(() => {
        hideLoading();
    }).catch(error => {
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function updateItemGroupSelected(e: Event): void {
    const target = e.target as HTMLInputElement;

    if (target.checked) {
        moveToItemGroupId.value = target.value;
    } else {
        moveToItemGroupId.value = undefined;
    }
}

function remove(item: TransactionItem | null, confirm: boolean): void {
    if (!item) {
        showAlert('An error occurred');
        return;
    }

    if (!confirm) {
        itemToDelete.value = item;
        showDeleteActionSheet.value = true;
        return;
    }

    showDeleteActionSheet.value = false;
    itemToDelete.value = null;
    showLoading();

    transactionItemsStore.deleteItem({
        item: item,
        beforeResolve: (done) => {
            onSwipeoutDeleted(getItemDomId(item), done);
        }
    }).then(() => {
        hideLoading();
    }).catch(error => {
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function setSortable(): void {
    if (sortable.value || hasEditingItem.value) {
        return;
    }

    showHidden.value = true;
    sortable.value = true;
    displayOrderModified.value = false;
}

function saveSortResult(): void {
    if (!displayOrderModified.value) {
        showHidden.value = false;
        sortable.value = false;
        return;
    }

    displayOrderSaving.value = true;
    showLoading();

    transactionItemsStore.updateItemDisplayOrders(activeItemGroupId.value).then(() => {
        displayOrderSaving.value = false;
        hideLoading();

        showHidden.value = false;
        sortable.value = false;
        displayOrderModified.value = false;
    }).catch(error => {
        displayOrderSaving.value = false;
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function cancelSort(): void {
    if (!displayOrderModified.value) {
        showHidden.value = false;
        sortable.value = false;
        return;
    }

    displayOrderSaving.value = true;
    showLoading();

    transactionItemsStore.loadAllItems({
        force: false
    }).then(() => {
        displayOrderSaving.value = false;
        hideLoading();

        showHidden.value = false;
        sortable.value = false;
        displayOrderModified.value = false;
    }).catch(error => {
        displayOrderSaving.value = false;
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function addItemGroup(): void {
    showPrompt(tt('New Item Group Name'), '', (value: string) => {
        showLoading();

        transactionItemsStore.saveItemGroup({
            itemGroup: TransactionItemGroup.createNewItemGroup(value)
        }).then(itemGroup => {
            hideLoading();
            activeItemGroupId.value = itemGroup.id;
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    });
}

function renameItemGroup(): void {
    const itemGroup = transactionItemsStore.allTransactionItemGroupsMap[activeItemGroupId.value];

    if (!itemGroup) {
        showToast('Unable to rename this item group');
        return;
    }

    showPrompt(tt('Rename Item Group'), itemGroup.name || '', (value: string) => {
        showLoading();

        const newItemGroup = itemGroup.clone();
        newItemGroup.name = value;

        transactionItemsStore.saveItemGroup({
            itemGroup: newItemGroup
        }).then(() => {
            hideLoading();
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    });
}

function removeItemGroup(): void {
    const itemGroup = transactionItemsStore.allTransactionItemGroupsMap[activeItemGroupId.value];

    if (!itemGroup) {
        showToast('Unable to delete this item group');
        return;
    }

    const currentItemGroupIndex = allItemGroupsWithDefault.value.findIndex(group => group.id === itemGroup.id);

    showConfirm('Are you sure you want to delete this item group?', () => {
        showLoading();

        transactionItemsStore.deleteItemGroup({
            itemGroup: itemGroup
        }).then(() => {
            hideLoading();

            if (allItemGroupsWithDefault.value[currentItemGroupIndex]) {
                const newActiveItemGroup = allItemGroupsWithDefault.value[currentItemGroupIndex];
                activeItemGroupId.value = newActiveItemGroup ? newActiveItemGroup.id : DEFAULT_ITEM_GROUP_ID;
            } else if (allItemGroupsWithDefault.value[currentItemGroupIndex - 1]) {
                const newActiveItemGroup = allItemGroupsWithDefault.value[currentItemGroupIndex - 1];
                activeItemGroupId.value = newActiveItemGroup ? newActiveItemGroup.id : DEFAULT_ITEM_GROUP_ID;
            } else {
                activeItemGroupId.value = DEFAULT_ITEM_GROUP_ID;
            }
        }).catch(error => {
            hideLoading();

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    });
}

function changeItemGroupDisplayOrder(): void {
    props.f7router.navigate('/item/group/list');
}

function onSort(event: { el: { id: string }, from: number, to: number }): void {
    if (!event || !event.el || !event.el.id) {
        showToast('Unable to move item');
        return;
    }

    const id = parseItemIdFromDomId(event.el.id);

    if (!id) {
        showToast('Unable to move item');
        return;
    }

    transactionItemsStore.changeItemDisplayOrder({
        itemId: id,
        from: event.from,
        to: event.to
    }).then(() => {
        displayOrderModified.value = true;
    }).catch(error => {
        showToast(error.message || error);
    });
}

function scrollPopoverToSelectedItem(event: { $el: Framework7Dom }): void {
    scrollToSelectedItem(event.$el[0], '.popover-inner', '.popover-inner', 'li.list-item-selected');
}

function onPageAfterIn(): void {
    if (transactionItemsStore.transactionItemListStateInvalid && !loading.value) {
        reload();
    }

    routeBackOnError(props.f7router, loadingError);
}

init();
</script>

<style>
.item-item-list.list .item-media + .item-inner {
    margin-inline-start: 5px;
}

.transaction-item-list-item-content {
    overflow: hidden;
    text-overflow: ellipsis;
}

.item-group-popover-menu .popover-inner {
    max-height: 440px;
    overflow-y: auto;
}
</style>

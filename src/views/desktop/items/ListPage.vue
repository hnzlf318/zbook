<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <v-layout>
                    <v-navigation-drawer :permanent="alwaysShowNav" v-model="showNav">
                        <div class="mx-6 my-4">
                            <span class="text-subtitle-2">{{ tt('Total items') }}</span>
                            <p class="transaction-items-statistic-item-value mt-1">
                                <span v-if="!loading || totalAvailableItemsCount > 0">{{ displayTotalAvailableItemsCount }}</span>
                                <span v-else-if="loading && totalAvailableItemsCount <= 0">
                                    <v-skeleton-loader class="skeleton-no-margin pt-2 pb-1" type="text" :loading="true"></v-skeleton-loader>
                                </span>
                            </p>
                        </div>
                        <v-divider />
                        <v-tabs show-arrows
                                class="scrollable-vertical-tabs"
                                style="max-height: calc(100% - 88px)"
                                direction="vertical"
                                :prev-icon="mdiMenuUp" :next-icon="mdiMenuDown"
                                :disabled="loading || updating" v-model="activeItemGroupId">
                            <v-tab class="tab-text-truncate" :disabled="loading || updating || displayOrderModified || hasEditingItem"
                                   :key="itemGroup.id" :value="itemGroup.id"
                                   v-for="itemGroup in allItemGroupsWithDefault"
                                   @click="switchItemGroup(itemGroup.id)">
                                <span class="text-truncate">{{ itemGroup.name }}</span>
                            </v-tab>
                            <template v-if="loading && (!allItemGroupsWithDefault || allItemGroupsWithDefault.length < 2)">
                                <v-skeleton-loader class="skeleton-no-margin mx-5 mt-4 mb-3" type="text"
                                                   :key="itemIdx" :loading="true" v-for="itemIdx in [ 1, 2, 3, 4, 5 ]"></v-skeleton-loader>
                            </template>
                        </v-tabs>
                    </v-navigation-drawer>
                    <v-main>
                        <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container" v-model="activeTab">
                            <v-window-item value="itemListPage">
                                <v-card variant="flat" min-height="780">
                                    <template #title>
                                        <div class="title-and-toolbar d-flex align-center">
                                            <v-btn class="me-3 d-md-none" density="compact" color="default" variant="plain"
                                                   :ripple="false" :icon="true" @click="showNav = !showNav">
                                                <v-icon :icon="mdiMenu" size="24" />
                                            </v-btn>
                                            <span>{{ tt('Transaction Items') }}</span>
                                            <v-btn class="ms-3" color="default" variant="outlined"
                                                   :disabled="loading || updating || hasEditingItem" @click="add">{{ tt('Add') }}</v-btn>
                                            <v-btn class="ms-3" color="primary" variant="tonal"
                                                   :disabled="loading || updating || hasEditingItem" @click="saveSortResult"
                                                   v-if="displayOrderModified">{{ tt('Save Display Order') }}</v-btn>
                                            <v-btn density="compact" color="default" variant="text" size="24"
                                                   class="ms-2" :icon="true" :disabled="loading || updating || hasEditingItem"
                                                   :loading="loading" @click="reload">
                                                <template #loader>
                                                    <v-progress-circular indeterminate size="20"/>
                                                </template>
                                                <v-icon :icon="mdiRefresh" size="24" />
                                                <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                                            </v-btn>
                                            <v-spacer/>
                                            <v-btn density="comfortable" color="default" variant="text" class="ms-2"
                                                   :disabled="loading || updating || hasEditingItem" :icon="true">
                                                <v-icon :icon="mdiDotsVertical" />
                                                <v-menu activator="parent">
                                                    <v-list>
                                                        <v-list-item :prepend-icon="mdiPlus" @click="addItemGroup">
                                                            <v-list-item-title>{{ tt('Add Item Group') }}</v-list-item-title>
                                                        </v-list-item>
                                                        <v-list-item :prepend-icon="mdiPencilOutline"
                                                                     @click="renameItemGroup"
                                                                     v-if="activeItemGroupId && activeItemGroupId !== DEFAULT_ITEM_GROUP_ID">
                                                            <v-list-item-title>{{ tt('Rename Item Group') }}</v-list-item-title>
                                                        </v-list-item>
                                                        <v-list-item :prepend-icon="mdiDeleteOutline"
                                                                     :disabled="items && items.length > 0"
                                                                     @click="removeItemGroup"
                                                                     v-if="activeItemGroupId && activeItemGroupId !== DEFAULT_ITEM_GROUP_ID">
                                                            <v-list-item-title>{{ tt('Delete Item Group') }}</v-list-item-title>
                                                        </v-list-item>
                                                        <v-divider class="my-2" v-if="allItemGroupsWithDefault.length >= 2"/>
                                                        <v-list-item :prepend-icon="mdiSort"
                                                                     :disabled="!allItemGroupsWithDefault || allItemGroupsWithDefault.length < 2"
                                                                     :title="tt('Change Group Display Order')"
                                                                     v-if="allItemGroupsWithDefault.length >= 2"
                                                                     @click="showChangeGroupDisplayOrderDialog"></v-list-item>
                                                        <v-divider class="my-2"/>
                                                        <v-list-item :prepend-icon="mdiEyeOutline"
                                                                     :title="tt('Show Hidden Transaction Items')"
                                                                     v-if="!showHidden" @click="showHidden = true"></v-list-item>
                                                        <v-list-item :prepend-icon="mdiEyeOffOutline"
                                                                     :title="tt('Hide Hidden Transaction Items')"
                                                                     v-if="showHidden" @click="showHidden = false"></v-list-item>
                                                    </v-list>
                                                </v-menu>
                                            </v-btn>
                                        </div>
                                    </template>

                                    <v-table class="transaction-items-table table-striped" :hover="!loading">
                                        <thead>
                                        <tr>
                                            <th>
                                                <div class="d-flex align-center">
                                                    <span>{{ tt('Item Title') }}</span>
                                                    <v-spacer/>
                                                    <span>{{ tt('Operation') }}</span>
                                                </div>
                                            </th>
                                        </tr>
                                        </thead>

                                        <tbody v-if="loading && noAvailableItem && !newItem">
                                        <tr :key="itemIdx" v-for="itemIdx in [ 1, 2, 3, 4, 5 ]">
                                            <td class="px-0">
                                                <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                                            </td>
                                        </tr>
                                        </tbody>

                                        <tbody v-if="!loading && noAvailableItem && !newItem">
                                        <tr>
                                            <td>{{ tt('No available item') }}</td>
                                        </tr>
                                        </tbody>

                                        <draggable-list tag="tbody"
                                                        item-key="id"
                                                        handle=".drag-handle"
                                                        ghost-class="dragging-item"
                                                        :class="{ 'has-bottom-border': newItem }"
                                                        :disabled="noAvailableItem"
                                                        v-model="items"
                                                        @change="onMove">
                                            <template #item="{ element }">
                                                <tr class="transaction-items-table-row-item text-sm" v-if="showHidden || !element.hidden">
                                                    <td>
                                                        <div class="d-flex align-center">
                                                            <div class="d-flex align-center" v-if="editingItem.id !== element.id">
                                                                <v-badge class="right-bottom-icon" color="secondary"
                                                                         location="bottom right" offset-x="8" :icon="mdiEyeOffOutline"
                                                                         v-if="element.hidden">
                                                                    <v-icon size="20" start :icon="mdiFormatListBulleted"/>
                                                                </v-badge>
                                                                <v-icon size="20" start :icon="mdiFormatListBulleted" v-else-if="!element.hidden"/>
                                                                <span class="transaction-item-name">{{ element.name }}</span>
                                                            </div>

                                                            <v-text-field class="w-100 me-2" type="text"
                                                                          density="compact" variant="underlined"
                                                                          :disabled="loading || updating"
                                                                          :placeholder="tt('Item Title')"
                                                                          v-model="editingItem.name"
                                                                          v-else-if="editingItem.id === element.id"
                                                                          @keyup.enter="save(editingItem)"
                                                            >
                                                                <template #prepend>
                                                                    <v-badge class="right-bottom-icon" color="secondary"
                                                                             location="bottom right" offset-x="8" :icon="mdiEyeOffOutline"
                                                                             v-if="element.hidden">
                                                                        <v-icon size="20" start :icon="mdiFormatListBulleted"/>
                                                                    </v-badge>
                                                                    <v-icon size="20" start :icon="mdiFormatListBulleted" v-else-if="!element.hidden"/>
                                                                </template>
                                                            </v-text-field>

                                                            <v-spacer/>

                                                            <v-btn class="px-2 ms-2" color="default"
                                                                   density="comfortable" variant="text"
                                                                   :class="{ 'd-none': loading, 'hover-display': !loading }"
                                                                   :prepend-icon="element.hidden ? mdiEyeOutline : mdiEyeOffOutline"
                                                                   :loading="itemHiding[element.id]"
                                                                   :disabled="loading || updating"
                                                                   v-if="editingItem.id !== element.id"
                                                                   @click="hide(element, !element.hidden)">
                                                                <template #loader>
                                                                    <v-progress-circular indeterminate size="20" width="2"/>
                                                                </template>
                                                                {{ element.hidden ? tt('Show') : tt('Hide') }}
                                                            </v-btn>
                                                            <v-btn class="px-2" color="default"
                                                                   density="comfortable" variant="text"
                                                                   :class="{ 'd-none': loading, 'hover-display': !loading }"
                                                                   :prepend-icon="mdiFolderMoveOutline"
                                                                   :loading="itemMoving[element.id]"
                                                                   :disabled="loading || updating || allItemGroupsWithDefault.length < 2"
                                                                   v-if="editingItem.id !== element.id">
                                                                <template #loader>
                                                                    <v-progress-circular indeterminate size="20" width="2"/>
                                                                </template>
                                                                {{ tt('Move') }}
                                                                <v-menu activator="parent" max-height="500">
                                                                    <v-list>
                                                                        <v-list-subheader :title="tt('Move to...')"/>
                                                                        <template :key="itemGroup.id" v-for="itemGroup in allItemGroupsWithDefault">
                                                                            <v-list-item class="text-sm" density="compact"
                                                                                         :value="itemGroup.id" v-if="activeItemGroupId !== itemGroup.id">
                                                                                <v-list-item-title class="cursor-pointer"
                                                                                                   @click="moveItemToGroup(element, itemGroup.id)">
                                                                                    <div class="d-flex align-center">
                                                                                        <span class="text-sm ms-3">{{ itemGroup.name }}</span>
                                                                                    </div>
                                                                                </v-list-item-title>
                                                                            </v-list-item>
                                                                        </template>
                                                                    </v-list>
                                                                </v-menu>
                                                            </v-btn>
                                                            <v-btn class="px-2" color="default"
                                                                   density="comfortable" variant="text"
                                                                   :class="{ 'd-none': loading, 'hover-display': !loading }"
                                                                   :prepend-icon="mdiPencilOutline"
                                                                   :loading="itemUpdating[element.id]"
                                                                   :disabled="loading || updating"
                                                                   v-if="editingItem.id !== element.id"
                                                                   @click="edit(element)">
                                                                <template #loader>
                                                                    <v-progress-circular indeterminate size="20" width="2"/>
                                                                </template>
                                                                {{ tt('Edit') }}
                                                            </v-btn>
                                                            <v-btn class="px-2" color="default"
                                                                   density="comfortable" variant="text"
                                                                   :class="{ 'd-none': loading, 'hover-display': !loading }"
                                                                   :prepend-icon="mdiDeleteOutline"
                                                                   :loading="itemRemoving[element.id]"
                                                                   :disabled="loading || updating"
                                                                   v-if="editingItem.id !== element.id"
                                                                   @click="remove(element)">
                                                                <template #loader>
                                                                    <v-progress-circular indeterminate size="20" width="2"/>
                                                                </template>
                                                                {{ tt('Delete') }}
                                                            </v-btn>
                                                            <v-btn class="px-2"
                                                                   density="comfortable" variant="text"
                                                                   :prepend-icon="mdiCheck"
                                                                   :loading="itemUpdating[element.id]"
                                                                   :disabled="loading || updating || !isItemModified(element)"
                                                                   v-if="editingItem.id === element.id" @click="save(editingItem)">
                                                                <template #loader>
                                                                    <v-progress-circular indeterminate size="20" width="2"/>
                                                                </template>
                                                                {{ tt('Save') }}
                                                            </v-btn>
                                                            <v-btn class="px-2" color="default"
                                                                   density="comfortable" variant="text"
                                                                   :prepend-icon="mdiClose"
                                                                   :disabled="loading || updating"
                                                                   v-if="editingItem.id === element.id" @click="cancelSave(editingItem)">
                                                                {{ tt('Cancel') }}
                                                            </v-btn>
                                                            <span class="ms-2">
                                                                <v-icon :class="!loading && !updating && !hasEditingItem && availableItemCount > 1 ? 'drag-handle' : 'disabled'"
                                                                        :icon="mdiDrag"/>
                                                                <v-tooltip activator="parent" v-if="!loading && !updating && !hasEditingItem && availableItemCount > 1">{{ tt('Drag to Reorder') }}</v-tooltip>
                                                            </span>
                                                        </div>
                                                    </td>
                                                </tr>
                                            </template>
                                        </draggable-list>

                                        <tbody v-if="newItem">
                                        <tr class="text-sm" :class="{ 'even-row': (availableItemCount & 1) === 1}">
                                            <td>
                                                <div class="d-flex align-center">
                                                    <v-text-field class="w-100 me-2" type="text" color="primary"
                                                                  density="compact" variant="underlined"
                                                                  :disabled="loading || updating" :placeholder="tt('Item Title')"
                                                                  v-model="newItem.name" @keyup.enter="save(newItem)">
                                                        <template #prepend>
                                                            <v-icon size="20" start :icon="mdiFormatListBulleted"/>
                                                        </template>
                                                    </v-text-field>

                                                    <v-spacer/>

                                                    <v-btn class="px-2" density="comfortable" variant="text"
                                                           :prepend-icon="mdiCheck"
                                                           :loading="itemUpdating['']"
                                                           :disabled="loading || updating || !isItemModified(newItem)"
                                                           @click="save(newItem)">
                                                        <template #loader>
                                                            <v-progress-circular indeterminate size="20" width="2"/>
                                                        </template>
                                                        {{ tt('Save') }}
                                                    </v-btn>
                                                    <v-btn class="px-2" color="default"
                                                           density="comfortable" variant="text"
                                                           :prepend-icon="mdiClose"
                                                           :disabled="loading || updating"
                                                           @click="cancelSave(newItem)">
                                                        {{ tt('Cancel') }}
                                                    </v-btn>
                                                    <span class="ms-2">
                                                        <v-icon class="disabled" :icon="mdiDrag"/>
                                                    </span>
                                                </div>
                                            </td>
                                        </tr>
                                        </tbody>
                                    </v-table>
                                </v-card>
                            </v-window-item>
                        </v-window>
                    </v-main>
                </v-layout>
            </v-card>
        </v-col>
    </v-row>

    <item-group-change-display-order-dialog ref="itemGroupChangeDisplayOrderDialog" />

    <rename-dialog ref="renameDialog"
                   :default-title="tt('Rename Item Group')"
                   :label="tt('Item Group Name')" :placeholder="tt('Item Group Name')" />
    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import ItemGroupChangeDisplayOrderDialog from './dialog/ItemGroupChangeDisplayOrderDialog.vue';
import RenameDialog from '@/components/desktop/RenameDialog.vue';
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, useTemplateRef, watch } from 'vue';
import { useDisplay } from 'vuetify';

import { useI18n } from '@/locales/helpers.ts';
import { useItemListPageBase } from '@/views/base/items/ItemListPageBase.ts';

import { useTransactionItemsStore } from '@/stores/transactionItem.ts';

import { DEFAULT_ITEM_GROUP_ID } from '@/consts/item.ts';

import { TransactionItemGroup } from '@/models/transaction_item_group.ts';
import { TransactionItem } from '@/models/transaction_item.ts';

import { getAvailableItemCount } from '@/lib/item.ts';

import {
    mdiRefresh,
    mdiMenuUp,
    mdiMenuDown,
    mdiPencilOutline,
    mdiCheck,
    mdiClose,
    mdiEyeOffOutline,
    mdiEyeOutline,
    mdiSort,
    mdiMenu,
    mdiPlus,
    mdiFolderMoveOutline,
    mdiDeleteOutline,
    mdiDrag,
    mdiDotsVertical,
    mdiFormatListBulleted
} from '@mdi/js';

type ItemGroupChangeDisplayOrderDialogType = InstanceType<typeof ItemGroupChangeDisplayOrderDialog>;
type RenameDialogType = InstanceType<typeof RenameDialog>;
type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const display = useDisplay();

const { tt, formatNumberToLocalizedNumerals } = useI18n();

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

const itemGroupChangeDisplayOrderDialog = useTemplateRef<ItemGroupChangeDisplayOrderDialogType>('itemGroupChangeDisplayOrderDialog');
const renameDialog = useTemplateRef<RenameDialogType>('renameDialog');
const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');

const updating = ref<boolean>(false);
const activeTab = ref<string>('itemListPage');
const alwaysShowNav = ref<boolean>(display.mdAndUp.value);
const showNav = ref<boolean>(display.mdAndUp.value);
const itemUpdating = ref<Record<string, boolean>>({});
const itemHiding = ref<Record<string, boolean>>({});
const itemMoving = ref<Record<string, boolean>>({});
const itemRemoving = ref<Record<string, boolean>>({});

const totalAvailableItemsCount = computed<number>(() => transactionItemsStore.allAvailableItemsCount);
const displayTotalAvailableItemsCount = computed<string>(() => formatNumberToLocalizedNumerals(transactionItemsStore.allAvailableItemsCount));
const availableItemCount = computed<number>(() => getAvailableItemCount(items.value, showHidden.value));

function reload(): void {
    if (hasEditingItem.value) {
        return;
    }

    loading.value = true;

    transactionItemsStore.loadAllItems({
        force: true
    }).then(() => {
        loading.value = false;
        displayOrderModified.value = false;

        snackbar.value?.showMessage('Item list has been updated');
    }).catch(error => {
        loading.value = false;

        if (error && error.isUpToDate) {
            displayOrderModified.value = false;
        }

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function addItemGroup(): void {
    renameDialog.value?.open('', tt('New Item Group Name')).then((newName: string) => {
        updating.value = true;

        transactionItemsStore.saveItemGroup({
            itemGroup: TransactionItemGroup.createNewItemGroup(newName)
        }).then(itemGroup => {
            updating.value = false;
            activeItemGroupId.value = itemGroup.id;
        }).catch(error => {
            updating.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function renameItemGroup(): void {
    const itemGroup = transactionItemsStore.allTransactionItemGroupsMap[activeItemGroupId.value];

    if (!itemGroup) {
        snackbar.value?.showMessage('Unable to rename this item group');
        return;
    }

    renameDialog.value?.open(itemGroup.name || '').then((newName: string) => {
        updating.value = true;

        const newItemGroup = itemGroup.clone();
        newItemGroup.name = newName;

        transactionItemsStore.saveItemGroup({
            itemGroup: newItemGroup
        }).then(() => {
            updating.value = false;
        }).catch(error => {
            updating.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function showChangeGroupDisplayOrderDialog(): void {
    itemGroupChangeDisplayOrderDialog.value?.open().then(() => {
        if (transactionItemsStore.transactionItemGroupListStateInvalid) {
            loading.value = true;

            transactionItemsStore.loadAllItemGroups({
                force: false
            }).then(() => {
                loading.value = false;
            }).catch(() => {
                loading.value = false;
            });
        }
    });
}

function removeItemGroup(): void {
    const itemGroup = transactionItemsStore.allTransactionItemGroupsMap[activeItemGroupId.value];

    if (!itemGroup) {
        snackbar.value?.showMessage('Unable to delete this item group');
        return;
    }

    const currentItemGroupIndex = allItemGroupsWithDefault.value.findIndex(group => group.id === itemGroup.id);

    confirmDialog.value?.open('Are you sure you want to delete this item group?').then(() => {
        updating.value = true;

        transactionItemsStore.deleteItemGroup({
            itemGroup: itemGroup
        }).then(() => {
            updating.value = false;

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
            updating.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function moveItemToGroup(item: TransactionItem, targetItemGroupId: string): void {
    updating.value = true;
    itemMoving.value[item.id] = true;

    const newItemObj = item.clone();
    newItemObj.groupId = targetItemGroupId;

    transactionItemsStore.saveItem({
        item: newItemObj
    }).then(() => {
        updating.value = false;
        itemMoving.value[item.id] = false;
    }).catch(error => {
        updating.value = false;
        itemMoving.value[item.id] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function save(item: TransactionItem): void {
    updating.value = true;
    itemUpdating.value[item.id || ''] = true;

    transactionItemsStore.saveItem({
        item: item
    }).then(() => {
        updating.value = false;
        itemUpdating.value[item.id || ''] = false;

        if (item.id) {
            editingItem.value.id = '';
            editingItem.value.name = '';
        } else {
            newItem.value = null;
        }
    }).catch(error => {
        updating.value = false;
        itemUpdating.value[item.id || ''] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
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

function saveSortResult(): void {
    if (!displayOrderModified.value) {
        return;
    }

    loading.value = true;

    transactionItemsStore.updateItemDisplayOrders(activeItemGroupId.value).then(() => {
        loading.value = false;
        displayOrderModified.value = false;
    }).catch(error => {
        loading.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function hide(item: TransactionItem, hidden: boolean): void {
    updating.value = true;
    itemHiding.value[item.id] = true;

    transactionItemsStore.hideItem({
        item: item,
        hidden: hidden
    }).then(() => {
        updating.value = false;
        itemHiding.value[item.id] = false;
    }).catch(error => {
        updating.value = false;
        itemHiding.value[item.id] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function remove(item: TransactionItem): void {
    confirmDialog.value?.open('Are you sure you want to delete this item?').then(() => {
        updating.value = true;
        itemRemoving.value[item.id] = true;

        transactionItemsStore.deleteItem({
            item: item
        }).then(() => {
            updating.value = false;
            itemRemoving.value[item.id] = false;
        }).catch(error => {
            updating.value = false;
            itemRemoving.value[item.id] = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function onMove(event: { moved: { element: { id: string }; oldIndex: number; newIndex: number } }): void {
    if (!event || !event.moved) {
        return;
    }

    const moveEvent = event.moved;

    if (!moveEvent.element || !moveEvent.element.id) {
        snackbar.value?.showMessage('Unable to move item');
        return;
    }

    transactionItemsStore.changeItemDisplayOrder({
        itemId: moveEvent.element.id,
        from: moveEvent.oldIndex,
        to: moveEvent.newIndex
    }).then(() => {
        displayOrderModified.value = true;
    }).catch(error => {
        snackbar.value?.showError(error);
    });
}

transactionItemsStore.loadAllItems({
    force: false
}).then(() => {
    loading.value = false;
}).catch(error => {
    loading.value = false;

    if (!error.processed) {
        snackbar.value?.showError(error);
    }
});

watch(() => display.mdAndUp.value, (newValue) => {
    alwaysShowNav.value = newValue;

    if (!showNav.value) {
        showNav.value = newValue;
    }
});
</script>

<style>
.transaction-items-statistic-item-value {
    font-size: 1rem;
}

.transaction-items-table tr.transaction-items-table-row-item .hover-display {
    display: none;
}

.transaction-items-table tr.transaction-items-table-row-item:hover .hover-display {
    display: inline-grid;
}

.transaction-items-table tr:not(:last-child) > td > div {
    padding-bottom: 1px;
}

.transaction-items-table .has-bottom-border tr:last-child > td > div {
    padding-bottom: 1px;
}

.transaction-items-table tr.transaction-items-table-row-item .right-bottom-icon .v-badge__badge {
    padding-bottom: 1px;
}

.transaction-items-table .v-text-field .v-input__prepend {
    margin-inline-end: 0;
    color: rgba(var(--v-theme-on-surface));
}

.transaction-items-table .v-text-field .v-input__prepend .v-badge > .v-badge__wrapper > .v-icon {
    opacity: var(--v-medium-emphasis-opacity);
}

.transaction-items-table .v-text-field.v-input--plain-underlined .v-input__prepend {
    padding-top: 10px;
}

.transaction-items-table .v-text-field .v-field__input {
    font-size: 0.875rem;
    padding-top: 0;
    color: rgba(var(--v-theme-on-surface));
}

.transaction-items-table .transaction-item-name {
    font-size: 0.875rem;
}

.transaction-items-table tr .v-text-field .v-field__input {
    padding-bottom: 1px;
}
</style>

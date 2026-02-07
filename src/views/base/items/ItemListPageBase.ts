import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useTransactionItemsStore } from '@/stores/transactionItem.ts';

import { DEFAULT_ITEM_GROUP_ID } from '@/consts/item.ts';

import { TransactionItemGroup } from '@/models/transaction_item_group.ts';
import { TransactionItem } from '@/models/transaction_item.ts';

import { isNoAvailableItem } from '@/lib/item.ts';

export function useItemListPageBase() {
    const { tt } = useI18n();

    const transactionItemsStore = useTransactionItemsStore();

    const activeItemGroupId = ref<string>(DEFAULT_ITEM_GROUP_ID);
    const newItem = ref<TransactionItem | null>(null);
    const editingItem = ref<TransactionItem>(TransactionItem.createNewItem());
    const loading = ref<boolean>(true);
    const showHidden = ref<boolean>(false);
    const displayOrderModified = ref<boolean>(false);

    const allItemGroupsWithDefault = computed<TransactionItemGroup[]>(() => {
        const allGroups: TransactionItemGroup[] = [];
        const defaultGroup = TransactionItemGroup.createNewItemGroup(tt('Default Group'));
        defaultGroup.id = DEFAULT_ITEM_GROUP_ID;
        allGroups.push(defaultGroup);
        allGroups.push(...transactionItemsStore.allTransactionItemGroups);
        return allGroups;
    });
    const items = computed<TransactionItem[]>(() => transactionItemsStore.allTransactionItemsByGroupMap[activeItemGroupId.value] || []);

    const noAvailableItem = computed<boolean>(() => isNoAvailableItem(items.value, showHidden.value));
    const hasEditingItem = computed<boolean>(() => !!(newItem.value || (editingItem.value.id && editingItem.value.id !== '')));

    function isItemModified(item: TransactionItem): boolean {
        if (item.id) {
            return editingItem.value.name !== '' && editingItem.value.name !== item.name;
        } else {
            return item.name !== '';
        }
    }

    function switchItemGroup(itemGroupId: string): void {
        activeItemGroupId.value = itemGroupId;

        if (newItem.value) {
            newItem.value.groupId = itemGroupId;
        }
    }

    function add(): void {
        newItem.value = TransactionItem.createNewItem('', activeItemGroupId.value);
    }

    function edit(item: TransactionItem): void {
        editingItem.value.id = item.id;
        editingItem.value.groupId = item.groupId;
        editingItem.value.name = item.name;
    }

    return {
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
    };
}

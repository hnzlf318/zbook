import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useTransactionItemsStore } from '@/stores/transactionItem.ts';

import { DEFAULT_ITEM_GROUP_ID } from '@/consts/item.ts';

import { TransactionItem } from '@/models/transaction_item.ts';

export type TransactionItemWithGroupHeader = TransactionItem | {
    type: 'subheader';
    title: string;
};

export interface CommonTransactionItemSelectionProps {
    modelValue: string[];
}

export function useTransactionItemSelectionBase(props: CommonTransactionItemSelectionProps, useClonedModelValue?: boolean) {
    const { tt } = useI18n();

    const transactionItemsStore = useTransactionItemsStore();

    const clonedModelValue = ref<string[]>(useClonedModelValue ? Array.from(props.modelValue) : []);
    const itemSearchContent = ref<string>('');

    const selectedItemIds = computed<Record<string, boolean>>(() => {
        const ret: Record<string, boolean> = {};

        if (useClonedModelValue) {
            for (const itemId of clonedModelValue.value) {
                ret[itemId] = true;
            }
        } else {
            for (const itemId of props.modelValue) {
                ret[itemId] = true;
            }
        }

        return ret;
    });

    const lowerCaseItemSearchContent = computed<string>(() => itemSearchContent.value.toLowerCase());

    const allItemsWithGroupHeader = computed<TransactionItemWithGroupHeader[]>(() => getItemsWithGroupHeader(item => {
        if (!item.hidden) {
            return true;
        }

        if (selectedItemIds.value[item.id]) {
            return true;
        }

        if (lowerCaseItemSearchContent.value && item.name.toLowerCase().indexOf(lowerCaseItemSearchContent.value) >= 0) {
            return true;
        }

        return false;
    }));

    function getItemsWithGroupHeader(itemFilterFn: (item: TransactionItem) => boolean): TransactionItemWithGroupHeader[] {
        const result: TransactionItemWithGroupHeader[] = [];
        const itemsInDefaultGroup = transactionItemsStore.allTransactionItemsByGroupMap[DEFAULT_ITEM_GROUP_ID];

        if (itemsInDefaultGroup && itemsInDefaultGroup.length > 0) {
            const visibleItems = itemsInDefaultGroup.filter(item => itemFilterFn(item));

            if (visibleItems.length > 0) {
                result.push({
                    type: 'subheader',
                    title: tt('Default Group')
                });

                result.push(...visibleItems);
            }
        }

        for (const itemGroup of transactionItemsStore.allTransactionItemGroups) {
            const items = transactionItemsStore.allTransactionItemsByGroupMap[itemGroup.id];

            if (!items || items.length < 1) {
                continue;
            }

            const visibleItems = items.filter(item => itemFilterFn(item));

            if (visibleItems.length > 0) {
                result.push({
                    type: 'subheader',
                    title: itemGroup.name
                });

                result.push(...visibleItems);
            }
        }

        return result;
    }

    const filteredItemsWithGroupHeader = computed<TransactionItemWithGroupHeader[]>(() => getItemsWithGroupHeader(item => {
        if (lowerCaseItemSearchContent.value) {
            if (item.name.toLowerCase().indexOf(lowerCaseItemSearchContent.value) >= 0 && (!item.hidden || selectedItemIds.value[item.id])) {
                return true;
            }
            return false;
        }
        return !item.hidden || !!selectedItemIds.value[item.id];
    }));

    return {
        clonedModelValue,
        itemSearchContent,
        allItemsWithGroupHeader,
        filteredItemsWithGroupHeader,
        selectedItemIds
    };
}

<template>
    <f7-sheet ref="sheet" swipe-to-close swipe-handler=".swipe-handler"
              style="height: auto" :opened="show"
              @sheet:open="onSheetOpen" @sheet:closed="onSheetClosed">
        <f7-toolbar class="toolbar-with-swipe-handler">
            <div class="swipe-handler"></div>
            <div class="left">
                <f7-link sheet-close icon-f7="xmark"></f7-link>
            </div>
            <f7-searchbar ref="searchbar" custom-search
                          :value="itemSearchContent"
                          :placeholder="tt('Find item')"
                          :disable-button="false"
                          v-if="enableFilter"
                          @input="itemSearchContent = ($event.target as HTMLInputElement).value">
            </f7-searchbar>
            <div class="right">
                <f7-button round fill icon-f7="checkmark_alt" @click="save"
                           v-if="filteredItemsWithGroupHeader && filteredItemsWithGroupHeader.length > 0"></f7-button>
            </div>
        </f7-toolbar>
        <f7-page-content :class="'margin-top ' + heightClass">
            <f7-list class="no-margin-top no-margin-bottom" v-if="!filteredItemsWithGroupHeader || filteredItemsWithGroupHeader.length < 1">
                <f7-list-item :title="tt('No available item')"></f7-list-item>
            </f7-list>
            <f7-list dividers class="no-margin-top no-margin-bottom item-selection-list" v-else>
                <template :key="(item instanceof TransactionItem) ? item.id : (item.type === 'subheader' ? `${item.type}-${index}-${(item as { title: string }).title}` : `${index}`)"
                          v-for="(item, index) in filteredItemsWithGroupHeader">
                    <f7-list-item group-title v-if="!(item instanceof TransactionItem) && (item as { type?: string }).type === 'subheader'">
                        <div class="item-selection-list-item">
                            {{ (item as { title: string }).title }}
                        </div>
                    </f7-list-item>
                    <f7-list-item checkbox
                                  :class="{ 'list-item-selected': selectedItemIds[item.id], 'disabled': item.hidden && !selectedItemIds[item.id] }"
                                  :value="item.id"
                                  :checked="selectedItemIds[item.id]"
                                  :key="item.id"
                                  v-else-if="item instanceof TransactionItem"
                                  @change="changeItemSelection">
                        <template #media>
                            <f7-icon class="transaction-item-icon" f7="list_bullet">
                                <f7-badge color="gray" class="right-bottom-icon" v-if="item.hidden">
                                    <f7-icon f7="eye_slash_fill"></f7-icon>
                                </f7-badge>
                            </f7-icon>
                        </template>
                        <template #title>
                            <div class="display-flex">
                                <div class="item-selection-list-item list-item-valign-middle padding-inline-start-half">
                                    {{ item.name }}
                                </div>
                            </div>
                        </template>
                    </f7-list-item>
                </template>
            </f7-list>
        </f7-page-content>
    </f7-sheet>
</template>

<script setup lang="ts">
import { computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { type CommonTransactionItemSelectionProps, useTransactionItemSelectionBase } from '@/components/base/TransactionItemSelectionBase.ts';
import { TransactionItem } from '@/models/transaction_item.ts';
import { scrollToSelectedItem } from '@/lib/ui/common.ts';
import { type Framework7Dom } from '@/lib/ui/mobile.ts';

interface TransactionItemSelectionSheetProps extends CommonTransactionItemSelectionProps {
    enableFilter?: boolean;
    show: boolean;
}

const props = defineProps<TransactionItemSelectionSheetProps>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: string[]): void;
    (e: 'update:show', value: boolean): void;
}>();

const { tt } = useI18n();

const {
    clonedModelValue,
    itemSearchContent,
    selectedItemIds,
    filteredItemsWithGroupHeader
} = useTransactionItemSelectionBase(props, true);

const heightClass = computed<string>(() => {
    const len = filteredItemsWithGroupHeader.value.length;
    if (len > 6) {
        return 'item-selection-huge-sheet';
    }
    if (len > 3) {
        return 'item-selection-large-sheet';
    }
    return 'item-selection-default-sheet';
});

function changeItemSelection(e: Event): void {
    const target = e.target as HTMLInputElement;
    const itemId = target.value;
    const index = clonedModelValue.value.indexOf(itemId);

    if (target.checked) {
        if (index < 0) {
            clonedModelValue.value.push(itemId);
        }
    } else {
        if (index >= 0) {
            clonedModelValue.value.splice(index, 1);
        }
    }
}

function save(): void {
    emit('update:modelValue', clonedModelValue.value);
    emit('update:show', false);
}

function onSheetOpen(event: { $el: Framework7Dom }): void {
    clonedModelValue.value = Array.from(props.modelValue);
    scrollToSelectedItem(event.$el[0], '.sheet-modal-inner', '.page-content', 'li.list-item-selected');
}

function onSheetClosed(): void {
    emit('update:show', false);
}
</script>

<style scoped>
.item-selection-list-item {
    overflow: hidden;
    text-overflow: ellipsis;
}

.item-selection-default-sheet {
    height: 40vh;
}

.item-selection-large-sheet {
    height: 60vh;
}

.item-selection-huge-sheet {
    height: 80vh;
}
</style>

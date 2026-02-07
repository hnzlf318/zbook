<template>
    <f7-page @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')" v-if="!displayOrderModified"></f7-nav-left>
            <f7-nav-left v-else-if="displayOrderModified">
                <f7-link icon-f7="xmark" :class="{ 'disabled': displayOrderSaving }" @click="cancelSort"></f7-link>
            </f7-nav-left>
            <f7-nav-title :title="tt('Change Group Display Order')"></f7-nav-title>
            <f7-nav-right :class="{ 'disabled': loading }">
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': displayOrderSaving || !displayOrderModified }" @click="saveSortResult"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers class="item-group-item-list margin-top skeleton-text" v-if="loading">
            <f7-list-item :key="itemIdx" v-for="itemIdx in [ 1, 2, 3 ]">
                <template #title>
                    <div class="display-flex">
                        <div class="transaction-item-group-list-item-content list-item-valign-middle padding-inline-start-half">Item Group Name</div>
                    </div>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="item-group-item-list margin-top" v-if="!loading && itemGroups.length < 1">
            <f7-list-item :title="tt('No available item group')"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers sortable sortable-enabled
                 class="item-group-item-list margin-top"
                 @sortable:sort="onSort"
                 v-if="!loading">
            <f7-list-item :id="getItemGroupDomId(itemGroup)"
                          :key="itemGroup.id"
                          v-for="itemGroup in itemGroups">
                <template #title>
                    <div class="display-flex">
                        <div class="transaction-item-group-list-item-content list-item-valign-middle padding-inline-start-half">{{ itemGroup.name }}</div>
                    </div>
                </template>
            </f7-list-item>
        </f7-list>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents, showLoading, hideLoading } from '@/lib/ui/mobile.ts';

import { useTransactionItemsStore } from '@/stores/transactionItem.ts';

import { TransactionItemGroup } from '@/models/transaction_item_group.ts';

const props = defineProps<{
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const { showToast, routeBackOnError } = useI18nUIComponents();

const transactionItemsStore = useTransactionItemsStore();

const loading = ref<boolean>(true);
const loadingError = ref<unknown | null>(null);
const displayOrderModified = ref<boolean>(false);
const displayOrderSaving = ref<boolean>(false);

const itemGroups = computed<TransactionItemGroup[]>(() => transactionItemsStore.allTransactionItemGroups);

function getItemGroupDomId(itemGroup: TransactionItemGroup): string {
    return 'itemGroup_' + itemGroup.id;
}

function parseItemGroupIdFromDomId(domId: string): string | null {
    if (!domId || domId.indexOf('itemGroup_') !== 0) {
        return null;
    }

    return domId.substring(9);
}

function init(): void {
    loading.value = true;

    transactionItemsStore.loadAllItemGroups({
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

function saveSortResult(): void {
    if (!displayOrderModified.value) {
        return;
    }

    displayOrderSaving.value = true;
    showLoading();

    transactionItemsStore.updateItemGroupDisplayOrders().then(() => {
        displayOrderSaving.value = false;
        hideLoading();

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
        return;
    }

    displayOrderSaving.value = true;
    showLoading();

    transactionItemsStore.loadAllItemGroups({
        force: false
    }).then(() => {
        displayOrderSaving.value = false;
        hideLoading();

        displayOrderModified.value = false;
    }).catch(error => {
        displayOrderSaving.value = false;
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function onSort(event: { el: { id: string }, from: number, to: number }): void {
    if (!event || !event.el || !event.el.id) {
        showToast('Unable to move item group');
        return;
    }

    const id = parseItemGroupIdFromDomId(event.el.id);

    if (!id) {
        showToast('Unable to move item group');
        return;
    }

    transactionItemsStore.changeItemGroupDisplayOrder({
        itemGroupId: id,
        from: event.from,
        to: event.to
    }).then(() => {
        displayOrderModified.value = true;
    }).catch(error => {
        showToast(error.message || error);
    });
}

function onPageAfterIn(): void {
    if (transactionItemsStore.transactionItemGroupListStateInvalid && !loading.value) {
        transactionItemsStore.loadAllItemGroups({}).catch(error => {
            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }

    routeBackOnError(props.f7router, loadingError);
}

init();
</script>

<style>
.transaction-item-group-list-item-content {
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>

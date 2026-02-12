<template>
    <f7-page>
        <f7-navbar :title="tt('OCR Bill Recognition')" :back-link="tt('Back')">
            <f7-nav-right v-if="recognizedList.length">
                <f7-link icon-f7="photo" :title="tt('Select Another Image')" @click="reset"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-block v-if="!recognizedList.length" class="ocr-select-block">
            <p class="ocr-hint">{{ tt('You can select a bill or transaction list screenshot to recognize.') }}</p>
            <f7-button large fill @click="triggerFileInput">{{ recognizing ? tt('Recognizing...') : tt('Select Image') }}</f7-button>
            <input ref="imageInputRef" type="file" class="display-none" :accept="SUPPORTED_IMAGE_EXTENSIONS" @change="onFileChange" />
        </f7-block>

        <f7-block v-else>
            <p class="ocr-hint">{{ tt('Recognized transactions (click Add to confirm and edit):') }}</p>
            <f7-list strong inset dividers>
                <f7-list-item v-for="(item, idx) in recognizedList" :key="idx" class="ocr-result-item">
                    <template #content>
                        <div class="ocr-item-inner">
                            <div class="ocr-fields">
                                <div class="ocr-field"><span class="ocr-label">{{ tt('Income or Expense') }}</span><span class="ocr-value">{{ getTypeLabel(item.type) }}</span></div>
                                <div class="ocr-field"><span class="ocr-label">{{ tt('Amount') }}</span><span class="ocr-value">{{ formatAmount(item.sourceAmount) }}</span></div>
                                <div class="ocr-field"><span class="ocr-label">{{ tt('Date') }}</span><span class="ocr-value">{{ formatTime(item.time) }}</span></div>
                                <div class="ocr-field"><span class="ocr-label">{{ tt('Category') }}</span><span class="ocr-value">{{ getCategoryName(item.categoryId) || '-' }}</span></div>
                                <div class="ocr-field"><span class="ocr-label">{{ tt('Account') }}</span><span class="ocr-value">{{ getAccountName(item.sourceAccountId) || '-' }}</span></div>
                                <div class="ocr-field" v-if="getItemNames(item.itemIds)"><span class="ocr-label">{{ tt('Transaction Items') }}</span><span class="ocr-value">{{ getItemNames(item.itemIds) }}</span></div>
                                <div class="ocr-field" v-if="getTagNames(item.tagIds)"><span class="ocr-label">{{ tt('Tags') }}</span><span class="ocr-value">{{ getTagNames(item.tagIds) }}</span></div>
                                <div class="ocr-field" v-if="item.comment"><span class="ocr-label">{{ tt('Description') }}</span><span class="ocr-value ocr-desc">{{ item.comment }}</span></div>
                            </div>
                            <div class="ocr-add-wrap">
                                <f7-button small fill color="blue" class="ocr-add-btn" :disabled="addedRowIndices.has(idx)" @click="onAdd(item, idx)">
                                    {{ tt('Add') }}
                                </f7-button>
                            </div>
                        </div>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-block>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, useTemplateRef } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTransactionItemsStore } from '@/stores/transactionItem.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';

import { TransactionType } from '@/core/transaction.ts';
import { SUPPORTED_IMAGE_EXTENSIONS } from '@/consts/file.ts';

import type { RecognizedReceiptImageResponse } from '@/models/large_language_model.ts';

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const { showToast } = useI18nUIComponents();

const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();
const transactionTagsStore = useTransactionTagsStore();
const transactionItemsStore = useTransactionItemsStore();
const transactionsStore = useTransactionsStore();

const imageInputRef = useTemplateRef<HTMLInputElement>('imageInputRef');

const recognizing = ref(false);
const recognizedList = ref<RecognizedReceiptImageResponse[]>([]);
const addedRowIndices = ref<Set<number>>(new Set());

function getCategoryName(categoryId?: string): string {
    if (!categoryId) return '-';
    const cat = transactionCategoriesStore.allTransactionCategoriesMap[categoryId];
    return cat?.name ?? '-';
}

function getAccountName(accountId?: string): string {
    if (!accountId) return '-';
    const acc = accountsStore.allAccountsMap[accountId];
    return acc?.name ?? '-';
}

function getTagNames(tagIds?: string[]): string {
    if (!tagIds?.length) return '';
    return tagIds.map(id => transactionTagsStore.allTransactionTagsMap[id]?.name ?? id).join(', ');
}

function getItemNames(itemIds?: string[]): string {
    if (!itemIds?.length) return '';
    const itemsMap = transactionItemsStore.allTransactionItemsMap;
    return itemIds.map(id => itemsMap[id]?.name ?? id).join(', ');
}

function getTypeLabel(type?: number): string {
    if (type === TransactionType.Income) return tt('Income');
    if (type === TransactionType.Expense) return tt('Expense');
    return '-';
}

function formatAmount(amount?: number): string {
    if (amount === undefined || amount === null) return '-';
    const n = amount / 100;
    return n >= 0 ? `¥${n.toFixed(2)}` : `-¥${(-n).toFixed(2)}`;
}

function formatTime(time?: number): string {
    if (time === undefined || time === null) return '-';
    const d = new Date(time * 1000);
    return d.toLocaleString(undefined, { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}

function triggerFileInput(): void {
    if (recognizing.value) return;
    imageInputRef.value?.click();
}

function onFileChange(event: Event): void {
    const el = event?.target as HTMLInputElement;
    if (!el?.files?.length || !el.files[0]) return;
    const file = el.files[0] as File;
    el.value = '';
    recognizing.value = true;
    Promise.all([
        accountsStore.loadAllAccounts({ force: false }),
        transactionCategoriesStore.loadAllCategories({ force: false }),
        transactionTagsStore.loadAllTags({ force: false }),
        transactionItemsStore.loadAllItems({ force: false }),
        transactionsStore.recognizeReceiptImageByOCR(file)
    ]).then((results) => {
        recognizedList.value = results[4] as RecognizedReceiptImageResponse[];
        recognizing.value = false;
    }).catch((error) => {
        recognizing.value = false;
        const msg = error?.message ?? (typeof error === 'string' ? error : 'Recognize failed');
        showToast(msg);
    });
}

function reset(): void {
    recognizedList.value = [];
    addedRowIndices.value = new Set();
}

function buildAddUrl(item: RecognizedReceiptImageResponse): string {
    const params = new URLSearchParams();
    if (item.time != null) params.set('time', String(item.time));
    if (item.type != null) params.set('type', String(item.type));
    if (item.categoryId) params.set('categoryId', item.categoryId);
    if (item.sourceAccountId) params.set('accountId', item.sourceAccountId);
    if (item.destinationAccountId) params.set('destinationAccountId', item.destinationAccountId);
    if (item.sourceAmount != null) params.set('amount', String(item.sourceAmount));
    if (item.destinationAmount != null) params.set('destinationAmount', String(item.destinationAmount));
    if (item.tagIds?.length) params.set('tagIds', item.tagIds.join(','));
    if (item.itemIds?.length) params.set('itemIds', item.itemIds.join(','));
    if (item.comment) params.set('comment', item.comment);
    params.set('noTransactionDraft', 'true');
    return `/transaction/add?${params.toString()}`;
}

function onAdd(item: RecognizedReceiptImageResponse, idx: number): void {
    if (addedRowIndices.value.has(idx)) return;
    addedRowIndices.value = new Set(addedRowIndices.value).add(idx);
    const url = buildAddUrl(item);
    props.f7router.navigate(url);
}
</script>

<style scoped>
.display-none {
    display: none;
}
.ocr-select-block {
    padding-top: 2rem;
    text-align: center;
}
.ocr-hint {
    margin-bottom: 1rem;
    color: var(--f7-block-title-medium-text-color);
}
.ocr-result-item {
    align-items: flex-start;
}
.ocr-item-inner {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    width: 100%;
}
.ocr-fields {
    text-align: left;
}
.ocr-field {
    display: flex;
    align-items: baseline;
    gap: 0.35rem;
    margin-bottom: 0.25rem;
}
.ocr-label {
    flex-shrink: 0;
    font-size: 0.85em;
    color: var(--f7-block-title-medium-text-color);
}
.ocr-value {
    min-width: 0;
    font-size: 0.9em;
}
.ocr-value.ocr-desc {
    word-break: break-word;
}
.ocr-add-wrap {
    align-self: flex-end;
    margin-top: 0.5rem;
}
.ocr-add-btn { flex-shrink: 0; }
</style>

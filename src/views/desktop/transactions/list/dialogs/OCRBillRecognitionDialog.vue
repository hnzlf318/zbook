<template>
    <v-dialog  :persistent="loading || recognizing || !!imageFile" v-model="showState" @paste="onPaste">
        <v-card class="pa-sm-1 pa-md-2 d-flex flex-column">
            <template #title>
                <h4 class="text-h4">{{ tt('OCR Bill Recognition') }}</h4>
            </template>

            <v-card-text class="d-flex flex-column flex-grow-1 overflow-y-auto ocr-card-body">
                <div v-if="!recognizedList.length" class="w-100 border position-relative ocr-drop-area"
                     @dragenter.prevent="onDragEnter"
                     @dragover.prevent
                     @dragleave.prevent="onDragLeave"
                     @drop.prevent="onDrop">
                    <div class="d-flex w-100 justify-center align-center text-center px-4 py-8"
                         :class="{ 'dropzone': true, 'dropzone-dark': isDarkMode, 'dropzone-blurry-bg': loading || isDragOver || recognizing, 'dropzone-dragover': isDragOver }">
                        <div class="d-inline-flex flex-column dropzone-content" v-if="!loading && !imageFile && !isDragOver"
                             @click="showOpenImageDialog">
                            <h3 class="pa-2">{{ tt('You can drag and drop, paste or click to select a bill or transaction list screenshot') }}</h3>
                            <span class="pa-2 text-medium-emphasis">{{ tt('OCR runs on server (no AI, no tokens). Supports Chinese bill format (e.g. 2月7日 21:49, -100.00).') }}</span>
                        </div>
                        <h3 class="pa-2" v-else-if="!loading && isDragOver">{{ tt('Release to load image') }}</h3>
                        <h3 class="pa-2" v-else-if="loading">{{ tt('Loading image...') }}</h3>
                        <h3 class="pa-2" v-else-if="recognizing">{{ tt('Recognizing...') }}</h3>
                    </div>
                    <v-img v-if="imageSrc" :class="{ 'cursor-pointer': !loading && !recognizing && !isDragOver }"
                           :src="imageSrc" @click="showOpenImageDialog" max-height="280" />
                </div>

                <div v-else class="w-100">
                    <p class="text-body-2 mb-2">{{ tt('Recognized transactions (click Add to confirm and edit):') }}</p>
                    <v-table density="compact" class="ocr-recognized-table">
                        <thead>
                            <tr>
                                <th class="ocr-cell-type">{{ tt('Income or Expense') }}</th>
                                <th class="ocr-cell-amount">{{ tt('Amount') }}</th>
                                <th class="ocr-cell-category">{{ tt('Category') }}</th>
                                <th class="ocr-cell-account">{{ tt('Account') }}</th>
                                <th class="ocr-cell-time">{{ tt('Time') }}</th>
                                <th class="ocr-cell-items">{{ tt('Transaction Items') }}</th>
                                <th class="ocr-cell-tags">{{ tt('Tags') }}</th>
                                <th class="ocr-cell-desc">{{ tt('Description') }}</th>
                                <th class="ocr-cell-actions"></th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="(item, idx) in recognizedList" :key="idx">
                                <td class="ocr-cell-type">{{ getTypeLabel(item.type) }}</td>
                                <td class="ocr-cell-amount text-end">{{ formatAmount(item.sourceAmount) }}</td>
                                <td class="ocr-cell-category">{{ getCategoryName(item.categoryId) || '-' }}</td>
                                <td class="ocr-cell-account">{{ getAccountName(item.sourceAccountId) || '-' }}</td>
                                <td class="ocr-cell-time text-end">{{ formatTime(item.time) }}</td>
                                <td class="ocr-cell-items">{{ getItemNames(item.itemIds) || '-' }}</td>
                                <td class="ocr-cell-tags">{{ getTagNames(item.tagIds) || '-' }}</td>
                                <td class="ocr-cell-desc">{{ item.comment || '-' }}</td>
                                <td class="ocr-cell-actions">
                                    <v-btn size="small" color="primary" variant="tonal"
                                           :disabled="addedRowIndices.has(idx)"
                                           @click="onAddClick(item, idx)">
                                        {{ tt('Add') }}
                                    </v-btn>
                                </td>
                            </tr>
                        </tbody>
                    </v-table>
                    <div class="d-flex mt-3">
                        <v-btn color="secondary" variant="tonal" @click="reset">{{ tt('Select Another Image') }}</v-btn>
                        <v-btn color="secondary" variant="text" @click="cancel" class="ms-2">{{ tt('Cancel') }}</v-btn>
                    </div>
                </div>
            </v-card-text>

            <v-card-actions class="ocr-card-actions pt-0">
                <v-spacer />
                <template v-if="!recognizedList.length">
                    <v-btn color="primary" variant="tonal" :disabled="loading || recognizing" @click="showOpenImageDialog">
                        {{ tt('Select Image') }}
                    </v-btn>
                    <v-btn class="ms-2" :disabled="loading || recognizing || !imageFile" @click="recognize">
                        {{ tt('Recognize') }}
                        <v-progress-circular v-if="recognizing" indeterminate size="22" class="ms-2" />
                    </v-btn>
                    <v-btn color="secondary" variant="tonal" class="ms-2" :disabled="loading || recognizing" @click="cancel">
                        {{ tt('Cancel') }}
                    </v-btn>
                </template>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <snack-bar ref="snackbar" />
    <input ref="imageInput" type="file" style="display: none" :accept="SUPPORTED_IMAGE_EXTENSIONS" @change="openImage($event)" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, useTemplateRef } from 'vue';
import { useTheme } from 'vuetify';

import { useI18n } from '@/locales/helpers.ts';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTransactionItemsStore } from '@/stores/transactionItem.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';

import { TransactionType } from '@/core/transaction.ts';
import { KnownFileType } from '@/core/file.ts';
import { ThemeType } from '@/core/theme.ts';
import { SUPPORTED_IMAGE_EXTENSIONS } from '@/consts/file.ts';

import type { RecognizedReceiptImageResponse } from '@/models/large_language_model.ts';

import { compressJpgImage } from '@/lib/ui/common.ts';
import logger from '@/lib/logger.ts';

type SnackBarType = InstanceType<typeof SnackBar>;

export interface OCRBillRecognitionOpenOptions {
    onAdd?(item: RecognizedReceiptImageResponse, rowIndex: number): void | Promise<void>;
}

const theme = useTheme();
const { tt } = useI18n();
const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();
const transactionTagsStore = useTransactionTagsStore();
const transactionItemsStore = useTransactionItemsStore();
const transactionsStore = useTransactionsStore();

const snackbar = useTemplateRef<SnackBarType>('snackbar');
const imageInput = useTemplateRef<HTMLInputElement>('imageInput');

let resolveFunc: ((response: RecognizedReceiptImageResponse) => void) | null = null;
let rejectFunc: ((reason?: unknown) => void) | null = null;
let openOptions: OCRBillRecognitionOpenOptions | undefined;

const showState = ref<boolean>(false);
const loading = ref<boolean>(false);
const recognizing = ref<boolean>(false);
const imageFile = ref<File | null>(null);
const imageSrc = ref<string | undefined>(undefined);
const isDragOver = ref<boolean>(false);
const recognizedList = ref<RecognizedReceiptImageResponse[]>([]);
const addedRowIndices = ref<Set<number>>(new Set());

const isDarkMode = computed<boolean>(() => theme.global.name.value === ThemeType.Dark);

function getCategoryName(categoryId?: string): string {
    if (!categoryId) return '';
    const cat = transactionCategoriesStore.allTransactionCategoriesMap[categoryId];
    return cat?.name ?? '';
}

function getAccountName(accountId?: string): string {
    if (!accountId) return '';
    const acc = accountsStore.allAccountsMap[accountId];
    return acc?.name ?? '';
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

function loadImage(file: File): void {
    loading.value = true;
    imageFile.value = null;
    imageSrc.value = undefined;
    recognizedList.value = [];

    compressJpgImage(file, 1280, 1280, 0.8).then(blob => {
        imageFile.value = KnownFileType.JPG.createFileFromBlob(blob, 'image');
        imageSrc.value = URL.createObjectURL(blob);
        loading.value = false;
    }).catch(error => {
        imageFile.value = null;
        imageSrc.value = undefined;
        loading.value = false;
        logger.error('failed to compress image', error);
        snackbar.value?.showError('Unable to load image');
    });
}

function open(config?: OCRBillRecognitionOpenOptions): Promise<RecognizedReceiptImageResponse> {
    showState.value = true;
    loading.value = false;
    recognizing.value = false;
    imageFile.value = null;
    imageSrc.value = undefined;
    recognizedList.value = [];
    addedRowIndices.value = new Set();
    openOptions = config;

    return new Promise((resolve, reject) => {
        resolveFunc = resolve;
        rejectFunc = reject;
    });
}

function markRowAdded(rowIndex: number): void {
    addedRowIndices.value = new Set(addedRowIndices.value).add(rowIndex);
}

async function onAddClick(item: RecognizedReceiptImageResponse, rowIndex: number): Promise<void> {
    if (addedRowIndices.value.has(rowIndex)) return;
    if (openOptions?.onAdd) {
        await openOptions.onAdd(item, rowIndex);
        markRowAdded(rowIndex);
        return;
    }
    addTransaction(item);
}

function showOpenImageDialog(): void {
    if (loading.value || recognizing.value || isDragOver.value) return;
    imageInput.value?.click();
}

function openImage(event: Event): void {
    const el = event?.target as HTMLInputElement;
    if (!el?.files?.length || !el.files[0]) return;
    loadImage(el.files[0] as File);
    el.value = '';
}

function recognize(): void {
    if (loading.value || recognizing.value || !imageFile.value) return;
    recognizing.value = true;
    const file = imageFile.value;
    Promise.all([
        accountsStore.loadAllAccounts({ force: false }),
        transactionCategoriesStore.loadAllCategories({ force: false }),
        transactionTagsStore.loadAllTags({ force: false }),
        transactionItemsStore.loadAllItems({ force: false }),
        transactionsStore.recognizeReceiptImageByOCR(file)
    ]).then((results) => {
        recognizedList.value = results[4] as RecognizedReceiptImageResponse[];
        recognizing.value = false;
    }).catch(error => {
        recognizing.value = false;
        if (error?.processed) return;
        snackbar.value?.showError(error as string | { message: string } | { error: import('@/core/api').ErrorResponse });
    });
}

function addTransaction(item: RecognizedReceiptImageResponse): void {
    resolveFunc?.(item);
    showState.value = false;
    resolveFunc = null;
    rejectFunc = null;
}

function reset(): void {
    recognizedList.value = [];
    imageFile.value = null;
    imageSrc.value = undefined;
    addedRowIndices.value = new Set();
}

function cancel(): void {
    rejectFunc?.();
    showState.value = false;
    loading.value = false;
    recognizing.value = false;
    imageFile.value = null;
    imageSrc.value = undefined;
    recognizedList.value = [];
    addedRowIndices.value = new Set();
    openOptions = undefined;
    resolveFunc = null;
    rejectFunc = null;
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

function onDragEnter(): void {
    if (!loading.value && !recognizing.value) isDragOver.value = true;
}

function onDragLeave(): void {
    isDragOver.value = false;
}

function onDrop(event: DragEvent): void {
    if (loading.value || recognizing.value) return;
    isDragOver.value = false;
    if (event.dataTransfer?.files?.length && event.dataTransfer.files[0]) {
        loadImage(event.dataTransfer.files[0] as File);
    }
}

function onPaste(event: ClipboardEvent): void {
    if (!event.clipboardData) return;
    for (let i = 0; i < event.clipboardData.items.length; i++) {
        const item = event.clipboardData.items[i];
        if (item?.type.startsWith('image/')) {
            const file = item.getAsFile();
            if (file) {
                loadImage(file);
                event.preventDefault();
            }
            return;
        }
    }
}

defineExpose({ open, markRowAdded });
</script>

<style scoped>
.dropzone {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    pointer-events: none;
    border-radius: 8px;
    z-index: 10;
}
.dropzone-content {
    pointer-events: auto;
    cursor: pointer;
}
.ocr-drop-area {
    min-height: 200px;
}
.ocr-card-body {
    min-height: 280px;
}
.ocr-card-actions {
    flex-shrink: 0;
}
.ocr-recognized-table {
    table-layout: auto;
}
.ocr-recognized-table th,
.ocr-recognized-table td {
    white-space: nowrap;
}
.ocr-recognized-table .ocr-cell-type {
    min-width: 4ch;
}
.ocr-recognized-table .ocr-cell-amount {
    min-width: 6ch;
}
.ocr-recognized-table .ocr-cell-category {
    min-width: 4ch;
}
.ocr-recognized-table .ocr-cell-account {
    min-width: 4ch;
}
.ocr-recognized-table .ocr-cell-time {
    min-width: 8ch;
}
.ocr-recognized-table .ocr-cell-items {
    min-width: 4ch;
}
.ocr-recognized-table .ocr-cell-tags {
    min-width: 4ch;
}
.ocr-recognized-table .ocr-cell-desc {
    min-width: 4ch;
}
.ocr-recognized-table .ocr-cell-actions {
    width: 1%;
}
.dropzone-blurry-bg { -webkit-backdrop-filter: blur(6px); backdrop-filter: blur(6px); }
.dropzone-dragover { border: 6px dashed rgba(var(--v-border-color),var(--v-border-opacity)); }
</style>

<template>
    <v-dialog width="900" :persistent="loading || recognizing || !!imageFile" v-model="showState" @paste="onPaste">
        <v-card class="pa-sm-1 pa-md-2">
            <template #title>
                <h4 class="text-h4">{{ tt('OCR Bill Recognition') }}</h4>
            </template>

            <v-card-text class="d-flex flex-column flex-md-row flex-grow-1 overflow-y-auto" style="min-height: 320px">
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
                            <span class="pa-2">{{ tt('OCR runs locally on the server. Supports Chinese bill format (e.g. 2月7日 21:49, -100.00).') }}</span>
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
                    <v-table density="compact">
                        <thead>
                            <tr>
                                <th>{{ tt('Description') }}</th>
                                <th class="text-end">{{ tt('Amount') }}</th>
                                <th class="text-end">{{ tt('Time') }}</th>
                                <th width="100"></th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="(item, idx) in recognizedList" :key="idx">
                                <td>{{ item.comment || '-' }}</td>
                                <td class="text-end">{{ formatAmount(item.sourceAmount) }}</td>
                                <td class="text-end">{{ formatTime(item.time) }}</td>
                                <td>
                                    <v-btn size="small" color="primary" variant="tonal" @click="addTransaction(item)">
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

            <v-card-text v-if="!recognizedList.length">
                <div class="w-100 d-flex justify-center flex-wrap gap-4 align-center">
                    <v-btn color="primary" variant="tonal" :disabled="loading || recognizing" @click="showOpenImageDialog">
                        {{ tt('Select Image') }}
                    </v-btn>
                    <v-btn :disabled="loading || recognizing || !imageFile" @click="recognize">
                        {{ tt('Recognize') }}
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="recognizing"></v-progress-circular>
                    </v-btn>
                    <v-btn color="secondary" variant="tonal" :disabled="loading || recognizing" @click="cancel">{{ tt('Cancel') }}</v-btn>
                </div>
            </v-card-text>
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

import { useTransactionsStore } from '@/stores/transaction.ts';

import { KnownFileType } from '@/core/file.ts';
import { ThemeType } from '@/core/theme.ts';
import { SUPPORTED_IMAGE_EXTENSIONS } from '@/consts/file.ts';

import type { RecognizedReceiptImageResponse } from '@/models/large_language_model.ts';

import { compressJpgImage } from '@/lib/ui/common.ts';
import logger from '@/lib/logger.ts';

type SnackBarType = InstanceType<typeof SnackBar>;

const theme = useTheme();
const { tt } = useI18n();
const transactionsStore = useTransactionsStore();

const snackbar = useTemplateRef<SnackBarType>('snackbar');
const imageInput = useTemplateRef<HTMLInputElement>('imageInput');

let resolveFunc: ((response: RecognizedReceiptImageResponse) => void) | null = null;
let rejectFunc: ((reason?: unknown) => void) | null = null;

const showState = ref<boolean>(false);
const loading = ref<boolean>(false);
const recognizing = ref<boolean>(false);
const imageFile = ref<File | null>(null);
const imageSrc = ref<string | undefined>(undefined);
const isDragOver = ref<boolean>(false);
const recognizedList = ref<RecognizedReceiptImageResponse[]>([]);

const isDarkMode = computed<boolean>(() => theme.global.name.value === ThemeType.Dark);

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

function open(): Promise<RecognizedReceiptImageResponse> {
    showState.value = true;
    loading.value = false;
    recognizing.value = false;
    imageFile.value = null;
    imageSrc.value = undefined;
    recognizedList.value = [];

    return new Promise((resolve, reject) => {
        resolveFunc = resolve;
        rejectFunc = reject;
    });
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
    transactionsStore.recognizeReceiptImageByOCR(imageFile.value).then(list => {
        recognizedList.value = list;
        recognizing.value = false;
    }).catch(error => {
        recognizing.value = false;
        if (!error.processed) snackbar.value?.showError(error);
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
}

function cancel(): void {
    rejectFunc?.();
    showState.value = false;
    loading.value = false;
    recognizing.value = false;
    imageFile.value = null;
    imageSrc.value = undefined;
    recognizedList.value = [];
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

defineExpose({ open });
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
.dropzone-blurry-bg { -webkit-backdrop-filter: blur(6px); backdrop-filter: blur(6px); }
.dropzone-dragover { border: 6px dashed rgba(var(--v-border-color),var(--v-border-opacity)); }
</style>

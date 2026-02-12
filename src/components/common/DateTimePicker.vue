<template>
    <vue-date-picker ref="datetimepicker"
                     inline auto-apply
                     six-weeks="center"
                     :class="`datetime-picker ${showAlternateDates && alternateCalendarType ? 'datetime-picker-with-alternate-date' : ''} ${datetimePickerClass}`"
                     :config="{ noSwipe: !!noSwipeAndScroll, monthChangeOnScroll: !noSwipeAndScroll }"
                     :time-config="{ enableTimePicker: enableTimePicker, enableSeconds: true, is24: is24Hour }"
                     :input-attrs="{ clearable: !!clearable }"
                     :dark="isDarkMode"
                     :vertical="vertical"
                     :disable-year-select="disableYearSelect"
                     :year-range="yearRange"
                     :day-names="dayNames"
                     :week-start="firstDayOfWeek"
                     :year-first="isYearFirst"
                     :min-date="minDate"
                     :max-date="maxDate"
                     :disabled-dates="disabledDates"
                     :range="isDateRange ? { partialRange: false } : undefined"
                     :preset-dates="presetRanges"
                     :model-value="internalValue"
                     @update:model-value="onPickerUpdate"
                     @select="onPickerSelect">
        <template #year="{ value }">
            {{ getDisplayYear(value) }}
        </template>
        <template #year-overlay-value="{ value }">
            {{ getDisplayYear(value) }}
        </template>
        <template #month="{ value }">
            {{ getDisplayMonth(value) }}
        </template>
        <template #month-overlay-value="{ value }">
            {{ getDisplayMonth(value) }}
        </template>
        <template #day="{ date }">
            <div class="datetime-picker-display-dates">
                <span>{{ getDisplayDay(date) }}</span>
                <span class="datetime-picker-alternate-date" v-if="showAlternateDates && alternateCalendarType && getAlternateDate(date)">{{ getAlternateDate(date) }}</span>
            </div>
        </template>
        <template #am-pm-button="{ toggle, value }">
            <button class="dp__pm_am_button" tabindex="0" @click="toggle">{{ tt(`datetime.${value}.content`) }}</button>
        </template>
    </vue-date-picker>
</template>

<script setup lang="ts">
import { computed, ref, useTemplateRef, watch } from 'vue';
import { type MenuView, VueDatePicker } from '@vuepic/vue-datepicker';

import { useI18n } from '@/locales/helpers.ts';

import { useUserStore } from '@/stores/user.ts';

import type { CalendarType } from '@/core/calendar.ts';
import { NumeralSystem } from '@/core/numeral.ts';
import type { PresetDateRange, WeekDayValue } from '@/core/datetime.ts';
import { isDefined, isArray, arrangeArrayWithNewStartIndex } from '@/lib/common.ts';
import { getAllowedYearRange, getYearMonthDayDateTime } from '@/lib/datetime.ts';

type VueDatePickerType = InstanceType<typeof VueDatePicker>;
type SupportedModelValue = Date | Date[] | null;

const props = defineProps<{
    modelValue: SupportedModelValue;
    datetimePickerClass?: string;
    isDarkMode: boolean;
    numeralSystem?: number;
    enableTimePicker: boolean;
    disableYearSelect?: boolean;
    vertical?: boolean;
    noSwipeAndScroll?: boolean;
    clearable?: boolean;
    minDate?: Date;
    maxDate?: Date;
    disabledDates?: (date: Date) => boolean;
    showAlternateDates?: boolean;
    presetRanges?: PresetDateRange[];
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: SupportedModelValue): void;
}>();

const {
    tt,
    getAllMinWeekdayNames,
    getCurrentCalendarDisplayType,
    getCurrentNumeralSystemType,
    isLongDateMonthAfterYear,
    isLongTime24HourFormat,
    getCalendarDisplayShortYearFromDateTime,
    getCalendarDisplayShortMonthFromDateTime,
    getCalendarDisplayDayOfMonthFromDateTime,
    getCalendarAlternateDate
} = useI18n();

const userStore = useUserStore();

const datetimepicker = useTemplateRef<VueDatePickerType>('datetimepicker');

const yearRange = getAllowedYearRange();

const dayNames = computed<string[]>(() => arrangeArrayWithNewStartIndex(getAllMinWeekdayNames(), firstDayOfWeek.value));
const firstDayOfWeek = computed<WeekDayValue>(() => userStore.currentUserFirstDayOfWeek);
const isYearFirst = computed<boolean>(() => isLongDateMonthAfterYear());
const is24Hour = computed<boolean>(() => isLongTime24HourFormat());
const alternateCalendarType = computed<CalendarType | undefined>(() => getCurrentCalendarDisplayType().secondaryCalendarType);

const actualNumeralSystem = computed<NumeralSystem>(() => {
    if (isDefined(props.numeralSystem)) {
        return NumeralSystem.valueOf(props.numeralSystem) ?? NumeralSystem.Default;
    } else {
        return getCurrentNumeralSystemType();
    }
});

const internalValue = ref<SupportedModelValue>(
    props.modelValue instanceof Date ? new Date(props.modelValue.getTime()) : props.modelValue
);

watch(() => props.modelValue, (v) => {
    internalValue.value = v instanceof Date ? new Date(v.getTime()) : v;
});

function onPickerUpdate(value: SupportedModelValue): void {
    internalValue.value = value instanceof Date ? new Date(value.getTime()) : value;
    const same = props.modelValue instanceof Date && internalValue.value instanceof Date
        ? props.modelValue.getTime() === internalValue.value.getTime()
        : props.modelValue === internalValue.value;
    if (!same) {
        emit('update:modelValue', internalValue.value);
    }
}

function onPickerSelect(value: SupportedModelValue): void {
    if (value instanceof Date) {
        internalValue.value = new Date(value.getTime());
        const same = props.modelValue instanceof Date && internalValue.value instanceof Date
            ? props.modelValue.getTime() === internalValue.value.getTime()
            : props.modelValue === internalValue.value;
        if (!same) {
            emit('update:modelValue', internalValue.value);
        }
    }
}

const dateTime = internalValue;

const isDateRange = computed<boolean>(() => isArray(internalValue.value));

function getAlternateDate(date: Date): string | undefined {
    if (!props.showAlternateDates) {
        return undefined;
    }

    return getCalendarAlternateDate({
        year: date.getFullYear(),
        month: date.getMonth() + 1,
        day: date.getDate()
    })?.displayDate;
};

function switchView(viewType: MenuView): void {
    datetimepicker.value?.switchView(viewType);
}

function getDisplayYear(year: number): string {
    return getCalendarDisplayShortYearFromDateTime(getYearMonthDayDateTime(year, 1, 1), actualNumeralSystem.value);
}

function getDisplayMonth(month: number): string {
    if (isArray(dateTime.value)) {
        return getCalendarDisplayShortMonthFromDateTime(getYearMonthDayDateTime(dateTime.value[0]!.getFullYear(), month + 1, 1), actualNumeralSystem.value);
    } else if (dateTime.value) {
        return getCalendarDisplayShortMonthFromDateTime(getYearMonthDayDateTime(dateTime.value.getFullYear(), month + 1, 1), actualNumeralSystem.value);
    } else {
        return getCalendarDisplayShortMonthFromDateTime(getYearMonthDayDateTime(new Date().getFullYear(), month + 1, 1), actualNumeralSystem.value);
    }
}

function getDisplayDay(date: Date): string {
    return getCalendarDisplayDayOfMonthFromDateTime(getYearMonthDayDateTime(date.getFullYear(), date.getMonth() + 1, date.getDate()), actualNumeralSystem.value);
}

function getModelValue(): SupportedModelValue {
    if (datetimepicker.value) {
        const picker = datetimepicker.value as unknown as { 
            modelValue?: SupportedModelValue; 
            selectedDate?: Date | Date[] | null;
            $el?: HTMLElement;
        };
        if (picker.modelValue instanceof Date) {
            return picker.modelValue;
        }
        if (picker.selectedDate instanceof Date) {
            return picker.selectedDate;
        }
        if (picker.$el) {
            const selectedCell = picker.$el.querySelector('.dp__cell_selected, .dp__active_date');
            if (selectedCell) {
                const dateAttr = selectedCell.getAttribute('data-date');
                if (dateAttr) {
                    const date = new Date(dateAttr);
                    if (!isNaN(date.getTime())) {
                        return date;
                    }
                }
            }
        }
    }
    return internalValue.value;
}

defineExpose({
    switchView,
    getModelValue
});
</script>

<style>
.datetime-picker-display-dates {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
}

.datetime-picker-alternate-date {
    margin-top: -2px;
    opacity: 0.5;
    font-size: 0.8rem;
}

.dp__cell_disabled .datetime-picker-alternate-date,
.dp__cell_offset .datetime-picker-alternate-date {
    opacity: 0.8;
}

.dp__main.datetime-picker .dp__calendar .dp__calendar_row > .dp__calendar_item .datetime-picker-display-dates > span.datetime-picker-alternate-date {
    display: block;
    width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.dp__main.datetime-picker.datetime-picker-with-alternate-date .dp__calendar .dp__calendar_row {
    --dp-cell-size: 45px;
}
</style>

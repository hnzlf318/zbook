import type { CurrencyInfo } from '@/core/currency.ts';

// 仅支持人民币
export const ALL_CURRENCIES: Record<string, CurrencyInfo> = {
    'CNY': {
        code: 'CNY',
        fraction: 2,
        symbol: {
            normal: '¥'
        },
        unit: 'Yuan'
    },
};

export const DEFAULT_CURRENCY_SYMBOL: string = '¤';
export const DEFAULT_CURRENCY_CODE: string = 'CNY';
export const PARENT_ACCOUNT_CURRENCY_PLACEHOLDER: string = '---';

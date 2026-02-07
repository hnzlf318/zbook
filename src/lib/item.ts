import { reversed } from '@/core/base.ts';
import { TransactionItem } from '@/models/transaction_item.ts';

export function isNoAvailableItem(items: TransactionItem[], showHidden: boolean): boolean {
    for (const item of items) {
        if (showHidden || !item.hidden) {
            return false;
        }
    }

    return true;
}

export function getAvailableItemCount(items: TransactionItem[], showHidden: boolean): number {
    let count = 0;

    for (const item of items) {
        if (showHidden || !item.hidden) {
            count++;
        }
    }

    return count;
}

export function getFirstShowingId(items: TransactionItem[], showHidden: boolean): string | null {
    for (const item of items) {
        if (showHidden || !item.hidden) {
            return item.id;
        }
    }

    return null;
}

export function getLastShowingId(items: TransactionItem[], showHidden: boolean): string | null {
    for (const item of reversed(items)) {
        if (showHidden || !item.hidden) {
            return item.id;
        }
    }

    return null;
}

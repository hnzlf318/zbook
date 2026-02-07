export class TransactionItem implements TransactionItemInfoResponse {
    public id: string;
    public name: string;
    public groupId: string;
    public displayOrder: number;
    public hidden: boolean;

    private constructor(id: string, name: string, groupId: string, displayOrder: number, hidden: boolean) {
        this.id = id;
        this.name = name;
        this.groupId = groupId;
        this.displayOrder = displayOrder;
        this.hidden = hidden;
    }

    public toCreateRequest(): TransactionItemCreateRequest {
        return {
            name: this.name,
            groupId: this.groupId
        };
    }

    public toModifyRequest(): TransactionItemModifyRequest {
        return {
            id: this.id,
            groupId: this.groupId,
            name: this.name
        };
    }

    public clone(): TransactionItem {
        return new TransactionItem(this.id, this.name, this.groupId, this.displayOrder, this.hidden);
    }

    public static of(itemResponse: TransactionItemInfoResponse): TransactionItem {
        return new TransactionItem(itemResponse.id, itemResponse.name, itemResponse.groupId, itemResponse.displayOrder, itemResponse.hidden);
    }

    public static ofMulti(itemResponses: TransactionItemInfoResponse[]): TransactionItem[] {
        const items: TransactionItem[] = [];

        for (const itemResponse of itemResponses) {
            items.push(TransactionItem.of(itemResponse));
        }

        return items;
    }

    public static createNewItem(name?: string, groupId?: string): TransactionItem {
        return new TransactionItem('', name || '', groupId || '0', 0, false);
    }
}

export interface TransactionItemCreateRequest {
    readonly groupId: string;
    readonly name: string;
}

export interface TransactionItemCreateBatchRequest {
    readonly items: TransactionItemCreateRequest[];
    readonly groupId: string;
    readonly skipExists: boolean;
}

export interface TransactionItemModifyRequest {
    readonly id: string;
    readonly groupId: string;
    readonly name: string;
}

export interface TransactionItemHideRequest {
    readonly id: string;
    readonly hidden: boolean;
}

export interface TransactionItemMoveRequest {
    readonly newDisplayOrders: TransactionItemNewDisplayOrderRequest[];
}

export interface TransactionItemNewDisplayOrderRequest {
    readonly id: string;
    readonly displayOrder: number;
}

export interface TransactionItemDeleteRequest {
    readonly id: string;
}

export interface TransactionItemInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly groupId: string;
    readonly displayOrder: number;
    readonly hidden: boolean;
}

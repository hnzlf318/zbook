export class TransactionItemGroup implements TransactionItemGroupInfoResponse {
    public id: string;
    public name: string;
    public displayOrder: number;

    private constructor(id: string, name: string, displayOrder: number) {
        this.id = id;
        this.name = name;
        this.displayOrder = displayOrder;
    }

    public toCreateRequest(): TransactionItemGroupCreateRequest {
        return {
            name: this.name
        };
    }

    public toModifyRequest(): TransactionItemGroupModifyRequest {
        return {
            id: this.id,
            name: this.name
        };
    }

    public clone(): TransactionItemGroup {
        return new TransactionItemGroup(this.id, this.name, this.displayOrder);
    }

    public static of(itemGroupResponse: TransactionItemGroupInfoResponse): TransactionItemGroup {
        return new TransactionItemGroup(itemGroupResponse.id, itemGroupResponse.name, itemGroupResponse.displayOrder);
    }

    public static ofMulti(itemGroupResponses: TransactionItemGroupInfoResponse[]): TransactionItemGroup[] {
        const itemGroups: TransactionItemGroup[] = [];

        for (const itemGroupResponse of itemGroupResponses) {
            itemGroups.push(TransactionItemGroup.of(itemGroupResponse));
        }

        return itemGroups;
    }

    public static createNewItemGroup(name?: string): TransactionItemGroup {
        return new TransactionItemGroup('', name || '', 0);
    }
}

export interface TransactionItemGroupCreateRequest {
    readonly name: string;
}

export interface TransactionItemGroupModifyRequest {
    readonly id: string;
    readonly name: string;
}

export interface TransactionItemGroupMoveRequest {
    readonly newDisplayOrders: TransactionItemGroupNewDisplayOrderRequest[];
}

export interface TransactionItemGroupNewDisplayOrderRequest {
    readonly id: string;
    readonly displayOrder: number;
}

export interface TransactionItemGroupDeleteRequest {
    readonly id: string;
}

export interface TransactionItemGroupInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly displayOrder: number;
}

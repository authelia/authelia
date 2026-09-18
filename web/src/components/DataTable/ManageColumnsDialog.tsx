import { type RowData, type Table } from "@tanstack/react-table";
import { ChevronDown, ChevronUp } from "lucide-react";
import { useTranslation } from "react-i18next";

import { type DataTableFeatures } from "@components/DataTable/features";
import { Button } from "@components/UI/Button";
import { Checkbox } from "@components/UI/Checkbox";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/UI/Dialog";

interface ManageColumnsDialogProps<T extends RowData> {
    columnOrder: string[];
    id: string;
    labels: Record<string, string>;
    onOpenChange: (open: boolean) => void;
    open: boolean;
    table: Table<DataTableFeatures, T>;
}

function move(order: string[], index: number, direction: -1 | 1): string[] {
    const next = [...order];
    const swapWith = index + direction;

    if (swapWith < 0 || swapWith >= next.length) {
        return next;
    }

    [next[index], next[swapWith]] = [next[swapWith], next[index]];

    return next;
}

function ManageColumnsDialog<T extends RowData>({
    columnOrder,
    id,
    labels,
    onOpenChange,
    open,
    table,
}: Readonly<ManageColumnsDialogProps<T>>) {
    const { t: translate } = useTranslation("settings");
    const order = columnOrder;

    return (
        <Dialog onOpenChange={onOpenChange} open={open}>
            <DialogContent id={id}>
                <DialogHeader>
                    <DialogTitle>{translate("Manage Columns")}</DialogTitle>
                </DialogHeader>
                <ul className="flex flex-col gap-1">
                    {order.map((field, index) => {
                        const column = table.getColumn(field);

                        if (!column) {
                            return null;
                        }

                        return (
                            <li className="flex items-center gap-2 rounded-md border px-2 py-1.5" key={field}>
                                <Checkbox
                                    checked={column.getIsVisible()}
                                    disabled={!column.getCanHide()}
                                    id={`${id}-${field}-visible`}
                                    onCheckedChange={(checked) => column.toggleVisibility(checked === true)}
                                />
                                <label className="flex-1 text-sm" htmlFor={`${id}-${field}-visible`}>
                                    {labels[field] ?? field}
                                </label>
                                <Button
                                    aria-label={translate("Move Up")}
                                    disabled={index === 0}
                                    onClick={() => table.setColumnOrder(move(order, index, -1))}
                                    size="icon-sm"
                                    variant="ghost"
                                >
                                    <ChevronUp />
                                </Button>
                                <Button
                                    aria-label={translate("Move Down")}
                                    disabled={index === order.length - 1}
                                    onClick={() => table.setColumnOrder(move(order, index, 1))}
                                    size="icon-sm"
                                    variant="ghost"
                                >
                                    <ChevronDown />
                                </Button>
                            </li>
                        );
                    })}
                </ul>
            </DialogContent>
        </Dialog>
    );
}

export { ManageColumnsDialog };

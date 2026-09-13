import { type KeyboardEvent, useState } from "react";

import { Combobox as ComboboxPrimitive } from "@base-ui/react/combobox";
import { XIcon } from "lucide-react";

import { Badge } from "@components/UI/Badge";
import { cn } from "@utils/Styles";

interface ComboboxProps {
    id?: string;
    value: string[];
    onChange: (value: string[]) => void;
    options: string[];
    freeSolo?: boolean;
    placeholder?: string;
    disabled?: boolean;
    error?: boolean;
    "aria-invalid"?: boolean;
    className?: string;
}

function Combobox({
    "aria-invalid": ariaInvalid,
    className,
    disabled,
    error,
    freeSolo,
    id,
    onChange,
    options,
    placeholder,
    value,
}: ComboboxProps) {
    const [inputValue, setInputValue] = useState("");

    function addValue(candidate: string) {
        const trimmed = candidate.trim();

        if (trimmed.length === 0) {
            return;
        }

        if (!value.includes(trimmed)) {
            onChange([...value, trimmed]);
        }

        setInputValue("");
    }

    function removeValue(candidate: string) {
        onChange(value.filter((item) => item !== candidate));
    }

    function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
        if (!freeSolo) {
            return;
        }

        if (event.key === "Enter" || event.key === ",") {
            if (inputValue.trim().length > 0) {
                event.preventDefault();
                addValue(inputValue);
            }
        } else if (event.key === "Backspace" && inputValue.length === 0 && value.length > 0) {
            removeValue(value[value.length - 1]);
        }
    }

    return (
        <ComboboxPrimitive.Root
            disabled={disabled}
            inputValue={inputValue}
            items={options}
            multiple
            onInputValueChange={setInputValue}
            onValueChange={(next) => onChange(next)}
            value={value}
        >
            <div
                data-slot="combobox"
                className={cn(
                    "flex min-h-10 w-full flex-wrap items-center gap-1.5 rounded-md border border-input bg-transparent px-2 py-1.5 shadow-xs transition-[color,box-shadow] has-[input:focus-visible]:border-ring has-[input:focus-visible]:ring-[3px] has-[input:focus-visible]:ring-ring/50",
                    (error || ariaInvalid) && "border-destructive ring-destructive/20 dark:ring-destructive/40",
                    disabled && "pointer-events-none cursor-not-allowed opacity-50",
                    className,
                )}
            >
                {value.map((item) => (
                    <Badge data-slot="combobox-badge" key={item} variant="secondary">
                        {item}
                        <button
                            aria-label={`Remove ${item}`}
                            data-slot="combobox-badge-remove"
                            disabled={disabled}
                            onClick={() => removeValue(item)}
                            type="button"
                        >
                            <XIcon className="size-3" />
                        </button>
                    </Badge>
                ))}
                <ComboboxPrimitive.Input
                    aria-invalid={ariaInvalid || error || undefined}
                    className="h-7 min-w-20 flex-1 border-0 bg-transparent text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed"
                    data-slot="combobox-input"
                    disabled={disabled}
                    id={id}
                    onKeyDown={handleKeyDown}
                    placeholder={value.length === 0 ? placeholder : undefined}
                    role="combobox"
                />
            </div>
            <ComboboxPrimitive.Portal>
                <ComboboxPrimitive.Positioner align="start" className="z-50" sideOffset={4}>
                    <ComboboxPrimitive.Popup
                        data-slot="combobox-popup"
                        className={cn(
                            "z-50 max-h-(--available-height) min-w-[8rem] overflow-x-hidden overflow-y-auto rounded-md border bg-popover p-1 text-popover-foreground shadow-md outline-none",
                        )}
                    >
                        <ComboboxPrimitive.Empty
                            className="px-2 py-1.5 text-sm text-muted-foreground"
                            data-slot="combobox-empty"
                        >
                            No options
                        </ComboboxPrimitive.Empty>
                        <ComboboxPrimitive.List data-slot="combobox-list">
                            {(item: string) => (
                                <ComboboxPrimitive.Item
                                    className={cn(
                                        "relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none",
                                        "data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground",
                                        "data-[selected]:font-medium",
                                    )}
                                    data-slot="combobox-item"
                                    key={item}
                                    value={item}
                                >
                                    {item}
                                </ComboboxPrimitive.Item>
                            )}
                        </ComboboxPrimitive.List>
                    </ComboboxPrimitive.Popup>
                </ComboboxPrimitive.Positioner>
            </ComboboxPrimitive.Portal>
        </ComboboxPrimitive.Root>
    );
}

export { Combobox };

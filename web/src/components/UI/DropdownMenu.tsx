import { type ComponentProps } from "react";

import { Menu as MenuPrimitive } from "@base-ui/react/menu";
import { CheckIcon, ChevronRightIcon, CircleIcon } from "lucide-react";

import { cn } from "@utils/Styles";

const popupClasses =
    "z-50 min-w-[8rem] overflow-x-hidden overflow-y-auto rounded-md border bg-popover p-1 text-popover-foreground shadow-md outline-none " +
    "origin-(--transform-origin) transition-[opacity,transform] duration-150 " +
    "data-[starting-style]:scale-95 data-[starting-style]:opacity-0 " +
    "data-[ending-style]:scale-95 data-[ending-style]:opacity-0";

const itemClasses =
    "relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none " +
    "data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 " +
    "[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 [&_svg:not([class*='text-'])]:text-muted-foreground";

function DropdownMenu({ ...props }: ComponentProps<typeof MenuPrimitive.Root>) {
    return <MenuPrimitive.Root {...props} />;
}

function DropdownMenuPortal({ ...props }: ComponentProps<typeof MenuPrimitive.Portal>) {
    return <MenuPrimitive.Portal {...props} />;
}

function DropdownMenuTrigger({ ...props }: ComponentProps<typeof MenuPrimitive.Trigger>) {
    return <MenuPrimitive.Trigger data-slot="dropdown-menu-trigger" {...props} />;
}

function DropdownMenuContent({
    align,
    className,
    side,
    sideOffset = 4,
    ...props
}: ComponentProps<typeof MenuPrimitive.Popup> & {
    align?: ComponentProps<typeof MenuPrimitive.Positioner>["align"];
    side?: ComponentProps<typeof MenuPrimitive.Positioner>["side"];
    sideOffset?: number;
}) {
    return (
        <MenuPrimitive.Portal>
            <MenuPrimitive.Positioner align={align} side={side} sideOffset={sideOffset} className="z-50">
                <MenuPrimitive.Popup
                    data-slot="dropdown-menu-content"
                    className={cn(popupClasses, "max-h-(--available-height)", className)}
                    {...props}
                />
            </MenuPrimitive.Positioner>
        </MenuPrimitive.Portal>
    );
}

function DropdownMenuGroup({ ...props }: ComponentProps<typeof MenuPrimitive.Group>) {
    return <MenuPrimitive.Group data-slot="dropdown-menu-group" {...props} />;
}

function DropdownMenuItem({
    className,
    inset,
    variant = "default",
    ...props
}: ComponentProps<typeof MenuPrimitive.Item> & {
    inset?: boolean;
    variant?: "default" | "destructive";
}) {
    return (
        <MenuPrimitive.Item
            data-slot="dropdown-menu-item"
            data-inset={inset}
            data-variant={variant}
            className={cn(
                itemClasses,
                "data-[inset]:pl-8 data-[variant=destructive]:text-destructive data-[variant=destructive]:data-[highlighted]:bg-destructive/10 data-[variant=destructive]:data-[highlighted]:text-destructive dark:data-[variant=destructive]:data-[highlighted]:bg-destructive/20 data-[variant=destructive]:*:[svg]:text-destructive!",
                className,
            )}
            {...props}
        />
    );
}

function DropdownMenuCheckboxItem({
    checked,
    children,
    className,
    ...props
}: ComponentProps<typeof MenuPrimitive.CheckboxItem>) {
    return (
        <MenuPrimitive.CheckboxItem
            data-slot="dropdown-menu-checkbox-item"
            className={cn(itemClasses, "py-1.5 pr-2 pl-8", className)}
            checked={checked}
            {...props}
        >
            <span className="pointer-events-none absolute left-2 flex size-3.5 items-center justify-center">
                <MenuPrimitive.CheckboxItemIndicator>
                    <CheckIcon className="size-4" />
                </MenuPrimitive.CheckboxItemIndicator>
            </span>
            {children}
        </MenuPrimitive.CheckboxItem>
    );
}

function DropdownMenuRadioGroup({ ...props }: ComponentProps<typeof MenuPrimitive.RadioGroup>) {
    return <MenuPrimitive.RadioGroup data-slot="dropdown-menu-radio-group" {...props} />;
}

function DropdownMenuRadioItem({ children, className, ...props }: ComponentProps<typeof MenuPrimitive.RadioItem>) {
    return (
        <MenuPrimitive.RadioItem
            data-slot="dropdown-menu-radio-item"
            className={cn(itemClasses, "py-1.5 pr-2 pl-8", className)}
            {...props}
        >
            <span className="pointer-events-none absolute left-2 flex size-3.5 items-center justify-center">
                <MenuPrimitive.RadioItemIndicator>
                    <CircleIcon className="size-2 fill-current" />
                </MenuPrimitive.RadioItemIndicator>
            </span>
            {children}
        </MenuPrimitive.RadioItem>
    );
}

function DropdownMenuLabel({
    className,
    inset,
    ...props
}: ComponentProps<typeof MenuPrimitive.GroupLabel> & {
    inset?: boolean;
}) {
    return (
        <MenuPrimitive.GroupLabel
            data-slot="dropdown-menu-label"
            data-inset={inset}
            className={cn("px-2 py-1.5 text-sm font-medium data-[inset]:pl-8", className)}
            {...props}
        />
    );
}

function DropdownMenuSeparator({ className, ...props }: ComponentProps<"div">) {
    return (
        <div
            data-slot="dropdown-menu-separator"
            role="separator"
            className={cn("-mx-1 my-1 h-px bg-border", className)}
            {...props}
        />
    );
}

function DropdownMenuShortcut({ className, ...props }: ComponentProps<"span">) {
    return (
        <span
            data-slot="dropdown-menu-shortcut"
            className={cn("ml-auto text-xs tracking-widest text-muted-foreground", className)}
            {...props}
        />
    );
}

function DropdownMenuSub({ ...props }: ComponentProps<typeof MenuPrimitive.SubmenuRoot>) {
    return <MenuPrimitive.SubmenuRoot {...props} />;
}

function DropdownMenuSubTrigger({
    children,
    className,
    inset,
    ...props
}: ComponentProps<typeof MenuPrimitive.SubmenuTrigger> & {
    inset?: boolean;
}) {
    return (
        <MenuPrimitive.SubmenuTrigger
            data-slot="dropdown-menu-sub-trigger"
            data-inset={inset}
            className={cn(itemClasses, "data-[inset]:pl-8 data-[popup-open]:bg-accent", className)}
            {...props}
        >
            {children}
            <ChevronRightIcon className="ml-auto size-4" />
        </MenuPrimitive.SubmenuTrigger>
    );
}

function DropdownMenuSubContent({ className, ...props }: ComponentProps<typeof MenuPrimitive.Popup>) {
    return (
        <MenuPrimitive.Portal>
            <MenuPrimitive.Positioner className="z-50">
                <MenuPrimitive.Popup
                    data-slot="dropdown-menu-sub-content"
                    className={cn(popupClasses, "shadow-lg", className)}
                    {...props}
                />
            </MenuPrimitive.Positioner>
        </MenuPrimitive.Portal>
    );
}

export {
    DropdownMenu,
    DropdownMenuPortal,
    DropdownMenuTrigger,
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuLabel,
    DropdownMenuItem,
    DropdownMenuCheckboxItem,
    DropdownMenuRadioGroup,
    DropdownMenuRadioItem,
    DropdownMenuSeparator,
    DropdownMenuShortcut,
    DropdownMenuSub,
    DropdownMenuSubTrigger,
    DropdownMenuSubContent,
};

"use client";

import { type ComponentProps, type ReactNode } from "react";

import { Toast as ToastPrimitive } from "@base-ui/react/toast";
import { CircleCheckIcon, InfoIcon, OctagonXIcon, TriangleAlertIcon } from "lucide-react";

import { cn } from "@utils/Styles";

const icons: Record<string, ReactNode> = {
    error: <OctagonXIcon className="size-4" />,
    info: <InfoIcon className="size-4" />,
    success: <CircleCheckIcon className="size-4" />,
    warning: <TriangleAlertIcon className="size-4" />,
};

const accent = cn(
    "[--toast-accent:var(--muted-foreground)]",
    "data-[type=success]:[--toast-accent:var(--success)]",
    "data-[type=info]:[--toast-accent:var(--info)]",
    "data-[type=warning]:[--toast-accent:var(--warning)]",
    "data-[type=error]:[--toast-accent:var(--destructive)]",
    "[--toast-ink:light-dark(oklch(from_var(--toast-accent)_0.52_c_h),oklch(from_var(--toast-accent)_0.84_c_h))]",
    "[--toast-surface:light-dark(oklch(from_var(--toast-accent)_0.97_0.03_h),oklch(from_var(--toast-accent)_0.24_0.05_h))]",
    "[--toast-edge:light-dark(oklch(from_var(--toast-accent)_0.90_0.09_h),oklch(from_var(--toast-accent)_0.36_0.08_h))]",
);

function ToastProvider({ ...props }: ComponentProps<typeof ToastPrimitive.Provider>) {
    return <ToastPrimitive.Provider {...props} />;
}

function Toaster({ className, ...props }: ComponentProps<typeof ToastPrimitive.Viewport>) {
    const { toasts } = ToastPrimitive.useToastManager();

    return (
        <ToastPrimitive.Portal>
            <ToastPrimitive.Viewport
                data-slot="toast-viewport"
                className={cn(
                    "fixed top-4 right-4 z-50 flex w-[356px] max-w-[calc(100vw-2rem)] flex-col gap-2",
                    className,
                )}
                {...props}
            >
                {toasts.map((toast) => (
                    <ToastPrimitive.Root
                        key={toast.id}
                        toast={toast}
                        data-slot="toast"
                        className={cn(
                            "notification flex items-center gap-1.5 rounded-(--radius) border p-4",
                            "font-sans text-[13px] text-(--toast-ink)",
                            "border-(--toast-edge) bg-(--toast-surface) shadow-[0_4px_12px_rgb(0_0_0/0.1)]",
                            accent,
                            "transition-[opacity,transform] duration-200",
                            "data-[starting-style]:translate-x-full data-[starting-style]:opacity-0",
                            "data-[ending-style]:translate-x-full data-[ending-style]:opacity-0",
                        )}
                    >
                        {toast.type ? <span className="flex shrink-0 items-center">{icons[toast.type]}</span> : null}
                        <ToastPrimitive.Title data-slot="toast-title" className="grow leading-normal font-medium" />
                    </ToastPrimitive.Root>
                ))}
            </ToastPrimitive.Viewport>
        </ToastPrimitive.Portal>
    );
}

export { ToastProvider, Toaster };

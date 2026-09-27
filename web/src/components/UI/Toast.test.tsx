import { useEffect, useRef } from "react";

import { Toast as ToastPrimitive } from "@base-ui/react/toast";
import { render, screen, waitFor } from "@testing-library/react";

import { ToastProvider, Toaster } from "@components/UI/Toast";

function Queue({ count }: { count: number }) {
    const manager = ToastPrimitive.useToastManager();
    const queuedRef = useRef(false);

    useEffect(() => {
        if (queuedRef.current) return;
        queuedRef.current = true;
        for (let n = 1; n <= count; n++) {
            manager.add({ timeout: 60000, title: `notification ${n}`, type: "info" });
        }
    }, [manager, count]);

    return <Toaster />;
}

const roots = () => Array.from(document.querySelectorAll('[data-slot="toast"]')) as HTMLElement[];

it("keeps toasts past the provider limit out of view", async () => {
    render(
        <ToastProvider limit={3}>
            <Queue count={4} />
        </ToastProvider>,
    );

    await waitFor(() => expect(screen.getByText("notification 4")).toBeInTheDocument());

    // Base UI mounts every toast and marks the overflow with data-limited, relying on CSS to hide it.
    expect(roots()).toHaveLength(4);
    expect(roots().filter((el) => !el.hasAttribute("data-limited"))).toHaveLength(3);
    expect(roots().filter((el) => el.hasAttribute("data-limited"))).toHaveLength(1);
    expect(roots().every((el) => el.className.includes("data-[limited]:hidden"))).toBe(true);
});

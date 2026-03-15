"use client";

import { type ComponentProps } from "react";

import { Collapsible as CollapsiblePrimitive } from "radix-ui";

function Collapsible({ ...props }: ComponentProps<typeof CollapsiblePrimitive.Root>) {
    return <CollapsiblePrimitive.Root data-slot="collapsible" {...props} />;
}

function CollapsibleTrigger({ ...props }: ComponentProps<typeof CollapsiblePrimitive.CollapsibleTrigger>) {
    return <CollapsiblePrimitive.CollapsibleTrigger data-slot="collapsible-trigger" {...props} />;
}

function CollapsibleContent({ ...props }: ComponentProps<typeof CollapsiblePrimitive.CollapsibleContent>) {
    return <CollapsiblePrimitive.CollapsibleContent data-slot="collapsible-content" {...props} />;
}

export { Collapsible, CollapsibleTrigger, CollapsibleContent };

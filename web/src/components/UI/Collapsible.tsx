"use client";

import { type ComponentProps } from "react";

import { Collapsible as CollapsiblePrimitive } from "@base-ui/react/collapsible";

function Collapsible({ ...props }: ComponentProps<typeof CollapsiblePrimitive.Root>) {
    return <CollapsiblePrimitive.Root data-slot="collapsible" {...props} />;
}

function CollapsibleTrigger({ ...props }: ComponentProps<typeof CollapsiblePrimitive.Trigger>) {
    return <CollapsiblePrimitive.Trigger data-slot="collapsible-trigger" {...props} />;
}

function CollapsibleContent({ ...props }: ComponentProps<typeof CollapsiblePrimitive.Panel>) {
    return <CollapsiblePrimitive.Panel data-slot="collapsible-content" {...props} />;
}

export { Collapsible, CollapsibleTrigger, CollapsibleContent };

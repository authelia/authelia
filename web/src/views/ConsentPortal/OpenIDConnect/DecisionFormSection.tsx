import { ReactNode } from "react";

import { FieldDescription } from "@components/UI/Field";
import { Item, ItemActions, ItemContent, ItemGroup, ItemMedia } from "@components/UI/Item";
import { cn } from "@utils/Styles";

export interface Props {
    children: ReactNode;
    className?: string;
    description?: string;
    id?: string;
    title: string;
}

function DecisionFormSection({ children, className, description, id, title }: Props) {
    return (
        <section id={id} className={cn("w-full text-left", className)}>
            <h3 className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">{title}</h3>
            {description ? <FieldDescription className="mt-0.5 text-xs">{description}</FieldDescription> : null}
            <ItemGroup className="mt-2 gap-1">{children}</ItemGroup>
        </section>
    );
}

export interface ItemProps {
    actions?: ReactNode;
    children: ReactNode;
    className?: string;
    icon?: ReactNode;
    id?: string;
    identifier?: null | string;
}

export function DecisionFormSectionItem({ actions, children, className, icon, id, identifier }: ItemProps) {
    return (
        <Item role="listitem" id={id} size={"sm"} className={cn("gap-3 px-2 py-1.5", className)}>
            {icon ? <ItemMedia className="text-muted-foreground">{icon}</ItemMedia> : null}
            <ItemContent className="min-w-0 flex-1">{children}</ItemContent>
            {identifier || actions ? (
                <ItemActions className="gap-2">
                    {identifier ? <span className="font-mono text-xs text-muted-foreground">{identifier}</span> : null}
                    {actions}
                </ItemActions>
            ) : null}
        </Item>
    );
}

export default DecisionFormSection;

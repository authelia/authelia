import { ReactNode, useState } from "react";

import { ChevronDown, CircleCheck, CircleSlash, TriangleAlert } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router";

import HomeButton from "@components/HomeButton";
import { Card } from "@components/UI/Card";
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia } from "@components/UI/Empty";
import { Item, ItemContent, ItemDescription, ItemGroup, ItemTitle } from "@components/UI/Item";
import { Separator } from "@components/UI/Separator";
import {
    Decision,
    DecisionAccepted,
    ErrorDebug,
    ErrorDescription,
    ErrorHint,
    Error as ErrorParam,
    ErrorURI,
} from "@constants/SearchParams";
import LoginLayout from "@layouts/LoginLayout";
import { cn } from "@utils/Styles";

type Outcome = "accepted" | "error" | "rejected";

const outcomes: Record<Outcome, { className: string; icon: ReactNode; title: string }> = {
    accepted: {
        className: "bg-success/10 text-success",
        icon: <CircleCheck />,
        title: "Consent has been accepted and processed",
    },
    error: {
        className: "bg-destructive/10 text-destructive",
        icon: <TriangleAlert />,
        title: "An error occurred processing the request",
    },
    rejected: {
        className: "bg-muted text-muted-foreground",
        icon: <CircleSlash />,
        title: "Consent has been rejected and processed",
    },
};

function CompletionView() {
    const { t: translate } = useTranslation(["consent"]);

    const [query] = useSearchParams();

    const decision = query.get(Decision);
    const error = query.get(ErrorParam);

    const outcome: Outcome = error ? "error" : decision === DecisionAccepted ? "accepted" : "rejected";
    const { className, icon, title } = outcomes[outcome];

    return (
        <LoginLayout id={"openid-completion-stage"} title={translate(title)} maxWidth={"sm"}>
            <Empty
                data-testid={"openid-completion-outcome"}
                data-outcome={outcome}
                className="gap-5 border-0 p-0 md:p-0"
            >
                <EmptyHeader className="max-w-none">
                    <EmptyMedia variant={"icon"} className={cn("mb-0 size-12 [&_svg]:size-7", className)}>
                        {icon}
                    </EmptyMedia>
                    <EmptyDescription>
                        {translate("You may close this tab or return home by clicking the home button")}.
                    </EmptyDescription>
                </EmptyHeader>
                {error ? (
                    <CompletionErrorView
                        error={error}
                        error_description={query.get(ErrorDescription)}
                        error_hint={query.get(ErrorHint)}
                        error_debug={query.get(ErrorDebug)}
                        error_uri={query.get(ErrorURI)}
                    />
                ) : null}
                <EmptyContent>
                    <HomeButton />
                </EmptyContent>
            </Empty>
        </LoginLayout>
    );
}

export interface ErrorProps {
    error: string;
    error_description: null | string;
    error_hint: null | string;
    error_debug: null | string;
    error_uri: null | string;
}

function CompletionErrorView({ error, error_debug, error_description, error_hint, error_uri }: ErrorProps) {
    const { t: translate } = useTranslation(["consent"]);

    const [open, setOpen] = useState(false);

    const details: { label: string; value: null | string }[] = [
        { label: translate("Error"), value: error },
        { label: translate("Description"), value: error_description },
        { label: translate("Hint"), value: error_hint },
        { label: translate("Documentation"), value: error_uri },
    ];

    return (
        <Card className="w-full gap-0 overflow-hidden border-destructive/40 py-0 text-left">
            <ItemGroup>
                {details.map(({ label, value }) =>
                    value ? (
                        <Item key={label} size={"sm"} className="items-start">
                            <ItemContent>
                                <ItemTitle className="text-xs tracking-wide text-muted-foreground uppercase">
                                    {label}
                                </ItemTitle>
                                <ItemDescription className="line-clamp-none font-mono text-xs break-all text-foreground">
                                    {value}
                                </ItemDescription>
                            </ItemContent>
                        </Item>
                    ) : null,
                )}
            </ItemGroup>
            {error_debug ? (
                <div>
                    <Separator />
                    <button
                        type="button"
                        className="flex w-full items-center gap-2 px-4 py-2.5 text-left text-xs font-medium text-muted-foreground hover:text-foreground"
                        aria-expanded={open}
                        aria-controls={"openid-completion-debug"}
                        onClick={() => setOpen((previous) => !previous)}
                    >
                        <ChevronDown className={cn("size-4 transition-transform", open ? "rotate-180" : "")} />
                        {translate("Debug Information")}
                    </button>
                    {open ? (
                        <pre
                            id={"openid-completion-debug"}
                            className="overflow-x-auto px-4 pb-4 font-mono text-xs whitespace-pre-wrap text-muted-foreground"
                        >
                            {error_debug}
                        </pre>
                    ) : null}
                </div>
            ) : null}
        </Card>
    );
}

export default CompletionView;

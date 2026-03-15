import { FC } from "react";

import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";

import HomeButton from "@components/HomeButton";
import { Card } from "@components/UI/Card";
import { Separator } from "@components/UI/Separator";
import {
    Decision,
    ErrorDebug,
    ErrorDescription,
    ErrorHint,
    Error as ErrorParam,
    ErrorURI,
} from "@constants/SearchParams";
import LoginLayout from "@layouts/LoginLayout";
const CompletionView = () => {
    const { t: translate } = useTranslation(["consent"]);

    const [query] = useSearchParams();

    const decision = query.get(Decision);
    const error = query.get(ErrorParam);
    const error_description = query.get(ErrorDescription);
    const error_hint = query.get(ErrorHint);
    const error_debug = query.get(ErrorDebug);
    const error_uri = query.get(ErrorURI);

    let title;
    if (error) {
        title = "An error occurred processing the request";
    } else if (decision === "accepted") {
        title = "Consent has been accepted and processed";
    } else {
        title = "Consent has been rejected and processed";
    }

    return (
        <LoginLayout id={"openid-completion-stage"} title={translate(title)} maxWidth={"sm"}>
            <div className="flex flex-col items-center justify-center gap-4">
                <HomeButton />
                {error ? (
                    <CompletionErrorView
                        error={error}
                        error_description={error_description}
                        error_hint={error_hint}
                        error_debug={error_debug}
                        error_uri={error_uri}
                    />
                ) : null}

                <p className="m-4 text-sm text-muted-foreground">
                    {translate("You may close this tab or return home by clicking the home button")}.
                </p>
            </div>
        </LoginLayout>
    );
};

export default CompletionView;

interface ErrorProps {
    error: string;
    error_description: null | string;
    error_hint: null | string;
    error_debug: null | string;
    error_uri: null | string;
}
const CompletionErrorView: FC<ErrorProps> = (props: ErrorProps) => {
    const { t: translate } = useTranslation(["consent"]);

    return (
        <Card className="p-4 shadow-xl">
            <div className="flex flex-col gap-4">
                <h6 className="text-lg font-semibold">
                    <strong>{translate("Error")}:</strong> {translate(props.error)}
                </h6>
                {props.error_description || props.error_hint || props.error_debug || props.error_uri ? (
                    <Separator />
                ) : null}
                {props.error_description ? (
                    <p>
                        <strong>{translate("Description")}:</strong> {translate(props.error_description)}
                    </p>
                ) : null}
                {props.error_hint ? (
                    <p>
                        <strong>{translate("Hint")}:</strong> {translate(props.error_hint)}
                    </p>
                ) : null}
                {props.error_debug ? (
                    <p>
                        <strong>{translate("Debug Information")}:</strong> {translate(props.error_debug)}
                    </p>
                ) : null}
                {props.error_uri ? (
                    <p>
                        <strong>{translate("Documentation")}:</strong> {translate(props.error_uri)}
                    </p>
                ) : null}
            </div>
        </Card>
    );
};

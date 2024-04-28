import { Fragment, ReactNode } from "react";

import LoadingPage from "@views/LoadingPage/LoadingPage";

interface ComponentOrLoadingProps {
    ready: boolean;

    children: ReactNode;
}

export const ComponentOrLoading = (props: ComponentOrLoadingProps) => {
    return (
        <Fragment>
            <div className={props.ready ? "hidden" : ""}>
                <LoadingPage />
            </div>
            {props.ready ? props.children : null}
        </Fragment>
    );
};

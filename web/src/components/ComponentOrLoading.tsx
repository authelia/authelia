import React, { Fragment, ReactNode } from "react";

import LoadingPage from "@views/LoadingPage/LoadingPage";

export interface Props {
    ready: boolean;

    children: ReactNode;
}

const ComponentOrLoading = function (props: Props) {
    return (
        <Fragment>
            <div className={props.ready ? "hidden" : ""}>
                <LoadingPage />
            </div>
            {props.ready ? props.children : null}
        </Fragment>
    );
};

export default ComponentOrLoading;

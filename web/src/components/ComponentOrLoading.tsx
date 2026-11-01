// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, ReactNode } from "react";

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

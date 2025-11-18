import { Fragment, ReactNode } from "react";

import { Box } from "@mui/material";

import LoadingPage from "@views/LoadingPage/LoadingPage";

export interface Props {
    ready: boolean;

    children: ReactNode;
}

const ComponentOrLoading = function (props: Props) {
    return (
        <Fragment>
            <Box className={props.ready ? "hidden" : ""}>
                <LoadingPage />
            </Box>
            {props.ready ? props.children : null}
        </Fragment>
    );
};

export default ComponentOrLoading;

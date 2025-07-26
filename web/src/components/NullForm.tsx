import React, { ReactNode } from "react";

export interface Props {
    autocomplete?: boolean;

    children: ReactNode;
}

const NulLForm = function (props: Props) {
    return (
        <form
            onSubmit={(e) => {
                e.preventDefault();
                return false;
            }}
            autoComplete={props.autocomplete ? "on" : "off"}
        >
            {props.children}
        </form>
    );
};

export default NulLForm;

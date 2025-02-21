export enum Errors {
    forbidden = "forbidden",
}

export interface ErrorInfo {
    errorCode?: string;
    redirectionUrl?: string;
}

export interface ErrorProps {
    info: ErrorInfo;
    children?: any;
}

export type ErrorComponent = (props: ErrorProps) => JSX.Element;
